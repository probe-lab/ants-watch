package ants

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/antslog"
	kadpb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/probe-lab/go-libdht/kad"
	"github.com/probe-lab/go-libdht/kad/key"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/probe-lab/go-libdht/kad/trie"
	"github.com/volatiletech/null/v8"

	"github.com/patrickmn/go-cache"
	"github.com/probe-lab/ants-watch/db"
	"github.com/probe-lab/ants-watch/db/models"
)

var logger = log.Logger("ants-queen")

type Queen struct {
	nebulaDB *NebulaDB
	keysDB   *KeysDB

	peerstore   peerstore.Peerstore
	datastore   ds.Batching
	agentsCache *cache.Cache

	ants     []*Ant
	antsLogs chan antslog.RequestLog

	upnp bool
	// portsOccupancy is a slice of bools that represent the occupancy of the ports
	// false corresponds to an available port, true to an occupied port
	// the first item of the slice corresponds to the firstPort
	portsOccupancy []bool
	firstPort      uint16

	clickhouseClient *db.Client

	resolveBatchSize int
	resolveBatchTime int // in sec
}

func NewQueen(ctx context.Context, dbConnString string, keysDbPath string, nPorts, firstPort uint16, clickhouseClient *db.Client) (*Queen, error) {
	nebulaDB := NewNebulaDB(dbConnString)
	keysDB := NewKeysDB(keysDbPath)
	peerstore, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, err
	}

	queen := &Queen{
		nebulaDB:         nebulaDB,
		keysDB:           keysDB,
		peerstore:        peerstore,
		datastore:        dssync.MutexWrap(ds.NewMapDatastore()),
		ants:             []*Ant{},
		antsLogs:         make(chan antslog.RequestLog, 1024),
		agentsCache:      cache.New(4*24*time.Hour, time.Hour), // 4 days of cache, clean every hour
		upnp:             true,
		resolveBatchSize: getBatchSize(),
		resolveBatchTime: getBatchTime(),
		clickhouseClient: clickhouseClient,
	}

	if nPorts != 0 {
		queen.upnp = false
		queen.firstPort = firstPort
		queen.portsOccupancy = make([]bool, nPorts)
	}

	logger.Info("queen created")

	return queen, nil
}

func getBatchSize() int {
	batchSizeEnvVal := os.Getenv("BATCH_SIZE")
	if len(batchSizeEnvVal) == 0 {
		batchSizeEnvVal = "1000"
	}
	batchSize, err := strconv.Atoi(batchSizeEnvVal)
	if err != nil {
		logger.Errorln("BATCH_SIZE should be an integer")
	}
	return batchSize
}

func getBatchTime() int {
	batchTimeEnvVal := os.Getenv("BATCH_TIME")
	if len(batchTimeEnvVal) == 0 {
		batchTimeEnvVal = "30"
	}
	batchTime, err := strconv.Atoi(batchTimeEnvVal)
	if err != nil {
		logger.Errorln("BATCH_TIME should be an integer")
	}
	return batchTime
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

func (q *Queen) Run(ctx context.Context) error {
	logger.Debugln("Queen.Run started")
	defer logger.Debugln("Queen.Run completing")

	go q.consumeAntsLogs(ctx)

	crawlTime := time.NewTicker(CRAWL_INTERVAL)
	defer crawlTime.Stop()

	q.routine(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Debugln("Queen.Run done..")
			q.persistLiveAntsKeys()
			return ctx.Err()
		case <-crawlTime.C:
			q.routine(ctx)
		}
	}
}

