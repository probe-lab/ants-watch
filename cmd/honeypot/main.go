package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/probe-lab/ants-watch"
	"github.com/probe-lab/ants-watch/metrics"
)

var logger = logging.Logger("ants-queen")

func runQueen(ctx context.Context, nebulaPostgresStr string, nPorts, firstPort int, upnp bool) error {
	var queen *ants.Queen
	var err error

	keyDBPath := os.Getenv("KEY_DB_PATH")
	if len(keyDBPath) == 0 {
		keyDBPath = "keys.db"
	}

	if upnp {
		queen, err = ants.NewQueen(ctx, nebulaPostgresStr, keyDBPath, 0, 0)
	} else {
		queen, err = ants.NewQueen(ctx, nebulaPostgresStr, keyDBPath, uint16(nPorts), uint16(firstPort))
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

	queenCmd := flag.NewFlagSet("queen", flag.ExitOnError)
	nebulaPostgresStr := queenCmd.String("postgres", "", "Postgres connection string, postgres://user:password@host:port/dbname")
	nPorts := queenCmd.Int("nPorts", 128, "Number of ports ants can listen on")
	firstPort := queenCmd.Int("firstPort", 6000, "First port ants can listen on")
	upnp := queenCmd.Bool("upnp", false, "Enable UPnP")

	healthCmd := flag.NewFlagSet("health", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("Expected 'queen' or 'health' subcommands")
		os.Exit(1)
	}

	if os.Args[1] != "health" {
		metricsHost := os.Getenv("METRICS_HOST")
		metricsPort := os.Getenv("METRICS_PORT")

		p, err := strconv.Atoi(metricsPort)
		if err != nil {
			logger.Errorf("Port should be an int %v\n", metricsPort)
		}
		logger.Infoln("Serving metrics endpoint")
		go metrics.ListenAndServe(metricsHost, p)
	}

	switch os.Args[1] {
	case "queen":
		queenCmd.Parse(os.Args[2:])

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		errChan := make(chan error, 1)
		go func() {
			errChan <- runQueen(ctx, *nebulaPostgresStr, *nPorts, *firstPort, *upnp)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				logger.Error(err)
				os.Exit(1)
			}
		case sig := <-sigChan:
			logger.Infof("Received signal: %v, initiating shutdown...", sig)
			cancel()
			<-errChan
		}

	case "health":
		healthCmd.Parse(os.Args[2:])

		ctx := context.Background()
		if err := HealthCheck(&ctx); err != nil {
			fmt.Printf("Health check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Health check passed")

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	logger.Debugln("Work is done")
}
