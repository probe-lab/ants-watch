package ants

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/antslog"
	kadpb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
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

	peerstore peerstore.Peerstore
	datastore ds.Batching

	ants     []*Ant
	antsLogs chan antslog.RequestLog

	seen map[peer.ID]struct{}

	upnp bool
	// portsOccupancy is a slice of bools that represent the occupancy of the ports
	// false corresponds to an available port, true to an occupied port
	// the first item of the slice corresponds to the firstPort
	portsOccupancy []bool
	firstPort      uint16
}

func NewQueen(dbConnString string, keysDbPath string, nPorts, firstPort uint16) (*Queen, error) {
	nebulaDB := NewNebulaDB(dbConnString)
	keysDB := NewKeysDB(keysDbPath)
	peerstore, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, err
	}

	queen := &Queen{
		nebulaDB:  nebulaDB,
		keysDB:    keysDB,
		peerstore: peerstore,
		datastore: dssync.MutexWrap(ds.NewMapDatastore()),
		ants:      []*Ant{},
		antsLogs:  make(chan antslog.RequestLog, 1024),
		seen:      make(map[peer.ID]struct{}),
	}

	if nPorts == 0 {
		queen.upnp = true
	} else {
		queen.upnp = false
		queen.firstPort = firstPort
		queen.portsOccupancy = make([]bool, nPorts)
	}

	logger.Debug("queen created")

	return queen, nil
}

func (q *Queen) takeAvailablePort() (uint16, error) {
	if q.upnp {
		return 0, nil
	}
	for i, occupied := range q.portsOccupancy {
		if !occupied {
			q.portsOccupancy[i] = true
			return q.firstPort + uint16(i), nil
		}
	}
	return 0, fmt.Errorf("no available port")
}

func (q *Queen) freePort(port uint16) {
	if !q.upnp {
		q.portsOccupancy[port-q.firstPort] = false
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
			maddrs := q.peerstore.Addrs(log.Requester)
			var agent string
			peerstoreAgent, err := q.peerstore.Get(log.Requester, "AgentVersion")
			if err != nil {
				agent = "unknown"
			} else {
				agent = peerstoreAgent.(string)
			}
			if false {
				reqType := kadpb.Message_MessageType(log.Type).String()
				// TODO: persistant logging
				fmt.Printf("%s \tself: %s \ttype: %s \trequester: %s \ttarget: %s \tagent: %s \tmaddrs: %s\n", log.Timestamp.Format(time.RFC3339), log.Self, reqType, log.Requester, log.Target.B58String(), agent, maddrs)
			} else {
				if _, ok := q.seen[log.Requester]; !ok {
					q.seen[log.Requester] = struct{}{}
					if strings.Contains(agent, "light") {
						lnCount++
					}
					// count := len(q.seen)
					// if count > 1 {
					// 	fmt.Printf("\033[F")
					// } else {
					// 	fmt.Printf("%s \tstarting sniffing\n", time.Now().Format(time.RFC3339))
					// }
					fmt.Fprintf(f, "\r%s    %s\n", log.Requester, agent)
					// fmt.Printf("%s\ttotal: %d \tlight: %d\n", time.Now().Format(time.RFC3339), len(q.seen), lnCount)
					logger.Debugf("total: %d \tlight: %d \tqueue len: %d", len(q.seen), lnCount, len(q.antsLogs))
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
		port := q.ants[index].port
		q.ants[index].Close()
		q.ants = append(q.ants[:index], q.ants[index+1:]...)
		q.freePort(port)
	}

	// add missing ants
	privKeys := q.keysDB.MatchingKeys(missingKeys)
	for _, key := range privKeys {
		port, err := q.takeAvailablePort()
		if err != nil {
			logger.Error("trying to spwan new ant: ")
			continue
		}
		ant, err := SpawnAnt(ctx, key, q.peerstore, q.datastore, port, q.antsLogs)
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