func (q *Queen) consumeAntsLogs(ctx context.Context) {
	requests := make([]models.RequestsDenormalized, 0, q.resolveBatchSize)
	// bulk insert for every batch size or N seconds, whichever comes first
	ticker := time.NewTicker(time.Duration(q.resolveBatchTime) * time.Second)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			logger.Debugln("Gracefully shutting down ants...")
			logger.Debugln("Number of requests remaining to be inserted:", len(requests))
			// if len(requests) > 0 {
			// 	err := db.BulkInsertRequests(context.Background(), q.dbc.Handler, requests)
			// 	if err != nil {
			// 		logger.Fatalf("Error inserting remaining requests: %v", err)
			// 	}
			// }
			return

		case log := <-q.antsLogs:
			reqType := kadpb.Message_MessageType(log.Type).String()
			maddrs := q.peerstore.Addrs(log.Requester)
			var agent string
			peerstoreAgent, err := q.peerstore.Get(log.Requester, "AgentVersion")
			if err != nil {
				if peerstoreAgent, ok := q.agentsCache.Get(log.Requester.String()); ok {
					agent = peerstoreAgent.(string)
				} else {
					agent = ""
				}
			} else {
				agent = peerstoreAgent.(string)
				q.agentsCache.Set(log.Requester.String(), agent, 0)
			}

			protocols, _ := q.peerstore.GetProtocols(log.Requester)
			protocolsAsStr := protocol.ConvertToStrings(protocols)

			request := models.RequestsDenormalized{
				RequestStartedAt: log.Timestamp,
				RequestType:      reqType,
				AntMultihash:     log.Self.String(),
				PeerMultihash:    log.Requester.String(),
				KeyMultihash:     log.Target.B58String(),
				MultiAddresses:   db.MaddrsToAddrs(maddrs),
				AgentVersion:     null.StringFrom(agent),
				Protocols:        protocolsAsStr,
			}
			requests = append(requests, request)
			if len(requests) >= q.resolveBatchSize {
				// err = db.BulkInsertRequests(ctx, q.dbc.Handler, requests)
				// if err != nil {
				// 	logger.Errorf("Error inserting requests: %v", err)
				// }
				// requests = requests[:0]
			}

		case <-ticker.C:
			if len(requests) > 0 {
				// err := db.BulkInsertRequests(ctx, q.dbc.Handler, requests)
				// if err != nil {
				// 	logger.Fatalf("Error inserting requests: %v", err)
				// }
				// requests = requests[:0]
			}

		default:
			// against busy-looping since <-q.antsLogs is a busy chan
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (q *Queen) persistLiveAntsKeys() {
	logger.Debugln("Persisting live ants keys")
	antsKeys := make([]crypto.PrivKey, 0, len(q.ants))
	for _, ant := range q.ants {
		antsKeys = append(antsKeys, ant.Host.Peerstore().PrivKey(ant.Host.ID()))
	}
	q.keysDB.MatchingKeys(nil, antsKeys)
	logger.Debugf("Number of antsKeys persisted: %d", len(antsKeys))
}

func (q *Queen) routine(ctx context.Context) {
	networkPeers, err := q.nebulaDB.GetLatestPeerIds(ctx)
	if err != nil {
		logger.Warn("unable to get latest peer ids from Nebula ", err)
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
	returnedKeys := make([]crypto.PrivKey, len(excessAntsIndices))
	for i, index := range excessAntsIndices {
		ant := q.ants[index]
		returnedKeys[i] = ant.privKey
		port := ant.port
		ant.Close()
		q.ants = append(q.ants[:index], q.ants[index+1:]...)
		q.freePort(port)
	}

	// add missing ants
	privKeys := q.keysDB.MatchingKeys(missingKeys, returnedKeys)
	for _, key := range privKeys {
		port, err := q.takeAvailablePort()
		if err != nil {
			logger.Error("trying to spawn new ant: ")
			continue
		}
		ant, err := SpawnAnt(ctx, key, q.peerstore, q.datastore, port, q.antsLogs)
		if err != nil {
			logger.Warn("error creating ant", err)
		}
		q.ants = append(q.ants, ant)
	}

	for _, ant := range q.ants {
		logger.Debugf("Upserting ant: %v\n", ant.Host.ID().String())
		// antID, err := q.dbc.UpsertPeer(ctx, ant.Host.ID().String(), null.StringFrom(ant.UserAgent), nil, time.Now())
		if err != nil {
			logger.Errorf("Couldn't upsert")
			// logger.Errorf("antID: %d could not be inserted because of %v", antID, err)
		}
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
