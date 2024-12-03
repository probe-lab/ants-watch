package ants

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru/v2"
	ds "github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/ants"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/probe-lab/ants-watch/db"
	"github.com/probe-lab/ants-watch/metrics"
	"github.com/probe-lab/go-libdht/kad"
	"github.com/probe-lab/go-libdht/kad/key"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/probe-lab/go-libdht/kad/trie"
)

var logger = log.Logger("ants-queen")

type QueenConfig struct {
	KeysDBPath         string
	NPorts             int
	FirstPort          int
	UPnP               bool
	BatchSize          int
	BatchTime          time.Duration
	CrawlInterval      time.Duration
	CacheSize          int
	NebulaDBConnString string
	BucketSize         int
	UserAgent          string
	Telemetry          *metrics.Telemetry
}

type Queen struct {
	cfg *QueenConfig

	id       string
	nebulaDB *NebulaDB
	keysDB   *KeysDB

	peerstore      peerstore.Peerstore
	datastore      ds.Batching
	agentsCache    *lru.Cache[string, string]
	protocolsCache *lru.Cache[string, []protocol.ID]

	ants       []*Ant
	antsEvents chan ants.RequestEvent

	// portsOccupancy is a slice of bools that represent the occupancy of the ports
	// false corresponds to an available port, true to an occupied port
	// the first item of the slice corresponds to the firstPort
	portsOccupancy   []bool
	clickhouseClient db.Client
}

func NewQueen(clickhouseClient db.Client, cfg *QueenConfig) (*Queen, error) {
	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, fmt.Errorf("creating peerstore: %w", err)
	}

	ldb, err := leveldb.NewDatastore("", nil) // empty string means in-memory
	if err != nil {
		return nil, fmt.Errorf("creating in-memory leveldb: %w", err)
	}

	agentsCache, err := lru.New[string, string](cfg.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("init agents cache: %w", err)
	}

	protocolsCache, err := lru.New[string, []protocol.ID](cfg.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("init agents cache: %w", err)
	}

	queen := &Queen{
		cfg:              cfg,
		id:               uuid.NewString(),
		nebulaDB:         NewNebulaDB(cfg.NebulaDBConnString, cfg.CrawlInterval),
		keysDB:           NewKeysDB(cfg.KeysDBPath),
		peerstore:        ps,
		datastore:        ldb,
		ants:             []*Ant{},
		antsEvents:       make(chan ants.RequestEvent, 1024),
		agentsCache:      agentsCache,
		protocolsCache:   protocolsCache,
		clickhouseClient: clickhouseClient,
		portsOccupancy:   make([]bool, cfg.NPorts),
	}

	return queen, nil
}

func (q *Queen) takeAvailablePort() (int, error) {
	if q.cfg.UPnP {
		return 0, nil
	}

	for i, occupied := range q.portsOccupancy {
		if occupied {
			continue
		}
		q.portsOccupancy[i] = true
		return q.cfg.FirstPort + i, nil
	}

	return 0, fmt.Errorf("no available port")
}

func (q *Queen) freePort(port int) {
	if !q.cfg.UPnP {
		q.portsOccupancy[port-q.cfg.FirstPort] = false
	}
}

