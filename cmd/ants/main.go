package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/probe-lab/ants-watch"
	"github.com/probe-lab/ants-watch/db"
	"github.com/probe-lab/ants-watch/metrics"
)

var logger = logging.Logger("ants-queen")

var rootConfig = struct {
	MetricsHost        string
	MetricsPort        int
	ClickhouseAddress  string
	ClickhouseDatabase string
	ClickhouseUsername string
	ClickhousePassword string
	ClickhouseSSL      bool
	NebulaDBConnString string
	KeyDBPath          string
	NumPorts           int
	FirstPort          int
	UPnp               bool
	BatchSize          int
	BatchTime          time.Duration
	CrawlInterval      time.Duration
	CacheSize          int
	BucketSize         int
	UserAgent          string
	ProtocolPrefix     string
	QueenID            string
}{
	MetricsHost:        "127.0.0.1",
	MetricsPort:        5999, // one below the FirstPort to not accidentally override it
	ClickhouseAddress:  "",
	ClickhouseDatabase: "",
	ClickhouseUsername: "",
	ClickhousePassword: "",
	ClickhouseSSL:      true,
	NebulaDBConnString: "",
	KeyDBPath:          "keys.db",
	NumPorts:           128,
	FirstPort:          6000,
	UPnp:               false,
	BatchSize:          1000,
	BatchTime:          20 * time.Second,
	CrawlInterval:      120 * time.Minute,
	CacheSize:          10_000,
	BucketSize:         20,
	UserAgent:          "celestiant",
	QueenID:            "",
}

