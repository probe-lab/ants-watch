// mostly copy-pastad from https://github.com/dennis-tra/nebula/blob/main/db/client_db.go
// except the `insertRequest` function

package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	lru "github.com/hashicorp/golang-lru"
	glog "github.com/ipfs/go-log/v2"
	_ "github.com/lib/pq"
	ma "github.com/multiformats/go-multiaddr"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

type DBClient struct {
	ctx context.Context
	cfg config.Database

	// Database handler
	Handler *sql.DB

	// protocols cache
	agentVersions *lru.Cache

	// protocols cache
	protocols *lru.Cache

	// protocols set cache
	protocolsSets *lru.Cache

	// Database telemetry
	telemetry *mt.Telemetry
}

func InitDBClient(ctx context.Context, cfg *config.Database) (*DBClient, error) {
	log.WithFields(log.Fields{
		"host": cfg.DatabaseHost,
		"port": cfg.DatabasePort,
		"name": cfg.DatabaseName,
		"user": cfg.DatabaseUser,
		"ssl":  cfg.DatabaseSSLMode,
	}).Infoln("Initializing database client")

	connString := cfg.DatabaseSourceName()
	dbh, err := otelsql.Open("postgres", connString,
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

	client := &DBClient{
		ctx:       ctx,
		cfg:       *cfg,
		telemetry: telemetry,
		Handler:   dbh,
	}
	client.applyMigrations(cfg, dbh)

	client.agentVersions, err = lru.New(cfg.AgentVersionsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new agent versions lru cache: %w", err)
	}

	client.protocols, err = lru.New(cfg.ProtocolsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new protocol lru cache: %w", err)
	}

	client.protocolsSets, err = lru.New(cfg.ProtocolsSetCacheSize)
	if err != nil {
		return nil, fmt.Errorf("new protocols set lru cache: %w", err)
	}

	if err = client.fillAgentVersionsCache(ctx); err != nil {
		return nil, fmt.Errorf("fill agent versions cache: %w", err)
	}

	if err = client.fillProtocolsCache(ctx); err != nil {
		return nil, fmt.Errorf("fill protocols cache: %w", err)
	}

	if err = client.fillProtocolsSetCache(ctx); err != nil {
		return nil, fmt.Errorf("fill protocols set cache: %w", err)
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

func (c *DBClient) ensurePartitions(ctx context.Context, baseDate time.Time) {
	lowerBound := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, baseDate.Location())
	upperBound := lowerBound.AddDate(0, 1, 0)

	query := partitionQuery(models.TableNames.PeerLogs, lowerBound, upperBound)
	if _, err := c.Handler.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create peer_logs partition")
	}

	query = partitionQuery(models.TableNames.Requests, lowerBound, upperBound)
	if _, err := c.Handler.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create requests partition")
	}

	query = partitionQuery(models.TableNames.RequestsDenormalized, lowerBound, upperBound)
	if _, err := c.Handler.ExecContext(ctx, query); err != nil {
		log.WithError(err).WithField("query", query).Warnln("could not create requests partition")
	}

}

func partitionQuery(table string, lower time.Time, upper time.Time) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%s_%s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')",
		table,
		lower.Format("2006"),
		lower.Format("01"),
		table,
		lower.Format("2006-01-02"),
		upper.Format("2006-01-02"),
	)
}

func (c *DBClient) applyMigrations(cfg *config.Database, Handler *sql.DB) {
	tmpDir, err := os.MkdirTemp("", "nebula")
	if err != nil {
		log.WithError(err).WithField("pattern", "ants-watch").Warnln("Could not create tmp directory for migrations")
		return
	}
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			log.WithError(err).WithField("tmpDir", tmpDir).Warnln("Could not clean up tmp directory")
		}
	}()
	log.WithField("dir", tmpDir).Debugln("Created temporary directory")

	err = fs.WalkDir(migrations, ".", func(path string, d fs.DirEntry, err error) error {
		join := filepath.Join(tmpDir, path)
		if d.IsDir() {
			return os.MkdirAll(join, 0o755)
		}

		data, err := migrations.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		return os.WriteFile(join, data, 0o644)
	})
	if err != nil {
		log.WithError(err).Warnln("Could not create migrations files")
		return
	}

	// Apply migrations
	driver, err := postgres.WithInstance(Handler, &postgres.Config{})
	if err != nil {
		log.WithError(err).Warnln("Could not create driver instance")
		return
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(tmpDir, "migrations"), cfg.DatabaseName, driver)
	if err != nil {
		log.WithError(err).Warnln("Could not create migrate instance")
		return
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.WithError(err).Warnln("Couldn't apply migrations")
		return
	}
}

