package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	lru "github.com/hashicorp/golang-lru"
	glog "github.com/ipfs/go-log/v2"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	// mh "github.com/multiformats/go-multihash"
	"github.com/dennis-tra/nebula-crawler/config"
	ndb "github.com/dennis-tra/nebula-crawler/db"
	nutils "github.com/dennis-tra/nebula-crawler/utils"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/probe-lab/ants-watch/db/models"
	mt "github.com/probe-lab/ants-watch/metrics"
)

//go:embed migrations
var migrations embed.FS

var logger = glog.Logger("client-db")

var (
	ErrEmptyAgentVersion = fmt.Errorf("empty agent version")
	ErrEmptyProtocol     = fmt.Errorf("empty protocol")
	ErrEmptyProtocolsSet = fmt.Errorf("empty protocols set")
)

// type DatabaseConfig struct {
// 	Host         string
// 	Port         int
// 	Name         string // Database name
// 	User         string
// 	Password     string
// 	SSLMode      string //one of "require" or "disable"
// 	MaxIdleConns int

// 	// The cache size to hold agent versions in memory to skip database queries.
// 	AgentVersionsCacheSize int

// 	// The cache size to hold protocols in memory to skip database queries.
// 	ProtocolsCacheSize int

// 	// The cache size to hold sets of protocols in memory to skip database queries.
// 	ProtocolsSetCacheSize int

// 	MeterProvider  metric.MeterProvider
// 	TracerProvider trace.TracerProvider
// }

// func (c *DatabaseConfig) DatabaseString() string {
// 	return fmt.Sprintf(
// 		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
// 		c.Host,
// 		c.Port,
// 		c.Name,
// 		c.User,
// 		c.Password,
// 		c.SSLMode,
// 	)
// }

// type DBClient struct {
// 	ctx context.Context
// 	cfg DatabaseConfig

// 	// Database handler
// 	DBH *sql.DB

// 	// protocols cache
// 	agentVersions *lru.Cache

// 	// protocols cache
// 	protocols *lru.Cache

// 	// protocols set cache
// 	protocolsSets *lru.Cache

// 	// Database telemetry
// 	telemetry *mt.Telemetry
// }

type AntsDBClient struct {
	*ndb.DBClient
	Telemetry *mt.Telemetry
}

func InitDBClient(ctx context.Context, cfg *config.Database) (*AntsDBClient, error) {
	log.WithFields(log.Fields{
		"host": cfg.DatabaseHost,
		"port": cfg.DatabasePort,
		"name": cfg.DatabaseName,
		"user": cfg.DatabaseUser,
		"ssl":  cfg.DatabaseSSLMode,
	}).Infoln("Initializing database client")

	dbh, err := otelsql.Open("postgres", cfg.DatabaseSourceName(),
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithMeterProvider(cfg.MeterProvider),
		otelsql.WithTracerProvider(cfg.TracerProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Set to match the writer worker
	dbh.SetMaxIdleConns(cfg.MaxIdleConns) // default is 2 which leads to many connection open/closings

	otelsql.ReportDBStatsMetrics(dbh, otelsql.WithMeterProvider(cfg.MeterProvider))

	// Ping database to verify connection.
	if err = dbh.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	telemetry, err := mt.NewTelemetry(cfg.TracerProvider, cfg.MeterProvider)
	if err != nil {
		return nil, fmt.Errorf("new telemetry: %w", err)
	}

	client := &AntsDBClient{
		DBClient: &ndb.DBClient{
			Ctx: ctx,
			Cfg: cfg,
			Dbh: dbh,
		},
		Telemetry: telemetry,
	}
	client.ApplyMigrations(cfg, dbh)

	client.AgentVersions, err = lru.New(cfg.AgentVersionsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new agent versions lru cache: %w", err)
	}

	client.Protocols, err = lru.New(cfg.ProtocolsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new protocol lru cache: %w", err)
	}

	client.ProtocolsSets, err = lru.New(cfg.ProtocolsSetCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new protocols set lru cache: %w", err)
	}

	err = client.FillCaches(ctx)
	if err != nil {
		return nil, err
	}

	client.ensurePartitions(ctx, time.Now())
	client.ensurePartitions(ctx, time.Now().Add(24*time.Hour))

	go func() {
		for range time.NewTicker(24 * time.Hour).C {
			client.ensurePartitions(ctx, time.Now().Add(12*time.Hour))
		}
	}()

	return client, nil
}

func (c *AntsDBClient) ensurePartitions(ctx context.Context, baseDate time.Time) {
	lowerBound := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, baseDate.Location())
	upperBound := lowerBound.AddDate(0, 1, 0)

	query := ndb.PartitionQuery(models.TableNames.PeerLogs, lowerBound, upperBound)
	if _, err := c.Dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create peer_logs partition")
	}

	query = ndb.PartitionQuery(models.TableNames.Requests, lowerBound, upperBound)
	if _, err := c.Dbh.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create requests partition")
	}
}

type InsertRequestResult struct {
	PID peer.ID
}

