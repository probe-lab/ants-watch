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
	flag.Parse()

	ctx := context.Background()
	queen := ants.NewQueen(*postgresStr, "keys.db")

	go queen.Run(ctx)

	<-ctx.Done()
}