type InsertRequestResult struct {
	PID string
}

func (c *DBClient) insertRequest(
	ctx context.Context,
	timestamp time.Time,
	requestType string,
	antID string,
	peerID string,
	keyID string,
	maddrs []string,
	protocolsSetID null.Int,
	agentVersionsID null.Int,
) (*InsertRequestResult, error) {
	start := time.Now()

	rows, err := queries.Raw("SELECT insert_request($1, $2, $3, $4, $5, $6, $7, $8)",
		timestamp,
		requestType,
		antID,
		peerID,
		keyID,
		types.StringArray(maddrs),
		protocolsSetID,
		agentVersionsID,
	).QueryContext(ctx, c.Handler)
	if err != nil {
		return nil, err
	}

	c.telemetry.InsertRequestHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(
		attribute.String("type", requestType),
		attribute.Bool("success", err == nil),
	))

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

func MaddrsToAddrs(maddrs []ma.Multiaddr) []string {
	addrs := make([]string, len(maddrs))
	for i, maddr := range maddrs {
		addrs[i] = maddr.String()
	}
	return addrs
}

// protocolsSetHash returns a unique hash digest for this set of protocol IDs as it's also generated by the database.
// It expects the list of protocolIDs to be sorted in ascending order.
func (c *DBClient) protocolsSetHash(protocolIDs []int64) string {
	protocolStrs := make([]string, len(protocolIDs))
	for i, id := range protocolIDs {
		protocolStrs[i] = strconv.Itoa(int(id)) // safe because protocol IDs are just integers in the database.
	}
	dat := []byte("{" + strings.Join(protocolStrs, ",") + "}")

	h := sha256.New()
	h.Write(dat)
	return string(h.Sum(nil))
}

func (c *DBClient) GetOrCreateProtocolsSetID(ctx context.Context, exec boil.ContextExecutor, protocols []string) (*int, error) {
	if len(protocols) == 0 {
		return nil, db.ErrEmptyProtocolsSet
	}

	protocolIDs := make([]int64, len(protocols))
	for i, protocol := range protocols {
		protocolID, err := c.GetOrCreateProtocol(ctx, exec, protocol)
		if errors.Is(err, db.ErrEmptyProtocol) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("get or create protocol: %w", err)
		}
		protocolIDs[i] = int64(*protocolID)
	}

	sort.Slice(protocolIDs, func(i, j int) bool { return protocolIDs[i] < protocolIDs[j] })

	key := c.protocolsSetHash(protocolIDs)
	if id, found := c.protocolsSets.Get(key); found {
		c.telemetry.CacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("entity", "protocol_set"),
			attribute.Bool("hit", true),
		))
		return id.(*int), nil
	}
	c.telemetry.CacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", "protocol_set"),
		attribute.Bool("hit", false),
	))

	log.WithField("key", hex.EncodeToString([]byte(key))).Infoln("Upsert protocols set")
	row := exec.QueryRowContext(ctx, "SELECT upsert_protocol_set_id($1)", types.Int64Array(protocolIDs))
	if row.Err() != nil {
		return nil, fmt.Errorf("unable to upsert protocols set: %w", row.Err())
	}

	var protocolsSetID *int
	if err := row.Scan(&protocolsSetID); err != nil {
		return nil, fmt.Errorf("unable to scan result from upsert protocol set id: %w", err)
	}

	if protocolsSetID == nil {
		return nil, fmt.Errorf("protocols set not created")
	}

	c.protocolsSets.Add(key, protocolsSetID)

	return protocolsSetID, nil
}

