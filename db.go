package ants

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/libp2p/go-libp2p/core/peer"
)

type NebulaDB struct {
	ConnString string

	connPool *pgxpool.Pool
}

func NewDB(connString string) *NebulaDB {
	return &NebulaDB{
		ConnString: connString,
	}
}

func (db *NebulaDB) Open(ctx context.Context) error {
	connPool, err := pgxpool.New(ctx, db.ConnString)
	if err == nil {
		db.connPool = connPool
	}
	return err
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

	crawlIdQuery := `
	    SELECT c.id
        FROM crawls c
        WHERE c.started_at > $1
        ORDER BY c.started_at ASC
        LIMIT 1
	`

	crawlIntervalAgo := time.Now().Add(-CRAWL_INTERVAL)
	var crawlId uint64
	err := db.connPool.QueryRow(ctx, crawlIdQuery, crawlIntervalAgo).Scan(&crawlId)
	if err != nil {
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

	beforeLastCrawlStarted := crawlIntervalAgo.Add(-CRAWL_INTERVAL)
	rows, err := db.connPool.Query(ctx, peersQuery, beforeLastCrawlStarted, crawlId)
	if err != nil {
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

	return peerIds, nil
}
