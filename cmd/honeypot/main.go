package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/probe-lab/ants-watch"
)

var logger = logging.Logger("ants-queen")

func main() {
	logging.SetLogLevel("ants-queen", "debug")
	logging.SetLogLevel("dht", "error")
	logging.SetLogLevel("basichost", "info")

	nebulaPostgresStr := flag.String("postgres", "", "Postgres connection string, postgres://user:password@host:port/dbname")
	nPorts := flag.Int("nPorts", 128, "Number of ports ants can listen on")
	firstPort := flag.Int("firstPort", 6000, "First port ants can listen on")
	upnp := flag.Bool("upnp", false, "Enable UPnP")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var queen *ants.Queen
	var err error
	if *upnp {
		queen, err = ants.NewQueen(ctx, *nebulaPostgresStr, "keys.db", 0, 0)
	} else {
		queen, err = ants.NewQueen(ctx, *nebulaPostgresStr, "keys.db", uint16(*nPorts), uint16(*firstPort))
	}
	if err != nil {
		panic(err)
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
			logger.Errorf("Queen.Run returned an error: %v", err)
		} else {
			logger.Debugln("Queen.Run completed successfully")
		}
	case sig := <-sigChan:
		logger.Infof("Received signal: %v, initiating shutdown...", sig)
	}

	cancel()

	select {
	case <-errChan:
		logger.Debugln("Queen.Run stopped after context cancellation")
	case <-time.After(30 * time.Second):
		logger.Warnln("Timeout waiting for Queen.Run to stop")
	}

	logger.Debugln("Work is done")
}
