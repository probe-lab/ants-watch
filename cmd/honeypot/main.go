package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

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

	go queen.Run(ctx)

	go func() {
		sig := <-sigChan
		logger.Infof("Received signal: %s, shutting down...", sig)
		cancel()
	}()

	<-ctx.Done()
	logger.Info("Context canceled, queen stopped")
}
