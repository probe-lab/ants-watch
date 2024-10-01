package ants

import (
	"context"
	"fmt"
	"os"
	"strconv"
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

	"github.com/probe-lab/ants-watch/db"
	"github.com/probe-lab/ants-watch/maxmind"
	"github.com/probe-lab/ants-watch/metrics"
	"github.com/probe-lab/ants-watch/udger"
)

var logger = log.Logger("ants-queen")

type Queen struct {
	nebulaDB *NebulaDB
	keysDB   *KeysDB

	ants     []*Ant
	antsLogs chan antslog.RequestLog

	seen map[peer.ID]struct{}

	upnp bool
	// portsOccupancy is a slice of bools that represent the occupancy of the ports
	// false corresponds to an available port, true to an occupied port
	// the first item of the slice corresponds to the firstPort
	portsOccupancy []bool
	firstPort      uint16

	dbc     *db.DBClient
	mmc     *maxmind.Client
	uclient *udger.Client

	resolveBatchSize int
}

func NewQueen(ctx context.Context, dbConnString string, keysDbPath string, nPorts, firstPort uint16) *Queen {
	nebulaDB := NewNebulaDB(dbConnString)
	keysDB := NewKeysDB(keysDbPath)

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fmt.Errorf("Port must be an integer", err)
	}

	mP, _ := metrics.NewMeterProvider()
	tP, _ := metrics.NewTracerProvider(ctx, "", 0)

	dbc, err := db.InitDBClient(ctx, &db.DatabaseConfig{
		Host:                   os.Getenv("DB_HOST"),
		Port:                   dbPort,
		Name:                   os.Getenv("DB_DATABASE"),
		User:                   os.Getenv("DB_USER"),
		Password:               os.Getenv("DB_PASSWORD"),
		MeterProvider:          mP,
		TracerProvider:         tP,
		ProtocolsCacheSize:     100,
		ProtocolsSetCacheSize:  200,
		AgentVersionsCacheSize: 200,
		SSLMode:                "disable",
	})
	if err != nil {
		logger.Errorf("Failed to initialize DB client: %v\n", err)
	}

	mmc, err := maxmind.NewClient(os.Getenv("MAXMIND_ASN_DB"), os.Getenv("MAXMIND_COUNTRY_DB"))
	if err != nil {
		logger.Errorf("Failed to initialized Maxmind client: %v\n", err)
	}

	filePathUdger := os.Getenv("UDGER_FILEPATH")
	var uclient *udger.Client
	if filePathUdger != "" {
		uclient, err = udger.NewClient(filePathUdger)
		if err != nil {
			logger.Errorf("Failed to initialize Udger client with %s: %v\n", filePathUdger, err)
		}
	}

	queen := &Queen{
		nebulaDB:         nebulaDB,
		keysDB:           keysDB,
		ants:             []*Ant{},
		antsLogs:         make(chan antslog.RequestLog, 1024),
		seen:             make(map[peer.ID]struct{}),
		dbc:              dbc,
		mmc:              mmc,
		uclient:          uclient,
		resolveBatchSize: 1000,
	}

	if nPorts == 0 {
		queen.upnp = true
	} else {
		queen.upnp = false
		queen.firstPort = firstPort
		queen.portsOccupancy = make([]bool, nPorts)
	}

	logger.Debug("queen created")

	return queen
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
	cctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	go q.consumeAntsLogs(cctx)
	t := time.NewTicker(CRAWL_INTERVAL)
	q.routine(ctx)

	for {
		select {
		case <-t.C:
			// crawl
			q.routine(cctx)
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
			reqType := kadpb.Message_MessageType(log.Type).String()
			fmt.Printf(
				"%s \tself: %s \ttype: %s \trequester: %s \ttarget: %s \tagent: %s \tmaddrs: %s\n",
				log.Timestamp.Format(time.RFC3339),
				log.Self,
				reqType,
				log.Requester,
				log.Target.B58String(),
				log.Agent,
				log.Maddrs,
			)

			// Keep this protocols slice empty for now,
			// because we don't need it yet and I don't know how to get it
			protocols := make([]string, 0)
			q.dbc.PersistRequest(
				ctx,
				log.Timestamp,
				reqType,
				log.Self,               // ant ID
				log.Requester,          // peer ID
				log.Target.B58String(), // key ID
				log.Maddrs,
				log.Agent,
				protocols)
			if _, ok := q.seen[log.Requester]; !ok {
				q.seen[log.Requester] = struct{}{}
				if strings.Contains(log.Agent, "light") {
					lnCount++
				}
				fmt.Fprintf(f, "\r%s    %s\n", log.Requester, log.Agent)
				logger.Debugf("total: %d \tlight: %d", len(q.seen), lnCount)
			}
			logger.Info("Fetching multi addresses...")
			dbmaddrs, err := q.dbc.FetchUnresolvedMultiAddresses(ctx, q.resolveBatchSize)
			if err != nil {
				logger.Errorf("fetching multi addresses: %v\n", err)
			}
			logger.Infof("Fetched %d multi addresses", len(dbmaddrs))
			if len(dbmaddrs) == 0 {
				logger.Errorf("Couldn't find multi addresses: %v\n", err)
			}

			if err = db.Resolve(ctx, q.dbc.DBH, q.mmc, q.uclient, dbmaddrs); err != nil {
				logger.Warnf("Error resolving multi addresses: %v\n", err)
			}

		case <-ctx.Done():
			logger.Info("Winding down..")
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
		ant, err := SpawnAnt(ctx, key, port, q.antsLogs)
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
