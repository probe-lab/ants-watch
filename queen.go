package ants

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/antslog"
	kadpb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/probe-lab/go-libdht/kad"
	"github.com/probe-lab/go-libdht/kad/key"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/probe-lab/go-libdht/kad/trie"

	"github.com/dennis-tra/nebula-crawler/maxmind"
	"github.com/dennis-tra/nebula-crawler/tele"
	"github.com/dennis-tra/nebula-crawler/udger"
	"github.com/probe-lab/ants-watch/db"
	"github.com/probe-lab/ants-watch/db/models"
	// "github.com/probe-lab/ants-watch/db/models"
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
	resolveBatchTime int // in sec
}

func NewQueen(ctx context.Context, dbConnString string, keysDbPath string, nPorts, firstPort uint16) *Queen {
	nebulaDB := NewNebulaDB(dbConnString)
	keysDB := NewKeysDB(keysDbPath)

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fmt.Errorf("Port must be an integer", err)
	}

	mP, _ := tele.NewMeterProvider()
	tP, _ := tele.NewTracerProvider(ctx, "", 0)

	dbc, err := db.InitDBClient(ctx, &config.Database{
		DatabaseHost:           os.Getenv("DB_HOST"),
		DatabasePort:           dbPort,
		DatabaseName:           os.Getenv("DB_DATABASE"),
		DatabaseUser:           os.Getenv("DB_USER"),
		DatabasePassword:       os.Getenv("DB_PASSWORD"),
		MeterProvider:          mP,
		TracerProvider:         tP,
		ProtocolsCacheSize:     100,
		ProtocolsSetCacheSize:  200,
		AgentVersionsCacheSize: 200,
		DatabaseSSLMode:        os.Getenv("DB_SSLMODE"),
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

	batchSizeEnvVal := os.Getenv("BATCH_SIZE")
	if len(batchSizeEnvVal) == 0 {
		batchSizeEnvVal = "1000"
	}
	batchSize, err := strconv.Atoi(batchSizeEnvVal)
	if err != nil {
		logger.Errorln("BATCH_SIZE should be an integer")
	}

	batchTimeEnvVal := os.Getenv("BATCH_SIZE")
	if len(batchTimeEnvVal) == 0 {
		batchTimeEnvVal = "30"
	}
	batchTime, err := strconv.Atoi(batchTimeEnvVal)
	if err != nil {
		logger.Errorln("BATCH_TIME should be an integer")
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
		resolveBatchSize: batchSize,
		resolveBatchTime: batchTime,
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

	requests := make([]models.RequestsDenormalized, 0, q.resolveBatchSize)
	// bulk insert for every batch size or N seconds, whichever comes first
	ticker := time.NewTicker(time.Duration(q.resolveBatchTime) * time.Second)
	defer ticker.Stop()

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
			// protocols := make([]string, 0)

			request := models.RequestsDenormalized{
				Timestamp:       log.Timestamp,
				RequestType:     reqType,
				AntID:           log.Self.String(),
				PeerID:          log.Requester.String(),
				KeyID:           log.Target.B58String(),
				MultiAddressIds: db.MaddrsToAddrs(log.Maddrs),
			}
			requests = append(requests, request)
			if len(requests) >= q.resolveBatchSize {
				err = db.BulkInsertRequests(q.dbc.Handler, requests)
				if err != nil {
					logger.Fatalf("Error inserting requests: %v", err)
				}
				requests = requests[:0]
			}
			if _, ok := q.seen[log.Requester]; !ok {
				q.seen[log.Requester] = struct{}{}
				if strings.Contains(log.Agent, "light") {
					lnCount++
				}
				fmt.Fprintf(f, "\r%s    %s\n", log.Requester, log.Agent)
				logger.Debugf("total: %d \tlight: %d", len(q.seen), lnCount)
			}

		case <-ticker.C:
			if len(requests) > 0 {
				err = db.BulkInsertRequests(q.dbc.Handler, requests)
				if err != nil {
					logger.Fatalf("Error inserting requests: %v", err)
				}
				requests = requests[:0]
			}

		case <-ctx.Done():
			if len(requests) > 0 {
				err = db.BulkInsertRequests(q.dbc.Handler, requests)
				if err != nil {
					logger.Fatalf("Error inserting remaining requests: %v", err)
				}
			}
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
