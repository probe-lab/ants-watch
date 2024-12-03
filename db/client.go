package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/net/proxy"
)

var logger = logging.Logger("db")

type Client interface {
	Ping(ctx context.Context) error
	BulkInsertRequests(ctx context.Context, requests []*Request) error
}

type ClickhouseClient struct {
	driver.Conn
}

var _ Client = (*ClickhouseClient)(nil)

func NewClient(address, database, username, password string, ssl bool) (*ClickhouseClient, error) {
	logger.Infoln("Creating new clickhouse client...")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{address},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
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

	client := &ClickhouseClient{
		Conn: conn,
	}

	return client, nil
}

func (c *ClickhouseClient) BulkInsertRequests(ctx context.Context, requests []*Request) error {
	batch, err := c.Conn.PrepareBatch(ctx, "INSERT INTO requests", driver.WithReleaseConnection())
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, r := range requests {
		err = batch.Append(
			r.UUID.String(),
			r.QueenID,
			r.AntID.String(),
			r.RemoteID.String(),
			r.AgentVersion,
			r.Protocols,
			r.StartedAt,
			r.Type,
			r.KeyID,
			r.MultiAddresses,
			r.IsSelfLookup,
		)
		if err != nil {
			return fmt.Errorf("append request to batch: %w", err)
		}
	}
	return batch.Send()
}
