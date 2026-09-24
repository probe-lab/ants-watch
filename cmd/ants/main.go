package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/probe-lab/ants-watch"
	"github.com/probe-lab/ants-watch/db"
	"github.com/probe-lab/ants-watch/metrics"
	nebulav1 "github.com/probe-lab/ants-watch/proto/nebula/v1"
	gccli "github.com/probe-lab/go-commons/cli"
	gcdb "github.com/probe-lab/go-commons/db"
)

var queenConfig = struct {
	NebulaSvcHost   string
	NebulaSvcPort   int
	KeyDBPath       string
	CertsPath       string
	NumPorts        int
	FirstPort       int
	UPnp            bool
	BatchSize       int
	BatchTime       time.Duration
	CrawlInterval   time.Duration
	CacheSize       int
	BucketSize      int
	UserAgent       string
	Network         string
	ThrottleTimeout time.Duration
}{
	NebulaSvcHost:   "localhost",
	NebulaSvcPort:   8383,
	KeyDBPath:       "keys.db",
	CertsPath:       "p2p-forge-certs",
	NumPorts:        128,
	FirstPort:       6000,
	UPnp:            false,
	BatchSize:       1000,
	BatchTime:       20 * time.Second,
	CrawlInterval:   120 * time.Minute,
	CacheSize:       10_000,
	BucketSize:      20,
	UserAgent:       ants.UserAgent(ants.CelestiaMainnet),
	Network:         string(ants.CelestiaMainnet),
	ThrottleTimeout: 5 * time.Minute,
}

func main() {
	_ = logging.SetLogLevel("dht", "error")
	_ = logging.SetLogLevel("basichost", "info")

	// ClickHouse connection defaults to an empty host: an unset host selects the
	// no-op writer, matching the previous "no address means don't persist" mode.
	chCfg := &gcdb.ClickHouseConfig{
		BaseConfig: &gcdb.ClickHouseBaseConfig{Port: 9000, SSL: true},
	}
	migrationsCfg := gcdb.DefaultClickHouseMigrationsConfig()

	cmd := &cli.Command{
		Name:  "ants",
		Usage: "Get DHT clients in your p2p network using a honeypot",
		Commands: []*cli.Command{
			queenCommand(chCfg, migrationsCfg),
			healthCommand(),
		},
	}

	root, _ := gccli.NewRootCommand(cmd)
	if err := root.Run(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("running app", "err", err)
		os.Exit(1)
	}
}

func queenCommand(chCfg *gcdb.ClickHouseConfig, migrationsCfg *gcdb.ClickHouseMigrationsConfig) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "network",
			Usage:       "Which network to use",
			Sources:     cli.EnvVars("ANTS_NETWORK"),
			Destination: &queenConfig.Network,
			Value:       queenConfig.Network,
		},
		&cli.StringFlag{
			Name:        "nebula.svc.host",
			Usage:       "The host where to reach the nebula service",
			Sources:     cli.EnvVars("ANTS_NEBULA_SERVICE_HOST"),
			Destination: &queenConfig.NebulaSvcHost,
			Value:       queenConfig.NebulaSvcHost,
		},
		&cli.IntFlag{
			Name:        "nebula.svc.port",
			Usage:       "The port where to reach the nebula service",
			Sources:     cli.EnvVars("ANTS_NEBULA_SERVICE_PORT"),
			Destination: &queenConfig.NebulaSvcPort,
			Value:       queenConfig.NebulaSvcPort,
		},
		&cli.IntFlag{
			Name:        "batch.size",
			Usage:       "The number of ants requests to buffer before flushing to ClickHouse",
			Sources:     cli.EnvVars("ANTS_BATCH_SIZE"),
			Destination: &queenConfig.BatchSize,
			Value:       queenConfig.BatchSize,
		},
		&cli.DurationFlag{
			Name:        "batch.time",
			Usage:       "The maximum time to wait between flushes",
			Sources:     cli.EnvVars("ANTS_BATCH_TIME"),
			Destination: &queenConfig.BatchTime,
			Value:       queenConfig.BatchTime,
		},
		&cli.DurationFlag{
			Name:        "crawl.interval",
			Usage:       "The time between two crawls",
			Sources:     cli.EnvVars("ANTS_CRAWL_INTERVAL"),
			Destination: &queenConfig.CrawlInterval,
			Value:       queenConfig.CrawlInterval,
		},
		&cli.IntFlag{
			Name:        "cache.size",
			Usage:       "How many agent versions and protocols should be cached in memory",
			Sources:     cli.EnvVars("ANTS_CACHE_SIZE"),
			Destination: &queenConfig.CacheSize,
			Value:       queenConfig.CacheSize,
		},
		&cli.StringFlag{
			Name:        "key.path",
			Usage:       "The path to the data store containing the keys",
			Sources:     cli.EnvVars("ANTS_KEY_PATH"),
			Destination: &queenConfig.KeyDBPath,
			Value:       queenConfig.KeyDBPath,
		},
		&cli.StringFlag{
			Name:        "certs.path",
			Usage:       "The path where we store the TLC certificates",
			Sources:     cli.EnvVars("ANTS_CERTS_PATH"),
			Destination: &queenConfig.CertsPath,
			Value:       queenConfig.CertsPath,
		},
		&cli.IntFlag{
			Name:        "first.port",
			Usage:       "First port ants can listen on",
			Sources:     cli.EnvVars("ANTS_FIRST_PORT"),
			Destination: &queenConfig.FirstPort,
			Value:       queenConfig.FirstPort,
		},
		&cli.IntFlag{
			Name:        "num.ports",
			Usage:       "Number of ports ants can listen on",
			Sources:     cli.EnvVars("ANTS_NUM_PORTS"),
			Destination: &queenConfig.NumPorts,
			Value:       queenConfig.NumPorts,
		},
		&cli.BoolFlag{
			Name:        "upnp",
			Usage:       "Enable UPnP",
			Sources:     cli.EnvVars("ANTS_UPNP"),
			Destination: &queenConfig.UPnp,
			Value:       queenConfig.UPnp,
		},
		&cli.IntFlag{
			Name:        "bucket.size",
			Usage:       "The bucket size for the ants DHT",
			Sources:     cli.EnvVars("ANTS_BUCKET_SIZE"),
			Destination: &queenConfig.BucketSize,
			Value:       queenConfig.BucketSize,
		},
		&cli.StringFlag{
			Name:        "user.agent",
			Usage:       "The user agent to use for the ants hosts",
			Sources:     cli.EnvVars("ANTS_USER_AGENT"),
			Destination: &queenConfig.UserAgent,
			Value:       queenConfig.UserAgent,
		},
		&cli.DurationFlag{
			Name:        "throttle.timeout",
			Usage:       "Time to throttle requests from the same identified peer (0 to disable)",
			Sources:     cli.EnvVars("ANTS_THROTTLE_TIMEOUT"),
			Destination: &queenConfig.ThrottleTimeout,
			Value:       queenConfig.ThrottleTimeout,
		},
	}
	flags = append(flags, gccli.ClickHouseFlags("ants", chCfg)...)
	flags = append(flags, gccli.ClickHouseMigrationsFlags("ANTS_", migrationsCfg)...)

	return &cli.Command{
		Name:   "queen",
		Usage:  "Starts the queen service",
		Flags:  flags,
		Action: runQueenCommand(chCfg, migrationsCfg),
	}
}

