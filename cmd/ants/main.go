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

var queenConfig = struct {
	MetricsHost        string
	MetricsPort        int
	ClickhouseAddress  string
	ClickhouseDatabase string
	ClickhouseUsername string
	ClickhousePassword string
	ClickhouseSSL      bool
	NebulaDBConnString string
	KeyDBPath          string
	CertsPath          string
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
	CertsPath:          "p2p-forge-certs",
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
						Destination: &queenConfig.MetricsHost,
						Value:       queenConfig.MetricsHost,
					},
					&cli.IntFlag{
						Name:        "metrics.port",
						Usage:       "On which port to expose the metrics",
						EnvVars:     []string{"ANTS_METRICS_PORT"},
						Destination: &queenConfig.MetricsPort,
						Value:       queenConfig.MetricsPort,
					},
					&cli.StringFlag{
						Name:        "clickhouse.address",
						Usage:       "ClickHouse address containing the host and port, 127.0.0.1:9000",
						EnvVars:     []string{"ANTS_CLICKHOUSE_ADDRESS"},
						Destination: &queenConfig.ClickhouseAddress,
						Value:       queenConfig.ClickhouseAddress,
					},
					&cli.StringFlag{
						Name:        "clickhouse.database",
						Usage:       "The ClickHouse database where ants requests will be recorded",
						EnvVars:     []string{"ANTS_CLICKHOUSE_DATABASE"},
						Destination: &queenConfig.ClickhouseDatabase,
						Value:       queenConfig.ClickhouseDatabase,
					},
					&cli.StringFlag{
						Name:        "clickhouse.username",
						Usage:       "The ClickHouse user that has the prerequisite privileges to record the requests",
						EnvVars:     []string{"ANTS_CLICKHOUSE_USERNAME"},
						Destination: &queenConfig.ClickhouseUsername,
						Value:       queenConfig.ClickhouseUsername,
					},
					&cli.StringFlag{
						Name:        "clickhouse.password",
						Usage:       "The password for the ClickHouse user",
						EnvVars:     []string{"ANTS_CLICKHOUSE_PASSWORD"},
						Destination: &queenConfig.ClickhousePassword,
						Value:       queenConfig.ClickhousePassword,
					},
					&cli.BoolFlag{
						Name:        "clickhouse.ssl",
						Usage:       "Whether to use SSL for the ClickHouse connection",
						EnvVars:     []string{"ANTS_CLICKHOUSE_SSL"},
						Destination: &queenConfig.ClickhouseSSL,
						Value:       queenConfig.ClickhouseSSL,
					},
					&cli.StringFlag{
						Name:        "nebula.connstring",
						Usage:       "The connection string for the Postgres Nebula database",
						EnvVars:     []string{"ANTS_NEBULA_CONNSTRING"},
						Destination: &queenConfig.NebulaDBConnString,
						Value:       queenConfig.NebulaDBConnString,
					},
					&cli.IntFlag{
						Name:        "batch.size",
						Usage:       "The number of ants to request to store at a time",
						EnvVars:     []string{"ANTS_BATCH_SIZE"},
						Destination: &queenConfig.BatchSize,
						Value:       queenConfig.BatchSize,
					},
					&cli.DurationFlag{
						Name:        "batch.time",
						Usage:       "The time to wait between batches",
						EnvVars:     []string{"ANTS_BATCH_TIME"},
						Destination: &queenConfig.BatchTime,
						Value:       queenConfig.BatchTime,
					},
					&cli.DurationFlag{
						Name:        "crawl.interval",
						Usage:       "The time between two crawls",
						EnvVars:     []string{"ANTS_CRAWL_INTERVAL"},
						Destination: &queenConfig.CrawlInterval,
						Value:       queenConfig.CrawlInterval,
					},
					&cli.IntFlag{
						Name:        "cache.size",
						Usage:       "How many agent versions and protocols should be cached in memory",
						EnvVars:     []string{"ANTS_CACHE_SIZE"},
						Destination: &queenConfig.CacheSize,
						Value:       queenConfig.CacheSize,
					},
					&cli.PathFlag{
						Name:        "key.path",
						Usage:       "The path to the data store containing the keys",
						EnvVars:     []string{"ANTS_KEY_PATH"},
						Destination: &queenConfig.KeyDBPath,
						Value:       queenConfig.KeyDBPath,
					},
					&cli.PathFlag{
						Name:        "certs.path",
						Usage:       "The path where we store the TLC certificates",
						EnvVars:     []string{"ANTS_CERTS_PATH"},
						Destination: &queenConfig.CertsPath,
						Value:       queenConfig.CertsPath,
					},
					&cli.IntFlag{
						Name:        "first_port",
						Usage:       "First port ants can listen on",
						EnvVars:     []string{"ANTS_FIRST_PORT"},
						Destination: &queenConfig.FirstPort,
						Value:       queenConfig.FirstPort,
					},
					&cli.IntFlag{
						Name:        "num_ports",
						Usage:       "Number of ports ants can listen on",
						EnvVars:     []string{"ANTS_NUM_PORTS"},
						Destination: &queenConfig.NumPorts,
						Value:       queenConfig.NumPorts,
					},
					&cli.BoolFlag{
						Name:        "upnp",
						Usage:       "Enable UPnP",
						EnvVars:     []string{"ANTS_UPNP"},
						Destination: &queenConfig.UPnp,
						Value:       queenConfig.UPnp,
					},
					&cli.IntFlag{
						Name:        "bucket.size",
						Usage:       "The bucket size for the ants DHT",
						EnvVars:     []string{"ANTS_BUCKET_SIZE"},
						Destination: &queenConfig.BucketSize,
						Value:       queenConfig.BucketSize,
					},
					&cli.StringFlag{
						Name:        "user.agent",
						Usage:       "The user agent to use for the ants hosts",
						EnvVars:     []string{"ANTS_USER_AGENT"},
						Destination: &queenConfig.UserAgent,
						Value:       queenConfig.UserAgent,
					},
					&cli.StringFlag{
						Name:        "queen.id",
						Usage:       "The ID for the queen that's orchestrating the ants",
						EnvVars:     []string{"ANTS_QUEEN_ID"},
						Destination: &queenConfig.QueenID,
						Value:       queenConfig.QueenID,
						DefaultText: "generated",
					},
				},
				Action: runQueenCommand,
			},
			{
				Name:   "health",
				Usage:  "Checks the health of the service",
				Action: HealthCheck,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "metrics.host",
						Usage:       "On which host to expose the metrics",
						EnvVars:     []string{"ANTS_METRICS_HOST"},
						Destination: &healthConfig.MetricsHost,
						Value:       healthConfig.MetricsHost,
					},
					&cli.IntFlag{
						Name:        "metrics.port",
						Usage:       "On which port to expose the metrics",
						EnvVars:     []string{"ANTS_METRICS_PORT"},
						Destination: &healthConfig.MetricsPort,
						Value:       healthConfig.MetricsPort,
					},
				},
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

	logger.Debugln("Starting metrics server", "host", queenConfig.MetricsHost, "port", queenConfig.MetricsPort)
	go metrics.ListenAndServe(queenConfig.MetricsHost, queenConfig.MetricsPort)

	var client db.Client
	if queenConfig.ClickhouseAddress == "" {
		logger.Warn("No clickhouse address provided, using no-op client.")
		client = db.NewNoopClient()
	} else {
		// initializing a new clickhouse client
		client, err = db.NewClickhouseClient(
			queenConfig.ClickhouseAddress,
			queenConfig.ClickhouseDatabase,
			queenConfig.ClickhouseUsername,
			queenConfig.ClickhousePassword,
			queenConfig.ClickhouseSSL,
			telemetry,
		)
		if err != nil {
			return fmt.Errorf("init database client: %w", err)
		}

	}

	// pinging database to check availability
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err = client.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping clickhouse: %w", err)
	}

	queenCfg := &ants.QueenConfig{
		KeysDBPath:         queenConfig.KeyDBPath,
		CertsPath:          queenConfig.CertsPath,
		NPorts:             queenConfig.NumPorts,
		FirstPort:          queenConfig.FirstPort,
		UPnP:               queenConfig.UPnp,
		BatchSize:          queenConfig.BatchSize,
		BatchTime:          queenConfig.BatchTime,
		CrawlInterval:      queenConfig.CrawlInterval,
		CacheSize:          queenConfig.CacheSize,
		NebulaDBConnString: queenConfig.NebulaDBConnString,
		BucketSize:         queenConfig.BucketSize,
		UserAgent:          queenConfig.UserAgent,
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
