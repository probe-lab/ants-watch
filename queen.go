package ants

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/antslog"
	kadpb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/probe-lab/go-libdht/kad"
	"github.com/probe-lab/go-libdht/kad/key"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/probe-lab/go-libdht/kad/trie"
)

var logger = log.Logger("ants-queen")

type Queen struct {
	nebulaDB *NebulaDB
	keysDB   *KeysDB

	ants     []*Ant
	antsLogs chan antslog.RequestLog

	seen map[peer.ID]struct{}
}

func NewQueen(dbConnString string, keysDbPath string) *Queen {
	nebulaDB := NewNebulaDB(dbConnString)
	keysDB := NewKeysDB(keysDbPath)

	logger.Debug("queen created")

	return &Queen{
		nebulaDB: nebulaDB,
		keysDB:   keysDB,
		ants:     []*Ant{},
		antsLogs: make(chan antslog.RequestLog, 1024),
		seen:     make(map[peer.ID]struct{}),
	}
}

func (q *Queen) Run(ctx context.Context) {
	go q.consumeAntsLogs(ctx)
	t := time.NewTicker(CRAWL_INTERVAL)
	q.routine(ctx)

	for {
		select {
		case <-t.C:
			// crawl
			q.routine(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (q *Queen) consumeAntsLogs(ctx context.Context) {
	lnCount := 0
	f, err := os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// TODO: improve this
		panic(err)
	}
	defer f.Close()

	for {
		select {
		case log := <-q.antsLogs:
			if false {
				reqType := kadpb.Message_MessageType(log.Type).String()
				// TODO: persistant logging
				fmt.Printf("%s \tself: %s \ttype: %s \trequester: %s \ttarget: %s \tagent: %s \tmaddrs: %s\n", log.Timestamp.Format(time.RFC3339), log.Self, reqType, log.Requester, log.Target.B58String(), log.Agent, log.Maddrs)
			} else {
				if _, ok := q.seen[log.Requester]; !ok {
					q.seen[log.Requester] = struct{}{}
					if strings.Contains(log.Agent, "light") {
						lnCount++
					}
					// count := len(q.seen)
					// if count > 1 {
					// 	fmt.Printf("\033[F")
					// } else {
					// 	fmt.Printf("%s \tstarting sniffing\n", time.Now().Format(time.RFC3339))
					// }
					fmt.Fprintf(f, "\r%s    %s\n", log.Requester, log.Agent)
					// fmt.Printf("%s\ttotal: %d \tlight: %d\n", time.Now().Format(time.RFC3339), len(q.seen), lnCount)
					logger.Debugf("total: %d \tlight: %d", len(q.seen), lnCount)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (q *Queen) routine(ctx context.Context) {
	networkPeers, err := q.nebulaDB.GetLatestPeerIds(ctx)
	if err != nil {
		logger.Warn("unable to get latest peer ids from Nebula", err)
		return
	}

	networkTrie := trie.New[bit256.Key, peer.ID]()
	for _, peerId := range networkPeers {
		networkTrie.Add(PeeridToKadid(peerId), peerId)
	}

	// zones correspond to the prefixes of the tries that must be covered by an ant
	zones := trieZones(networkTrie, BUCKET_SIZE)
	logger.Debugf("%d zones must be covered by ants", len(zones))

	// convert string zone to bitstr.Key
	missingKeys := make([]bitstr.Key, len(zones))
	for i, zoneStr := range zones {
		missingKeys[i] = bitstr.Key(zoneStr)
	}

	var excessAntsIndices []int
	// remove keys covered by existing ants, and mark useless ants
	for index, ant := range q.ants {
		matchedKey := false
		for i, missingKey := range missingKeys {
			if key.CommonPrefixLength(ant.KadId, missingKey) == missingKey.BitLen() {
				// remove key from missingKeys since covered by current ant
				missingKeys = append(missingKeys[:i], missingKeys[i+1:]...)
				matchedKey = true
				break
			}
		}
		if !matchedKey {
			// this ant is not needed anymore
			// two ants end up in the same zone, the younger one is discarded
			excessAntsIndices = append(excessAntsIndices, index)
		}
	}
	logger.Debugf("currently have %d ants", len(q.ants))
	logger.Debugf("need %d extra ants", len(missingKeys))
	logger.Debugf("removing %d ants", len(excessAntsIndices))

	// remove ants
	for _, index := range excessAntsIndices {
		q.ants[index].Close()
		q.ants = append(q.ants[:index], q.ants[index+1:]...)
	}

	// add missing ants
	privKeys := q.keysDB.MatchingKeys(missingKeys)
	for _, key := range privKeys {
		ant, err := SpawnAnt(ctx, key, q.antsLogs)
		if err != nil {
			logger.Warn("error creating ant", err)
		}
		q.ants = append(q.ants, ant)
	}

	logger.Debugf("ants count: %d", len(q.ants))
	logger.Debug("queen routine over")
}

func trieZones[K kad.Key[K], T any](t *trie.Trie[K, T], zoneSize int) []string {
	if t.Size() < zoneSize {
		return []string{""}
	}

	zones := []string{}
	if !t.Branch(0).IsLeaf() {
		for _, zone := range trieZones(t.Branch(0), zoneSize) {
			zones = append(zones, "0"+zone)
		}
	}
	for _, zone := range trieZones(t.Branch(1), zoneSize) {
		zones = append(zones, "1"+zone)
	}
	return zones
}
