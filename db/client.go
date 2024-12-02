package db

import (
	"context"
	"crypto/tls"
	"golang.org/x/net/proxy"
	"net"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	mt "github.com/probe-lab/ants-watch/metrics"
)

type Client struct {
	ctx  context.Context
	conn driver.Conn

	telemetry *mt.Telemetry
}

func NewDatabaseClient(ctx context.Context, address, database, username, password string, ssl bool) (*Client, error) {
	logger.Infoln("Creating new database client...")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{address},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Debug: true,
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d proxy.ContextDialer
			if ssl {
				d = &tls.Dialer{}
			} else {
				d = &net.Dialer{}
			}

			return d.DialContext(ctx, "tcp", addr)
		},
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
