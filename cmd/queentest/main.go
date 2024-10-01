package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/probe-lab/ants-watch"
)

func main() {
	log.SetLogLevel("ants-queen", "debug") // debug
	log.SetLogLevel("dht", "error")        // warn
	log.SetLogLevel("basichost", "info")
	// log.SetLogLevel("nat", "debug")

	postgresStr := flag.String("postgres", "", "Postgres connection string, postgres://user:password@host:port/dbname")
	nPorts := flag.Int("nPorts", 128, "Number of ports ants can listen on")
	firstPort := flag.Int("firstPort", 6000, "First port ants can listen on")
	upnp := flag.Bool("upnp", false, "Enable UPnP")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	var queen *ants.Queen
	if *upnp {
		queen = ants.NewQueen(ctx, *postgresStr, "keys.db", 0, 0)
	} else {
		queen = ants.NewQueen(ctx, *postgresStr, "keys.db", uint16(*nPorts), uint16(*firstPort))
	}

	go queen.Run(ctx)

	// Simulate some condition to cancel the context after 5 minutes
	time.Sleep(30 * time.Second)
	cancel() // This will stop the Queen

	fmt.Println("Context canceled manually")
}
