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

func runQueen(ctx context.Context, clickhouseClient *db.Client) error {
	var queen *ants.Queen
	var err error

	if rootConfig.UPnp {
		queen, err = ants.NewQueen(ctx, rootConfig.NebulaDBConnString, rootConfig.KeyDBPath, 0, 0, clickhouseClient)
	} else {
		queen, err = ants.NewQueen(ctx, rootConfig.NebulaDBConnString, rootConfig.KeyDBPath, uint16(rootConfig.NumPorts), uint16(rootConfig.FirstPort), clickhouseClient)
	}
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
						Name:    "ants.clickhouse.password",
						Usage:   "The password for the ClickHouse user",
						EnvVars: []string{"ANTS_CLICKHOUSE_PASSWORD"},
					},
					&cli.StringFlag{
						Name:    "nebula.db.connstring",
						Usage:   "The connection string for the Postgres Nebula database",
						EnvVars: []string{"NEBULA_DB_CONNSTRING"},
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
	client, err := db.NewDatabaseClient(
		c.Context,
		rootConfig.AntsClickhouseAddress,
		rootConfig.AntsClickhouseDatabase,
		rootConfig.AntsClickhouseUsername,
		rootConfig.AntsClickhousePassword,
	)

	if err != nil {
		logger.Errorln(err)
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- runQueen(c.Context, client)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}
