package db

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4/database/multistmt"
	"github.com/google/uuid"
	mt "github.com/probe-lab/ants-watch/metrics"
)

type Client struct {
	ctx  context.Context
	conn driver.Conn

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

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return &Client{
		ctx:  ctx,
		conn: conn,
	}, nil
}

type BatchRequest struct {
	ctx context.Context

	insertStatement string

	conn  driver.Conn
	batch driver.Batch
}

func NewBatch(ctx context.Context, conn driver.Conn, insertStatement string) (*BatchRequest, error) {
	batch, err := conn.PrepareBatch(ctx, insertStatement, driver.WithReleaseConnection())
	if err != nil {
		return nil, err
	}

	return &BatchRequest{
		ctx:             ctx,
		insertStatement: insertStatement,
		conn:            conn,
		batch:           batch,
	}, nil
}

func (b *BatchRequest) Append(id, antMultihash, remoteMultihash, agentVersion string, protocols []string, startedAt time.Time, requestType, keyMultihash string, multiAddresses []string) error {
	return b.batch.Append(
		id,
		antMultihash,
		remoteMultihash,
		agentVersion,
		protocols,
		startedAt,
		requestType,
		keyMultihash,
		multiAddresses,
	)
}
