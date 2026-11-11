# ants-watch

ProbeLab repository. Follow the `probelab` skill: it holds the company rules. Load it before you start any task here.

`ants-watch` records the DHT clients of a libp2p Kademlia network. The **queen** spreads lightweight DHT server nodes, the **ants**, across the keyspace so that every routing table refresh of a client reaches at least one ant. Each request an ant receives goes to the ClickHouse table `requests`. The queen asks the Nebula service (gRPC) for the current DHT servers and moves the ants as the network changes. It watches Celestia, Avail and AGNTCY; ProbeLab runs it for `celestia/mainnet` and `agntcy/mainnet`.

## Layout

- `cmd/ants/` — CLI: `queen` and `health` commands and their flags (`main.go`, `health.go`)
- `*.go` at the root — package `ants`: the queen (`queen.go`), one ant (`ant.go`), ant keys and the file that stores them (`keys.go`, `ANTS_KEY_PATH`), networks and bootstrap peers (`bootstrappers.go`), public IPv4 lookup (`publicips.go`)
- `db/` — the `requests` row type, writers, and the embedded ClickHouse migrations (`db/migrations/`)
- `metrics/` — OpenTelemetry instruments
- `proto/nebula/v1/` — generated gRPC client of the Nebula service, copied from the `api` repo; never edit by hand
- `resources/` — README images

## Commands

- `just build`, `just test`, `just lint`, `just format`, `just ci` (the checks CI runs)
- `just local-clickhouse` starts ClickHouse in Docker; `cp .env.example .env` gives matching settings
- Run locally: `go run ./cmd/ants queen --upnp`, or with forwarded ports `--first.port=<port> --num.ports=<count>`. The queen needs a Nebula service at `ANTS_NEBULA_SERVICE_HOST:ANTS_NEBULA_SERVICE_PORT` and a public IP.
- `go run ./cmd/ants health` checks `/healthz` on the metrics port
- `just docker` builds the image as `ants-sha<commit>`; `just docker-push` pushes it to ECR
- `just deploy [commit]` deploys a commit to both ECS services through `ci.yml`

## Conventions

- A new network is an entry in `networks` in `bootstrappers.go`: project, network, protocol ID, user agent and bootstrap peers.
- A schema change is a new numbered `up`/`down` pair in `db/migrations/`. The queen applies migrations on startup. It strips the `Replicated` engine prefix unless `ANTS_CLICKHOUSE_MIGRATIONS_REPLICATED_TABLE_ENGINES=true`, which production sets.
- `just lint` runs staticcheck on every package except `proto/`, as in `api`.
- The packages are not yet under `pkg/`. Keep new code next to the package it belongs to until they move.
- The image tag is `ants-sha<commit>`. `tf/ants.tf` in `probelab-infra` pins it for both services.

## Related repositories

- `api` — serves the Nebula gRPC service the queen queries (`svc/nebula`) and holds the source of `proto/nebula/v1`
- `probelab-infra` — the ECS services and task definitions (`tf/modules/ants`, `tf/ants.tf`)
- `go-commons` — CLI setup, logging, metrics, ClickHouse connection, migrations and batch inserts

## Owner

@probe-lab/probelab-core
