package db

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	// "github.com/dennis-tra/nebula-crawler/config"
	lru "github.com/hashicorp/golang-lru"
	mt "github.com/probe-lab/ants-watch/metrics"
	// log "github.com/ipfs/go-log/v2"
)

type Client struct {
	ctx  context.Context
	conn driver.Conn

	agentVersion  *lru.Cache
	protocols     *lru.Cache
	protocolsSets *lru.Cache

	telemetry *mt.Telemetry
}

func NewDatabaseClient(ctx context.Context, address, database, username, password string) (*Client, error) {
	logger.Infoln("Creating new database client...")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{address},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Debug: true,
	})

	if err != nil {
		return nil, err
	}

	return &Client{
		ctx:  ctx,
		conn: conn,
	}, nil
}
