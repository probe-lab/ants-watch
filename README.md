# Ants Watch

[![ProbeLab](https://img.shields.io/badge/made%20by-ProbeLab-blue.svg)](https://probelab.io)
![Build Status](https://img.shields.io/github/actions/workflow/status/probe-lab/ants-watch/ci.yml?branch=main)
![License](https://img.shields.io/github/license/probe-lab/ants-watch)

DHT Client Population Monitor.

<img src="./resources/ants.png" alt="Ants Watch" height="300"/>

Authors: [guillaumemichel](https://github.com/guillaumemichel), [kasteph](https://github.com/kasteph)

## Overview

* `ants-watch` is a DHT honeypot monitoring tool, logging the activity of all nodes in a DHT network.
* An `ant` is a lightweight [libp2p DHT node](https://github.com/libp2p/go-libp2p-kad-dht), participating in the DHT network, and logging incoming requests.
* `ants` participate in the DHT network as DHT server nodes. `ants-watch` needs to be dialable by other nodes in the network. Hence, it must run on a public IP address either with port forwarding properly configured (including local and gateway firewalls) or UPnP enabled.
* The tool releases `ants` (i.e., spawns new `ant` nodes) at targeted locations in the keyspace in order to _occupy_ and _watch_ the full keyspace.
* The tool's logic is based on the fact that peer routing requests are distributed to `k` closest nodes in the keyspace and routing table updates by DHT client (and server) nodes need to find the `k` closest DHT server peers to themselves. Therefore, placing approximately 1 `ant` node every `k` DHT server nodes can capture all DHT client nodes over time.
* The routing table update process varies across implementations, but is by default set to 10 mins in the go-libp2p implementation. This means that `ants` will record the existence of DHT client nodes approximately every 10 mins (or whatever the routing table update interval is).
* Depending on the network size, the number of `ants` as well as their location in the keyspace is adjusted automatically.
* Network size and peers distribution is obtained by querying an external [Nebula database](https://github.com/dennis-tra/nebula).
* All `ants` run from within the same process, sharing the same DHT records.
* The `ant queen` is responsible for spawning, adjusting the number and monitoring the ants as well as gathering their logs and persisting them to a central database.
* `ants-watch` does not operate like a crawler, where after one run the number of DHT client nodes is captured. `ants-watch` logs all received DHT requests and therefore, it must run continuously to provide the number of DHT client nodes over time.

### Supported networks

* [Celestia](https://celestia.org/)
* Can be extended to support other networks using the [libp2p DHT](https://github.com/libp2p/specs/tree/master/kad-dht).

## Setup

Before installing dependencies:

``` shell
git submodule init
git submodule update --init --recursive --remote
```

Then, `go mod tidy`.

You'll also need to install some tools: `make tools`.

You need to setup a postgres database with:

* username: `ants-watch`
* password: `password` <-- you should change this in the [Makefile](./Makefile)
* database: `ants-watch`

Migrations can now be applied: `make migrate-up`.

## Configuration

The following environment variables should be set:

```sh
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=ants-watch
DB_USER=ants-watch
DB_PASSWORD=password
DB_SSLMODE=disable

NEBULA_POSTGRES_CONNURL=postgres://nebula:password@localhost/nebula?sslmode=disable

UDGER_FILEPATH=/location/to/udgerdb.dat               # optional, used to detect datacenters
MAXMIND_ASN_DB=/location/to/GeoLite2-ASN.mmdb         # optional, used to extract ASN from IP
MAXMIND_COUNRTY_DB=/location/to/GeoLite2-Country.mmdb # optional, used to extract country from IP

KEY_DB_PATH=/location/to/keys.db  # optional, used to store the generated libp2p keys

METRICS_HOST=0.0.0.0  # optional, used to expose metrics
METRICS_PORT=6666     # optional, used to expose metrics

TRACES_HOST=0.0.0.0   # optional, used to expose traces
TRACES_PORT=6667      # optional, used to expose traces

BATCH_TIME=20 # optional, time in seconds between each batch of requests
```

## Usage

Once the database is setup and migrations are applied, you can start the honeypot.

`ants-watch` needs to be dialable by other nodes in the network. Hence, it must run on a public IP address either with port forwarding properly configured (including local and gateway firewalls) or UPnP enabled.

### Queen

In [`cmd/honeypot/`](./cmd/honeypot/), you can run the following command:

```sh
go run . queen -upnp=true # for UPnP
# or
go run . queen -firstPort=<port> -nPorts=<count> # for port forwarding
```

When UPnP is disabled, ports from `firstPort` to `firstPort + nPorts - 1` must be forwarded to the machine running `ants-watch`. `ants-watch` will be able to spawn at most `nPorts` distinct `ants`.

### Health

You can run a health check on the honeypot by running the following command:

```sh
go run . health
```

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
