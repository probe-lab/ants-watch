package ants

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	ds "github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/ipfs/go-log/v2"
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

type cacheEntry[T any] struct {
	value   T
	addedAt time.Time
}

func (c cacheEntry[T]) IsExpired() bool {
	return time.Since(c.addedAt) > 7*24*time.Hour
}

type QueenConfig struct {
	KeysDBPath         string
	CertsPath          string
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

	peerstore peerstore.Peerstore
	datastore ds.Batching

	ants       []*Ant
	antsEvents chan RequestEvent

	// portsOccupancy is a slice of bools that represent the occupancy of the ports
	// false corresponds to an available port, true to an occupied port
	// the first item of the slice corresponds to the firstPort
	portsOccupancy []bool

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

	queen := &Queen{
		cfg:              cfg,
		id:               uuid.NewString(),
		nebulaDB:         NewNebulaDB(cfg.NebulaDBConnString, cfg.UserAgent, cfg.CrawlInterval),
		keysDB:           NewKeysDB(cfg.KeysDBPath),
		peerstore:        ps,
		datastore:        ldb,
		ants:             []*Ant{},
		antsEvents:       make(chan RequestEvent, 1024),
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

// Run makes the queen orchestrate the ant nest
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

			request, err := q.handleRequestEvent(ctx, evt, requests)
			if err != nil {
				logger.Warn("Error handling request event: ", err)
				continue
			}

			requests = append(requests, request)
			if len(requests) >= q.cfg.BatchSize {
				if err := q.clickhouseClient.BulkInsertRequests(ctx, requests); err != nil {
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

func (q *Queen) handleRequestEvent(ctx context.Context, evt RequestEvent, requests []*db.Request) (*db.Request, error) {
	q.cfg.Telemetry.TrackedRequestsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", evt.Type.String()),
	))

	protocolStrs := protocol.ConvertToStrings(evt.Protocols)
	sort.Strings(protocolStrs)

	maddrStrs := evt.MaddrStrings()
	sort.Strings(maddrStrs)

	uuidv7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("creating uuid: %w", err)
	}

	return &db.Request{
		UUID:           uuidv7,
		QueenID:        q.id,
		AntID:          evt.Self,
		RemoteID:       evt.Remote,
		RequestType:    evt.Type,
		AgentVersion:   evt.AgentVersion,
		Protocols:      protocolStrs,
		StartedAt:      evt.Timestamp,
		KeyID:          evt.Target.B58String(),
		MultiAddresses: maddrStrs,
		ConnMaddr:      evt.ConnMaddr.String(),
	}, nil
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

// routine must be called periodically to ensure that the number and positions
// of ants is still relevant given the latest observed DHT servers.
func (q *Queen) routine(ctx context.Context) {
	// get online DHT servers from the Nebula database
	networkPeers, err := q.nebulaDB.GetLatestPeerIds(ctx)
	if err != nil {
		logger.Warn("unable to get latest peer ids from Nebula ", err)
		return
	}

	// build a binary trie from the network peers
	networkTrie := trie.New[bit256.Key, peer.ID]()
	for _, peerId := range networkPeers {
		networkTrie.Add(PeerIDToKadID(peerId), peerId)
	}

	// zones correspond to the prefixes of the tries that must be covered by an
	// ant. One ant's kademlia ID MUST match each of the returned prefixes in
	// order to ensure global coverage.
	zones := trieZones(networkTrie, q.cfg.BucketSize-1)
	logger.Debugf("%d zones must be covered by ants", len(zones))

	// convert string zone to bitstr.Key
	missingKeys := make([]bitstr.Key, len(zones))
	for i, zoneStr := range zones {
		missingKeys[i] = bitstr.Key(zoneStr)
	}

	var excessAntsIndices []int
	// remove keys covered by existing ants, and mark ants that aren't needed anymore
	for index, ant := range q.ants {
		matchedKey := false
		for i, missingKey := range missingKeys {
			if key.CommonPrefixLength(ant.kadID, missingKey) == missingKey.BitLen() {
				// remove key from missingKeys since covered by exisitng
				missingKeys = append(missingKeys[:i], missingKeys[i+1:]...)
				matchedKey = true
				break
			}
		}
		if !matchedKey {
			// This ant is not needed anymore. Two ants end up in the same zone, the
			// younger one is discarded.
			excessAntsIndices = append(excessAntsIndices, index)
		}
	}
	logger.Debugf("currently have %d ants", len(q.ants))
	logger.Debugf("need %d extra ants", len(missingKeys))
	logger.Debugf("removing %d ants", len(excessAntsIndices))

	// kill ants that are not needed anymore
	// sort indices in descending order to remove from end first
	sort.Sort(sort.Reverse(sort.IntSlice(excessAntsIndices)))
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

	// get libp2p private keys whose kademlia id matches the missing key prefixes
	privKeys := q.keysDB.MatchingKeys(missingKeys, returnedKeys)
	// add missing ants
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
			RequestsChan:   q.antsEvents,
			CertPath:       q.cfg.CertsPath,
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

// trieZones is a recursive function returning the prefixes that the ants must
// have in order to cover the complete keyspace. The prefixes correspond to
// subtries/branches, that have at most zoneSize (=bucketSize-1) peers. They
// must be the largest subtries with at most zoneSize peers. The returned
// prefixes cover the whole keyspace even if they don't all have the same
// length.
//
// e.g ["00", "010", "001", "1"] is a valid return value since the prefixes
// cover all possible values. In this specific example, the trie would be
// unbalanced, and would have only a few peers with the prefix "1", than
// starting with "0".
func trieZones[K kad.Key[K], T any](t *trie.Trie[K, T], zoneSize int) []string {
	if t.Size() < zoneSize {
		// We've hit the bottom of the trie. There are less peers in the (sub)trie
		// than the zone size, hence spawning a single ant is enough to cover this
		// (sub)trie.
		//
		// Since we are't aware of the subtrie location in the greater trie, it is
		// the parent's responsibility to add the prefix.
		return []string{""}
	}

	// a trie is composed of two branches, respectively starting with "0" and
	// "1". Take the returned prefixes from each branch (subtrie), and add the
	// corresponding prefix before returning them to the parent.
	zones := []string{}
	if !t.Branch(0).IsLeaf() {
		for _, zone := range trieZones(t.Branch(0), zoneSize) {
			zones = append(zones, "0"+zone)
		}
	}
	if !t.Branch(1).IsLeaf() {
		for _, zone := range trieZones(t.Branch(1), zoneSize) {
			zones = append(zones, "1"+zone)
		}
	}
	return zones
}
