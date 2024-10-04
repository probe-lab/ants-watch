package main

import (
	"context"
	"flag"

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

	ctx := context.Background()
	var queen *ants.Queen
	var err error
	if *upnp {
		queen, err = ants.NewQueen(*postgresStr, "keys.db", 0, 0)
	} else {
		queen, err = ants.NewQueen(*postgresStr, "keys.db", uint16(*nPorts), uint16(*firstPort))
	}
	if err != nil {
		panic(err)
	}

	go queen.Run(ctx)

	<-ctx.Done()
}