func (c *DBClient) GetOrCreateProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string) (*int, error) {
	if protocol == "" {
		return nil, db.ErrEmptyProtocol
	}

	if id, found := c.protocols.Get(protocol); found {
		c.telemetry.CacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("entity", "protocol"),
			attribute.Bool("hit", true),
		))
		return id.(*int), nil
	}
	c.telemetry.CacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", "protocol"),
		attribute.Bool("hit", false),
	))

	log.WithField("protocol", protocol).Infoln("Upsert protocol")
	row := exec.QueryRowContext(ctx, "SELECT upsert_protocol($1)", protocol)
	if row.Err() != nil {
		return nil, fmt.Errorf("unable to upsert protocol: %w", row.Err())
	}

	var protocolID *int
	if err := row.Scan(&protocolID); err != nil {
		return nil, fmt.Errorf("unable to scan result from upsert protocol: %w", err)
	}

	if protocolID == nil {
		return nil, fmt.Errorf("protocol not created")
	}

	c.protocols.Add(protocol, protocolID)

	return protocolID, nil
}

func (c *DBClient) PersistRequest(
	ctx context.Context,
	timestamp time.Time,
	requestType string,
	antID string,
	peerID string,
	keyID string,
	maddrs []string,
	agentVersion null.String,
	protocols []string,
) (*InsertRequestResult, error) {
	var agentVersionID, protocolsSetID *int
	var avidErr, psidErr error

	agentVersionString := agentVersion.String

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		agentVersionID, avidErr = c.GetOrCreateAgentVersionID(ctx, c.Handler, agentVersionString)
		if avidErr != nil && !errors.Is(avidErr, db.ErrEmptyAgentVersion) && !errors.Is(psidErr, context.Canceled) {
			log.WithError(avidErr).WithField("agentVersion", agentVersion).Warnln("Error getting or creating agent version id")
		}
		wg.Done()
	}()
	go func() {
		protocolsSetID, psidErr = c.GetOrCreateProtocolsSetID(ctx, c.Handler, protocols)
		if psidErr != nil && !errors.Is(psidErr, db.ErrEmptyProtocolsSet) && !errors.Is(psidErr, context.Canceled) {
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

func (c *DBClient) GetOrCreateAgentVersionID(ctx context.Context, exec boil.ContextExecutor, agentVersion string) (*int, error) {
	if agentVersion == "" {
		return nil, db.ErrEmptyAgentVersion
	}

	if id, found := c.agentVersions.Get(agentVersion); found {
		c.telemetry.CacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("entity", "agent_version"),
			attribute.Bool("hit", true),
		))
		return id.(*int), nil
	}
	c.telemetry.CacheQueriesCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", "agent_version"),
		attribute.Bool("hit", false),
	))

	log.WithField("agentVersion", agentVersion).Infoln("Upsert agent version")
	row := exec.QueryRowContext(ctx, "SELECT upsert_agent_version($1)", agentVersion)
	if row.Err() != nil {
		return nil, fmt.Errorf("unable to upsert agent version: %w", row.Err())
	}

	var agentVersionID *int
	if err := row.Scan(&agentVersionID); err != nil {
		return nil, fmt.Errorf("unable to scan result from upsert agent version: %w", err)
	}

	if agentVersionID == nil {
		return nil, fmt.Errorf("agentVersion not created")
	}

	c.agentVersions.Add(agentVersion, agentVersionID)

	return agentVersionID, nil
}

// fillAgentVersionsCache fetches all rows until agent version cache size from the agent_versions table and
// initializes the DB clients agent version cache.
func (c *DBClient) fillAgentVersionsCache(ctx context.Context) error {
	if c.cfg.AgentVersionsCacheSize == 0 {
		return nil
	}

	avs, err := models.AgentVersions(qm.Limit(c.cfg.AgentVersionsCacheSize)).All(ctx, c.Handler)
	if err != nil {
		return err
	}

	for _, av := range avs {
		c.agentVersions.Add(av.AgentVersion, &av.ID)
	}

	return nil
}

