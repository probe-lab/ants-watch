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

func runQueen(ctx context.Context, nebulaPostgresStr string, nPorts, firstPort int, upnp bool, clickhouseClient *db.Client) error {
	var queen *ants.Queen
	var err error

	keyDBPath := os.Getenv("KEY_DB_PATH")
	if len(keyDBPath) == 0 {
		keyDBPath = "keys.db"
	}

	if upnp {
		queen, err = ants.NewQueen(ctx, nebulaPostgresStr, keyDBPath, 0, 0, clickhouseClient)
	} else {
		queen, err = ants.NewQueen(ctx, nebulaPostgresStr, keyDBPath, uint16(nPorts), uint16(firstPort), clickhouseClient)
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
						Name:    "clickhouseAddress",
						Usage:   "ClickHouse address containing the host and port, 127.0.0.1:9000",
						EnvVars: []string{"CLICKHOUSE_ADDRESS"},
					},
					&cli.StringFlag{
						Name:    "clickhouseDatabase",
						Usage:   "The ClickHouse database where ants requests will be recorded",
						EnvVars: []string{"CLICKHOUSE_DATABASE"},
					},
					&cli.StringFlag{
						Name:    "clickhouseUsername",
						Usage:   "The ClickHouse user that has the prerequisite privileges to record the requests",
						EnvVars: []string{"CLICKHOUSE_USERNAME"},
					},
					&cli.StringFlag{
						Name:    "clickhousePassword",
						Usage:   "The password for the ClickHouse user",
						EnvVars: []string{"CLICKHOUSE_PASSWORD"},
					},
					&cli.StringFlag{
						Name:    "nebulaDatabaseConnString",
						Usage:   "The connection string for the Postgres Nebula database",
						EnvVars: []string{"NEBULA_DB_CONNSTRING"},
					},
					&cli.IntFlag{
						Name:  "nPorts",
						Value: 128,
						Usage: "Number of ports ants can listen on",
					},
					&cli.IntFlag{
						Name:  "firstPort",
						Value: 6000,
						Usage: "First port ants can listen on",
					},
					&cli.BoolFlag{
						Name:  "upnp",
						Value: false,
						Usage: "Enable UPnP",
					},
				},
				Action: func(c *cli.Context) error {
					return runQueenCommand(c)
				},
			},
			{
				Name:  "health",
				Usage: "Checks the health of the service",
				Action: func(c *cli.Context) error {
					return healthCheckCommand()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Warnf("Error running app: %v\n", err)
		os.Exit(1)
	}

	logger.Debugln("Work is done")
}

func runQueenCommand(c *cli.Context) error {
	nebulaPostgresStr := c.String("nebulaDatabaseConnString")
	nPorts := c.Int("nPorts")
	firstPort := c.Int("firstPort")
	upnp := c.Bool("upnp")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	address := c.String("clickhouseAddress")
	database := c.String("clickhouseDatabase")
	username := c.String("clickhouseUsername")
	password := c.String("clickhousePassword")

	client, err := db.NewDatabaseClient(
		ctx, address, database, username, password,
	)
	if err != nil {
		logger.Errorln(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		errChan <- runQueen(ctx, nebulaPostgresStr, nPorts, firstPort, upnp, client)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			logger.Error(err)
			return err
		}
	case sig := <-sigChan:
		logger.Infof("Received signal: %v, initiating shutdown...", sig)
		cancel()
		<-errChan
	}
	return nil
}

func healthCheckCommand() error {
	ctx := context.Background()
	if err := HealthCheck(&ctx); err != nil {
		logger.Infof("Health check failed: %v\n", err)
		return err
	}
	logger.Infoln("Health check passed")
	return nil
}
