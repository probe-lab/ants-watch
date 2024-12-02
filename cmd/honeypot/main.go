package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/probe-lab/ants-watch"
	"github.com/probe-lab/ants-watch/db"
	"github.com/urfave/cli/v2"
)

var logger = logging.Logger("ants-queen")

var rootConfig = struct {
	AntsClickhouseAddress  string
	AntsClickhouseDatabase string
	AntsClickhouseUsername string
	AntsClickhousePassword string
	AntsClickhouseSSL      bool
	NebulaDBConnString     string
	KeyDBPath              string
	NumPorts               int
	FirstPort              int
	UPnp                   bool
	BatchSize              int
	BatchTime              time.Duration
	CrawlInterval          time.Duration
	CacheSize              int
}{
	AntsClickhouseAddress:  "",
	AntsClickhouseDatabase: "",
	AntsClickhouseUsername: "",
	AntsClickhousePassword: "",
	AntsClickhouseSSL:      true,
	NebulaDBConnString:     "",
	KeyDBPath:              "keys.db",
	NumPorts:               128,
	FirstPort:              6000,
	UPnp:                   false,
	BatchSize:              1000,
	BatchTime:              time.Second,
	CrawlInterval:          120 * time.Minute,
	CacheSize:              10_000,
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
						Name:        "ants.clickhouse.address",
						Usage:       "ClickHouse address containing the host and port, 127.0.0.1:9000",
						EnvVars:     []string{"ANTS_CLICKHOUSE_ADDRESS"},
						Destination: &rootConfig.AntsClickhouseAddress,
						Value:       rootConfig.AntsClickhouseAddress,
					},
					&cli.StringFlag{
						Name:        "ants.clickhouse.database",
						Usage:       "The ClickHouse database where ants requests will be recorded",
						EnvVars:     []string{"ANTS_CLICKHOUSE_DATABASE"},
						Destination: &rootConfig.AntsClickhouseDatabase,
						Value:       rootConfig.AntsClickhouseDatabase,
					},
					&cli.StringFlag{
						Name:        "ants.clickhouse.username",
						Usage:       "The ClickHouse user that has the prerequisite privileges to record the requests",
						EnvVars:     []string{"ANTS_CLICKHOUSE_USERNAME"},
						Destination: &rootConfig.AntsClickhouseUsername,
						Value:       rootConfig.AntsClickhouseUsername,
					},
					&cli.StringFlag{
						Name:        "ants.clickhouse.password",
						Usage:       "The password for the ClickHouse user",
						EnvVars:     []string{"ANTS_CLICKHOUSE_PASSWORD"},
						Destination: &rootConfig.AntsClickhousePassword,
						Value:       rootConfig.AntsClickhousePassword,
					},
					&cli.BoolFlag{
						Name:        "ants.clickhouse.ssl",
						Usage:       "Whether to use SSL for the ClickHouse connection",
						EnvVars:     []string{"ANTS_CLICKHOUSE_SSL"},
						Destination: &rootConfig.AntsClickhouseSSL,
						Value:       rootConfig.AntsClickhouseSSL,
					},
					&cli.StringFlag{
						Name:        "nebula.db.connstring",
						Usage:       "The connection string for the Postgres Nebula database",
						EnvVars:     []string{"NEBULA_DB_CONNSTRING"},
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
						Name:    "key.db_path",
						Usage:   "The path to the data store containing the keys",
						EnvVars: []string{"KEY_DB_PATH"},
					},
					&cli.IntFlag{
						Name:  "num_ports",
						Value: 128,
						Usage: "Number of ports ants can listen on",
					},
					&cli.IntFlag{
						Name:  "first_port",
						Value: 6000,
						Usage: "First port ants can listen on",
					},
					&cli.BoolFlag{
						Name:  "upnp",
						Value: false,
						Usage: "Enable UPnP",
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

	// initializing new clickhouse client
	client, err := db.NewClient(
		rootConfig.AntsClickhouseAddress,
		rootConfig.AntsClickhouseDatabase,
		rootConfig.AntsClickhouseUsername,
		rootConfig.AntsClickhousePassword,
		rootConfig.AntsClickhouseSSL,
	)
	if err != nil {
		logger.Errorln(err)
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
	}

	// initializting queen
	queen, err := ants.NewQueen(client, queenCfg)
	if err != nil {
		return fmt.Errorf("failed to create queen: %w", err)
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
