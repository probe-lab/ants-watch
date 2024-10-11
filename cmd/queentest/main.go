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
	"github.com/probe-lab/ants-watch/db"
)

var logger = logging.Logger("ants-queen")

func main() {
	logging.SetLogLevel("ants-queen", "debug")
	logging.SetLogLevel("dht", "error")
	logging.SetLogLevel("basichost", "info")

	postgresStr := flag.String("postgres", "", "Postgres connection string, postgres://user:password@host:port/dbname")
	nPorts := flag.Int("nPorts", 128, "Number of ports ants can listen on")
	firstPort := flag.Int("firstPort", 6000, "First port ants can listen on")
	upnp := flag.Bool("upnp", false, "Enable UPnP")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var queen *ants.Queen
	if *upnp {
		queen = ants.NewQueen(ctx, *postgresStr, "keys.db", 0, 0)
	} else {
		queen = ants.NewQueen(ctx, *postgresStr, "keys.db", uint16(*nPorts), uint16(*firstPort))
	}

	go queen.Run(ctx)

	go func() {
		nctx, ncancel := context.WithCancel(context.Background())
		defer ncancel()

		logger.Info("Starting continuous normalization...")

		for {
			select {
			case <-nctx.Done():
				logger.Info("Normalization context canceled, stopping normalization loop.")
				return
			default:
				err := db.NormalizeRequests(nctx, queen.Client.Handler, queen.Client)
				if err != nil {

					logger.Errorf("Error during normalization: %w", err)
				} else {
					logger.Info("Normalization completed for current batch.")
				}

				// TODO: remove the hardcoded time
				time.Sleep(10 * time.Second)
			}
		}
	}()

	go func() {
		sig := <-sigChan
		logger.Infof("Received signal: %s, shutting down...", sig)
		cancel()
	}()

	<-ctx.Done()
	logger.Info("Context canceled, queen stopped")
}