func (c *AntsDBClient) insertRequest(
	ctx context.Context,
	timestamp time.Time,
	requestType string,
	antID peer.ID,
	peerID peer.ID,
	keyID string,
	maddrs []ma.Multiaddr,
	protocolsSetID null.Int,
	agentVersionsID null.Int,
) (*InsertRequestResult, error) {
	maddrStrs := nutils.MaddrsToAddrs(maddrs)
	start := time.Now()

	// we want to upsert peers before anything else
	// since it could be an ant or a key
	peer, err := c.UpsertPeerForAnts(
		peerID.String(),
		agentVersionsID,
		protocolsSetID,
		timestamp,
	)
	if err != nil {
		log.WithError(err).Printf("Could not upsert: < %s >\n", peerID.String())
	}

	// insert ant to peers table if cannot be found
	antModel, err := models.Peers(qm.Where("multi_hash=?", antID.String())).One(ctx, c.Dbh)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var ant int
	if antModel == nil {
		ant, _ = c.UpsertPeerForAnts(
			antID.String(),
			agentVersionsID,
			protocolsSetID,
			timestamp,
		)
	} else {
		ant = antModel.ID
	}

	// taking care of keys last since it could be a peer upserted
	// from looking for an ant or a peer above
	var key int
	keyModel, _ := models.Keys(qm.Where("multi_hash=?", keyID)).One(ctx, c.Dbh)
	if keyModel == nil {
		var row *sql.Rows
		var sqlErr error

		// Decode the keyID to check if it's a peer ID ("identity" multihash)
		// if decoded, err := mh.Decode([]byte(keyID)); err == nil && decoded.Name == "identity" {
		peerModel, _ := models.Peers(qm.Where("multi_hash=?", keyID)).One(ctx, c.Dbh)
		if peerModel != nil {
			// If it's an identity peer, insert it with NULL multi_hash
			row, sqlErr = queries.Raw("SELECT insert_key($1, NULL)", peerModel.ID).QueryContext(ctx, c.Dbh)
		} else {
			// If not an identity peer, insert it with NULL peer_id and the multihash
			row, sqlErr = queries.Raw("SELECT insert_key(NULL, $1)", keyID).QueryContext(ctx, c.Dbh)
		}
		if sqlErr != nil {
			fmt.Printf("SQL ERR: %v", sqlErr)
			return nil, sqlErr
		}
		row.Next()
		err := row.Scan(&key)
		if err != nil {
			log.WithError(err).Printf("Could not scan key: < %d >\n", &key)
		}

		defer func() {
			if err := row.Close(); err != nil {
				log.WithError(err).Warnln("Could not close rows")
			}
		}()
	} else {
		key = keyModel.ID
	}

	rows, err := queries.Raw("SELECT insert_request($1, $2, $3, $4, $5, $6, $7, $8)",
		timestamp,
		requestType,
		ant,
		peer,
		key,
		types.StringArray(maddrStrs),
		protocolsSetID,
		agentVersionsID,
	).QueryContext(ctx, c.Dbh)
	if err != nil {
		fmt.Printf("\n\nERROR $1", err)
		return nil, err
	}

	c.Telemetry.InsertRequestHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(
		attribute.String("type", requestType),
		attribute.Bool("success", err == nil),
	))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close rows")
		}
	}()

	ivr := InsertRequestResult{
		PID: peerID,
	}
	if !rows.Next() {
		return &ivr, nil
	}

	if err = rows.Scan(&ivr); err != nil {
		return nil, err
	}

	return &ivr, nil
}

func (c *AntsDBClient) PersistRequest(
	ctx context.Context,
	timestamp time.Time,
	requestType string,
	antID peer.ID,
	peerID peer.ID,
	keyID string,
	maddrs []ma.Multiaddr,
	agentVersion string,
	protocols []string,
) (*InsertRequestResult, error) {
	var agentVersionID, protocolsSetID *int
	var avidErr, psidErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		agentVersionID, avidErr = c.GetOrCreateAgentVersionID(ctx, c.Dbh, agentVersion)
		if avidErr != nil && !errors.Is(avidErr, ErrEmptyAgentVersion) && !errors.Is(psidErr, context.Canceled) {
			log.WithError(avidErr).WithField("agentVersion", agentVersion).Warnln("Error getting or creating agent version id")
		}
		wg.Done()
	}()
	go func() {
		protocolsSetID, psidErr = c.GetOrCreateProtocolsSetID(ctx, c.Dbh, protocols)
		if psidErr != nil && !errors.Is(psidErr, ErrEmptyProtocolsSet) && !errors.Is(psidErr, context.Canceled) {
			log.WithError(psidErr).WithField("protocols", protocols).Warnln("Error getting or creating protocols set id")
		}
		wg.Done()
	}()
	wg.Wait()

	return c.insertRequest(
		ctx,
		timestamp,
		requestType,
		antID,
		peerID,
		keyID,
		maddrs,
		null.IntFromPtr(protocolsSetID),
		null.IntFromPtr(agentVersionID),
	)
}

func (c *AntsDBClient) UpsertPeerForAnts(mh string, agentVersionID null.Int, protocolSetID null.Int, timestamp time.Time) (int, error) {
	rows, err := queries.Raw("SELECT upsert_peer($1, $2, $3, $4)",
		mh, agentVersionID, protocolSetID, timestamp,
	).Query(c.Dbh)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close rows")
		}
	}()

	id := 0
	if !rows.Next() {
		return id, nil
	}

	if err = rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (c *AntsDBClient) FetchUnresolvedMultiAddrsForAnts(ctx context.Context, limit int) (models.MultiAddressSlice, error) {
	return models.MultiAddresses(
		models.MultiAddressWhere.Resolved.EQ(false),
		qm.OrderBy(models.MultiAddressColumns.CreatedAt),
		qm.Limit(limit),
	).All(ctx, c.Dbh)
}