func main() {
	logging.SetLogLevel("ants-queen", "debug")
	logging.SetLogLevel("dht", "error")
	logging.SetLogLevel("basichost", "info")

	app := &cli.App{
		Name:  "ants-watch",
		Usage: "Get DHT clients in your p2p network using a honeypot",
		Commands: []*cli.Command{
			{
				Name:  "queen",
				Usage: "Starts the queen service",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "metrics.host",
						Usage:       "On which host to expose the metrics",
						EnvVars:     []string{"ANTS_METRICS_HOST"},
						Destination: &rootConfig.MetricsHost,
						Value:       rootConfig.MetricsHost,
					},
					&cli.IntFlag{
						Name:        "metrics.port",
						Usage:       "On which port to expose the metrics",
						EnvVars:     []string{"ANTS_METRICS_PORT"},
						Destination: &rootConfig.MetricsPort,
						Value:       rootConfig.MetricsPort,
					},
					&cli.StringFlag{
						Name:        "clickhouse.address",
						Usage:       "ClickHouse address containing the host and port, 127.0.0.1:9000",
						EnvVars:     []string{"ANTS_CLICKHOUSE_ADDRESS"},
						Destination: &rootConfig.ClickhouseAddress,
						Value:       rootConfig.ClickhouseAddress,
					},
					&cli.StringFlag{
						Name:        "clickhouse.database",
						Usage:       "The ClickHouse database where ants requests will be recorded",
						EnvVars:     []string{"ANTS_CLICKHOUSE_DATABASE"},
						Destination: &rootConfig.ClickhouseDatabase,
						Value:       rootConfig.ClickhouseDatabase,
					},
					&cli.StringFlag{
						Name:        "clickhouse.username",
						Usage:       "The ClickHouse user that has the prerequisite privileges to record the requests",
						EnvVars:     []string{"ANTS_CLICKHOUSE_USERNAME"},
						Destination: &rootConfig.ClickhouseUsername,
						Value:       rootConfig.ClickhouseUsername,
					},
					&cli.StringFlag{
						Name:        "clickhouse.password",
						Usage:       "The password for the ClickHouse user",
						EnvVars:     []string{"ANTS_CLICKHOUSE_PASSWORD"},
						Destination: &rootConfig.ClickhousePassword,
						Value:       rootConfig.ClickhousePassword,
					},
					&cli.BoolFlag{
						Name:        "clickhouse.ssl",
						Usage:       "Whether to use SSL for the ClickHouse connection",
						EnvVars:     []string{"ANTS_CLICKHOUSE_SSL"},
						Destination: &rootConfig.ClickhouseSSL,
						Value:       rootConfig.ClickhouseSSL,
					},
					&cli.StringFlag{
						Name:        "nebula.connstring",
						Usage:       "The connection string for the Postgres Nebula database",
						EnvVars:     []string{"ANTS_NEBULA_CONNSTRING"},
						Destination: &rootConfig.NebulaDBConnString,
						Value:       rootConfig.NebulaDBConnString,
					},
					&cli.IntFlag{
						Name:        "batch.size",
						Usage:       "The number of ants to request to store at a time",
						EnvVars:     []string{"ANTS_BATCH_SIZE"},
						Destination: &rootConfig.BatchSize,
						Value:       rootConfig.BatchSize,
					},
					&cli.DurationFlag{
						Name:        "batch.time",
						Usage:       "The time to wait between batches",
						EnvVars:     []string{"ANTS_BATCH_TIME"},
						Destination: &rootConfig.BatchTime,
						Value:       rootConfig.BatchTime,
					},
					&cli.DurationFlag{
						Name:        "crawl.interval",
						Usage:       "The time between two crawls",
						EnvVars:     []string{"ANTS_CRAWL_INTERVAL"},
						Destination: &rootConfig.CrawlInterval,
						Value:       rootConfig.CrawlInterval,
					},
					&cli.IntFlag{
						Name:        "cache.size",
						Usage:       "How many agent versions and protocols should be cached in memory",
						EnvVars:     []string{"ANTS_CACHE_SIZE"},
						Destination: &rootConfig.CacheSize,
						Value:       rootConfig.CacheSize,
					},
					&cli.PathFlag{
						Name:        "key.path",
						Usage:       "The path to the data store containing the keys",
						EnvVars:     []string{"ANTS_KEY_PATH"},
						Destination: &rootConfig.KeyDBPath,
						Value:       rootConfig.KeyDBPath,
					},
					&cli.IntFlag{
						Name:        "first_port",
						Usage:       "First port ants can listen on",
						EnvVars:     []string{"ANTS_FIRST_PORT"},
						Destination: &rootConfig.FirstPort,
						Value:       rootConfig.FirstPort,
					},
					&cli.IntFlag{
						Name:        "num_ports",
						Usage:       "Number of ports ants can listen on",
						EnvVars:     []string{"ANTS_NUM_PORTS"},
						Destination: &rootConfig.NumPorts,
						Value:       rootConfig.NumPorts,
					},
					&cli.BoolFlag{
						Name:        "upnp",
						Usage:       "Enable UPnP",
						EnvVars:     []string{"ANTS_UPNP"},
						Destination: &rootConfig.UPnp,
						Value:       rootConfig.UPnp,
					},
					&cli.IntFlag{
						Name:        "bucket.size",
						Usage:       "The bucket size for the ants DHT",
						EnvVars:     []string{"ANTS_BUCKET_SIZE"},
						Destination: &rootConfig.BucketSize,
						Value:       rootConfig.BucketSize,
					},
					&cli.StringFlag{
						Name:        "user.agent",
						Usage:       "The user agent to use for the ants hosts",
						EnvVars:     []string{"ANTS_USER_AGENT"},
						Destination: &rootConfig.UserAgent,
						Value:       rootConfig.UserAgent,
					},
					&cli.StringFlag{
						Name:        "queen.id",
						Usage:       "The ID for the queen that's orchestrating the ants",
						EnvVars:     []string{"ANTS_QUEEN_ID"},
						Destination: &rootConfig.QueenID,
						Value:       rootConfig.QueenID,
						DefaultText: "generated",
					},
				},
				Action: runQueenCommand,
			},
			{
				Name:   "health",
				Usage:  "Checks the health of the service",
				Action: HealthCheck,
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sctx, stop := signal.NotifyContext(ctx, syscall.SIGINT)
	defer stop()

	if err := app.RunContext(sctx, os.Args); err != nil {
		logger.Warnf("Error running app: %v\n", err)
		os.Exit(1)
	}

	logger.Debugln("Work is done")
}

func runQueenCommand(c *cli.Context) error {
	ctx := c.Context

	meterProvider, err := metrics.NewMeterProvider()
	if err != nil {
		return fmt.Errorf("init meter provider: %w", err)
	}

	telemetry, err := metrics.NewTelemetry(noop.NewTracerProvider(), meterProvider)
	if err != nil {
		return fmt.Errorf("init telemetry: %w", err)
	}

	logger.Debugln("Starting metrics server", "host", rootConfig.MetricsHost, "port", rootConfig.MetricsPort)
	go metrics.ListenAndServe(rootConfig.MetricsHost, rootConfig.MetricsPort)

	// initializing a new clickhouse client
	client, err := db.NewClient(
		rootConfig.ClickhouseAddress,
		rootConfig.ClickhouseDatabase,
		rootConfig.ClickhouseUsername,
		rootConfig.ClickhousePassword,
		rootConfig.ClickhouseSSL,
		telemetry,
	)
	if err != nil {
		return fmt.Errorf("init database client: %w", err)
	}

	// pinging database to check availability
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err = client.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping clickhouse: %w", err)
	}

	queenCfg := &ants.QueenConfig{
		KeysDBPath:         rootConfig.KeyDBPath,
		NPorts:             rootConfig.NumPorts,
		FirstPort:          rootConfig.FirstPort,
		UPnP:               rootConfig.UPnp,
		BatchSize:          rootConfig.BatchSize,
		BatchTime:          rootConfig.BatchTime,
		CrawlInterval:      rootConfig.CrawlInterval,
		CacheSize:          rootConfig.CacheSize,
		NebulaDBConnString: rootConfig.NebulaDBConnString,
		BucketSize:         rootConfig.BucketSize,
		UserAgent:          rootConfig.UserAgent,
		Telemetry:          telemetry,
	}

	// initializing queen
	queen, err := ants.NewQueen(client, queenCfg)
	if err != nil {
		return fmt.Errorf("create queen: %w", err)
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Debugln("Starting Queen.Run")
		errChan <- queen.Run(ctx)
		logger.Debugln("Queen.Run completed")
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("queen.Run returned an error: %w", err)
		}
		logger.Debugln("Queen.Run completed successfully")
	case <-ctx.Done():
		select {
		case <-errChan:
			logger.Debugln("Queen.Run stopped after context cancellation")
		case <-time.After(30 * time.Second):
			logger.Warnln("Timeout waiting for Queen.Run to stop")
		}
	}

	return nil
}
