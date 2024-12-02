package ants

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/libp2p/go-libp2p/core/peer"
)

type NebulaProvider interface {
	GetLatestPeerIds(ctx context.Context) ([]peer.ID, error)
}

type NebulaDB struct {
	ConnString    string
	CrawlInterval time.Duration

	connPool *pgxpool.Pool
}

var _ NebulaProvider = (*NebulaDB)(nil)

func NewNebulaDB(connString string, crawlInterval time.Duration) *NebulaDB {
	return &NebulaDB{
		ConnString:    connString,
		CrawlInterval: crawlInterval,
	}
}

func (db *NebulaDB) Open(ctx context.Context) error {
	if db.connPool != nil {
		return nil
	}
	connPool, err := pgxpool.New(ctx, db.ConnString)
	if err != nil {
		logger.Warn("unable to open connection to Nebula DB: ", err)
		return err
	}
	logger.Debug("opened connection to Nebula DB")
	db.connPool = connPool
	return nil
}

func (db *NebulaDB) Close() {
	db.connPool.Close()
	db.connPool = nil
}

func (db *NebulaDB) GetLatestPeerIds(ctx context.Context) ([]peer.ID, error) {
	// Open a connection if it's not already open
	if db.connPool == nil {
		err := db.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}

	logger.Debug("getting last crawl from Nebula DB")

	crawlIdQuery := `
	    SELECT c.id
        FROM crawls c
        WHERE c.started_at > $1
        ORDER BY c.started_at ASC
        LIMIT 1
	`

	crawlIntervalAgo := time.Now().Add(-db.CrawlInterval)
	var crawlId uint64
	err := db.connPool.QueryRow(ctx, crawlIdQuery, crawlIntervalAgo).Scan(&crawlId)
	if err != nil {
		logger.Warn("unable to get last crawl from Nebula DB: ", err)
		return nil, err
	}

	peersQuery := `
		SELECT p.multi_hash
		FROM visits v
		JOIN peers p ON p.id = v.peer_id
		WHERE v.visit_started_at >= $1
			AND v.crawl_id = $2
			AND v.connect_error IS NULL
	`

	beforeLastCrawlStarted := crawlIntervalAgo.Add(-db.CrawlInterval)
	rows, err := db.connPool.Query(ctx, peersQuery, beforeLastCrawlStarted, crawlId)
	if err != nil {
		logger.Warn("unable to get peers from Nebula DB: ", err)
		return nil, err
	}

	var peerIds []peer.ID
	for rows.Next() {
		var multiHash string
		err := rows.Scan(&multiHash)
		if err != nil {
			continue
		}
		peerId, err := peer.Decode(multiHash)
		if err != nil {
			continue
		}
		peerIds = append(peerIds, peerId)
	}

	logger.Debugf("found %d peers during the last Nebula crawl", len(peerIds))

	return peerIds, nil
}

type NebulaServiceProvider struct {
	address string
}

var _ NebulaProvider = (*NebulaServiceProvider)(nil)

func NewNebulaServiceProvider() *NebulaServiceProvider {
	// TODO: init gRPC service client
	return &NebulaServiceProvider{}
}

func (n NebulaServiceProvider) GetLatestPeerIds(ctx context.Context) ([]peer.ID, error) {
	// TODO: request peers from Nebula service
	return nil, nil
}