// fillProtocolsSetCache fetches all rows until protocolSet cache size from the protocolsSets table and
// initializes the DB clients protocolsSets cache.
func (c *DBClient) fillProtocolsSetCache(ctx context.Context) error {
	if c.cfg.ProtocolsSetCacheSize == 0 {
		return nil
	}

	protSets, err := models.ProtocolsSets(qm.Limit(c.cfg.ProtocolsSetCacheSize)).All(ctx, c.Handler)
	if err != nil {
		return err
	}

	for _, ps := range protSets {
		c.protocolsSets.Add(string(ps.Hash), &ps.ID)
	}

	return nil
}

// fillProtocolsCache fetches all rows until protocol cache size from the protocols table and
// initializes the DB clients protocols cache.
func (c *DBClient) fillProtocolsCache(ctx context.Context) error {
	if c.cfg.ProtocolsCacheSize == 0 {
		return nil
	}

	prots, err := models.Protocols(qm.Limit(c.cfg.ProtocolsCacheSize)).All(ctx, c.Handler)
	if err != nil {
		return err
	}

	for _, p := range prots {
		c.protocols.Add(p.Protocol, &p.ID)
	}

	return nil
}

func (c *DBClient) UpsertPeer(mh string, agentVersionID null.Int, protocolSetID null.Int, timestamp time.Time) (int, error) {
	rows, err := queries.Raw("SELECT upsert_peer($1, $2, $3, $4)",
		mh, agentVersionID, protocolSetID, timestamp,
	).Query(c.Handler)
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

// FetchUnresolvedMultiAddresses fetches all multi addresses that were not resolved yet.
func (c *DBClient) FetchUnresolvedMultiAddresses(ctx context.Context, limit int) (models.MultiAddressSlice, error) {
	return models.MultiAddresses(
		models.MultiAddressWhere.Resolved.EQ(false),
		qm.OrderBy(models.MultiAddressColumns.CreatedAt),
		qm.Limit(limit),
	).All(ctx, c.Handler)
}

func BulkInsertRequests(db *sql.DB, requests []models.RequestsDenormalized) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}
	i := 1

	for _, request := range requests {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6))
		valueArgs = append(valueArgs, request.Timestamp, request.RequestType, request.AntID, request.PeerID, request.KeyID, request.MultiAddressIds, request.AgentVersion)
		i += 7
	}

	stmt := fmt.Sprintf("INSERT INTO requests_denormalized (timestamp, request_type, ant_id, peer_id, key_id, multi_address_ids, agent_version) VALUES %s RETURNING id;",
		strings.Join(valueStrings, ", "))

	rows, err := db.Query(stmt, valueArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func NormalizeRequests(ctx context.Context, db *sql.DB, dbClient *DBClient) error {
	// Fetch rows from `requests_denormalized` where `normalized_at` is NULL
	rows, err := db.Query("SELECT id, timestamp, request_type, ant_id, peer_id, key_id, multi_address_ids, agent_version FROM requests_denormalized WHERE normalized_at IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var request models.RequestsDenormalized
		if err := rows.Scan(&request.ID, &request.Timestamp, &request.RequestType, &request.AntID, &request.PeerID, &request.KeyID, &request.MultiAddressIds, &request.AgentVersion); err != nil {
			return err
		}

		_, err = dbClient.PersistRequest(
			ctx,
			request.Timestamp,
			request.RequestType,
			request.AntID,
			request.PeerID,
			request.KeyID,
			request.MultiAddressIds,
			request.AgentVersion, // agent versions
			nil,                  // protocol sets
		)
		if err != nil {
			return fmt.Errorf("failed to normalize request ID %d: %w", request.ID, err)
		}

		_, err = db.Exec("UPDATE requests_denormalized SET normalized_at = NOW() WHERE id = $1", request.ID)
		if err != nil {
			return fmt.Errorf("failed to update normalized_at for request ID %d: %w", request.ID, err)
		}
	}

	return nil
}

func AddrsToMaddrs(addrs []string) ([]ma.Multiaddr, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maddrs[i] = maddr
	}

	return maddrs, nil
}
