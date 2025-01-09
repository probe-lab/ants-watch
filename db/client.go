package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/probe-lab/ants-watch/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

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

type NoopClient struct{}

var _ Client = (*NoopClient)(nil)

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (n NoopClient) Ping(ctx context.Context) error {
	return nil
}

func (n NoopClient) BulkInsertRequests(ctx context.Context, requests []*Request) error {
	return nil
}

type ClickhouseClient struct {
	driver.Conn
	telemetry *metrics.Telemetry
}

var _ Client = (*ClickhouseClient)(nil)

func NewClickhouseClient(address, database, username, password string, ssl bool, telemetry *metrics.Telemetry) (*ClickhouseClient, error) {
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
		Conn:      conn,
		telemetry: telemetry,
	}

	return client, nil
}

func (c *ClickhouseClient) BulkInsertRequests(ctx context.Context, requests []*Request) (err error) {
	start := time.Now()
	defer func() {
		c.telemetry.BulkInsertCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("success", strconv.FormatBool(err == nil))))
		c.telemetry.BulkInsertSizeHist.Record(ctx, int64(len(requests)))
		c.telemetry.BulkInsertLatencyMsHist.Record(ctx, time.Since(start).Milliseconds())
	}()
	batch, err := c.Conn.PrepareBatch(ctx, "INSERT INTO requests", driver.WithReleaseConnection())
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, r := range requests {
		err = batch.AppendStruct(r)
		if err != nil {
			return fmt.Errorf("append request to batch: %w", err)
		}
	}
	return batch.Send()
}