func runQueenCommand(chCfg *gcdb.ClickHouseConfig, migrationsCfg *gcdb.ClickHouseMigrationsConfig) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		telemetry, err := metrics.NewTelemetry()
		if err != nil {
			return fmt.Errorf("init telemetry: %w", err)
		}

		// Apply pending migrations before writing, unless running without a
		// ClickHouse backend (empty host selects the no-op writer).
		if chCfg.BaseConfig.Host != "" {
			if err := chCfg.Validate(); err != nil {
				return fmt.Errorf("clickhouse config: %w", err)
			}
			if err := migrationsCfg.Apply(chCfg.Options(), db.Migrations); err != nil {
				return fmt.Errorf("apply migrations: %w", err)
			}
		}

		writer, err := newRequestWriter(ctx, chCfg)
		if err != nil {
			return err
		}

		if !c.IsSet("user.agent") {
			queenConfig.UserAgent = ants.UserAgent(ants.Network(queenConfig.Network))
		}

		options := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		nebulaSvcAddr := net.JoinHostPort(queenConfig.NebulaSvcHost, fmt.Sprint(queenConfig.NebulaSvcPort))
		nebulaConn, err := grpc.NewClient(nebulaSvcAddr, options...)
		if err != nil {
			return fmt.Errorf("new gRPC Nebula connection client %s: %w", nebulaSvcAddr, err)
		}
		defer func() {
			if err := nebulaConn.Close(); err != nil {
				slog.Error("failed to close gRPC Nebula client", "err", err)
			}
		}()

		nebulaClient := nebulav1.NewNebulaServiceClient(nebulaConn)
		slog.Info("Initialized Nebula service client", "addr", nebulaSvcAddr)

		queenCfg := &ants.QueenConfig{
			KeysDBPath:      queenConfig.KeyDBPath,
			CertsPath:       queenConfig.CertsPath,
			NPorts:          queenConfig.NumPorts,
			FirstPort:       queenConfig.FirstPort,
			UPnP:            queenConfig.UPnp,
			CrawlInterval:   queenConfig.CrawlInterval,
			CacheSize:       queenConfig.CacheSize,
			BucketSize:      queenConfig.BucketSize,
			UserAgent:       queenConfig.UserAgent,
			ThrottleTimeout: queenConfig.ThrottleTimeout,
			BootstrapPeers:  ants.BootstrapPeers(ants.Network(queenConfig.Network)),
			ProtocolID:      ants.ProtocolID(ants.Network(queenConfig.Network)),
			Telemetry:       telemetry,
		}

		queen, err := ants.NewQueen(writer, nebulaClient, queenCfg)
		if err != nil {
			return fmt.Errorf("create queen: %w", err)
		}

		if err := queen.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("queen run: %w", err)
		}

		return nil
	}
}

// newRequestWriter returns a no-op writer when no ClickHouse host is configured,
// otherwise an async BatchInserter connected to the "requests" table.
func newRequestWriter(ctx context.Context, chCfg *gcdb.ClickHouseConfig) (db.RequestWriter, error) {
	if chCfg.BaseConfig.Host == "" {
		slog.Warn("No clickhouse host provided, using no-op writer")
		return db.NewNoopWriter(), nil
	}

	conn, err := chCfg.OpenAndPing(ctx)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	inserterCfg := gcdb.DefaultBatchInserterConfig[db.Request]()
	inserterCfg.MaxBatchSize = queenConfig.BatchSize
	inserterCfg.FlushInterval = queenConfig.BatchTime

	inserter, err := gcdb.NewBatchInserter[db.Request](conn, "requests", inserterCfg)
	if err != nil {
		return nil, fmt.Errorf("new batch inserter: %w", err)
	}

	return inserter, nil
}