func (q *Queen) Run(ctx context.Context) error {
	logger.Infoln("Queen.Run started")
	defer logger.Infoln("Queen.Run completing")

	if err := q.nebulaDB.Open(ctx); err != nil {
		return fmt.Errorf("opening nebula db: %w", err)
	}

	go q.consumeAntsEvents(ctx)

	crawlTime := time.NewTicker(q.cfg.CrawlInterval)
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

func (q *Queen) consumeAntsEvents(ctx context.Context) {
	requests := make([]*db.Request, 0, q.cfg.BatchSize)

	// bulk insert for every batch size or N seconds, whichever comes first
	ticker := time.NewTicker(q.cfg.BatchTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debugln("Gracefully shutting down ants...")
			logger.Debugln("Number of requests remaining to be inserted:", len(requests))

			if len(requests) > 0 {
				if err := q.clickhouseClient.BulkInsertRequests(ctx, requests); err != nil {
					logger.Errorf("Error inserting requests: %v", err)
				}
				requests = requests[:0]
			}
			return

		case evt := <-q.antsEvents:

			// transform multi addresses
			maddrStrs := make([]string, len(evt.Maddrs))
			for i, maddr := range evt.Maddrs {
				maddrStrs[i] = maddr.String()
			}

			// cache agent version
			if evt.AgentVersion == "" {
				var found bool
				evt.AgentVersion, found = q.agentsCache.Get(evt.Remote.String())
				q.cfg.Telemetry.CacheHitCounter.Add(ctx, 1, metric.WithAttributes(
					attribute.String("hit", strconv.FormatBool(found)),
					attribute.String("cache", "agent_version"),
				))
				if found {
					continue
				}
			} else {
				q.agentsCache.Add(evt.Remote.String(), evt.AgentVersion)
			}

			// cache protocols
			var protocols []protocol.ID
			if len(evt.Protocols) == 0 {
				var found bool
				protocols, found = q.protocolsCache.Get(evt.Remote.String())
				q.cfg.Telemetry.CacheHitCounter.Add(ctx, 1, metric.WithAttributes(
					attribute.String("hit", strconv.FormatBool(found)),
					attribute.String("cache", "protocols"),
				))
			} else {
				protocols = evt.Protocols
				q.protocolsCache.Add(evt.Remote.String(), evt.Protocols)
			}
			protocolStrs := protocol.ConvertToStrings(protocols)

			uuidv7, err := uuid.NewV7()
			if err != nil {
				logger.Warn("Error generating uuid")
				continue
			}

			request := &db.Request{
				UUID:           uuidv7,
				QueenID:        q.id,
				AntID:          evt.Self,
				RemoteID:       evt.Remote,
				Type:           evt.Type,
				AgentVersion:   evt.AgentVersion,
				Protocols:      protocolStrs,
				StartedAt:      evt.Timestamp,
				KeyID:          evt.Target.B58String(),
				MultiAddresses: maddrStrs,
			}

			requests = append(requests, request)

			if len(requests) >= q.cfg.BatchSize {

				if err = q.clickhouseClient.BulkInsertRequests(ctx, requests); err != nil {
					logger.Errorf("Error inserting requests: %v", err)
				}
				requests = requests[:0]
			}

		case <-ticker.C:
			if len(requests) == 0 {
				continue
			}

			if err := q.clickhouseClient.BulkInsertRequests(ctx, requests); err != nil {
				logger.Errorf("Error inserting requests: %v", err)
			}
			requests = requests[:0]
		}
	}
}

func (q *Queen) persistLiveAntsKeys() {
	logger.Debugln("Persisting live ants keys")
	antsKeys := make([]crypto.PrivKey, 0, len(q.ants))
	for _, ant := range q.ants {
		antsKeys = append(antsKeys, ant.cfg.PrivateKey)
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
		networkTrie.Add(PeerIDToKadID(peerId), peerId)
	}

	// zones correspond to the prefixes of the tries that must be covered by an ant
	zones := trieZones(networkTrie, q.cfg.BucketSize)
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
			if key.CommonPrefixLength(ant.kadID, missingKey) == missingKey.BitLen() {
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
		returnedKeys[i] = ant.cfg.PrivateKey
		port := ant.cfg.Port

		if err := ant.Close(); err != nil {
			logger.Warn("error closing ant", err)
		}

		q.ants = append(q.ants[:index], q.ants[index+1:]...)
		q.freePort(port)
	}

	// add missing ants
	privKeys := q.keysDB.MatchingKeys(missingKeys, returnedKeys)
	for _, key := range privKeys {
		port, err := q.takeAvailablePort()
		if err != nil {
			logger.Error("trying to spawn new ant: ", err)
			continue
		}

		antCfg := &AntConfig{
			PrivateKey:     key,
			UserAgent:      q.cfg.UserAgent,
			Port:           port,
			ProtocolPrefix: fmt.Sprintf("/celestia/%s", celestiaNet), // TODO: parameterize
			BootstrapPeers: BootstrapPeers(celestiaNet),              // TODO: parameterize
			EventsChan:     q.antsEvents,
		}

		ant, err := SpawnAnt(ctx, q.peerstore, q.datastore, antCfg)
		if err != nil {
			logger.Warn("error creating ant", err)
			continue
		}

		q.ants = append(q.ants, ant)
	}

	q.cfg.Telemetry.AntsCountGauge.Record(ctx, int64(len(q.ants)))

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
