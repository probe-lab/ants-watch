# Prints available recipes
default:
    @just --list --justfile {{ justfile() }}

repo := env_var_or_default("REPO", "019120760881.dkr.ecr.us-east-1.amazonaws.com")
repo_user := env_var_or_default("REPO_USER", "AWS")
repo_region := env_var_or_default("REPO_REGION", "us-east-1")
tag := env_var_or_default("TAG", `date +%F` + "-" + `echo -n "${ANTS_REPO:-github.com/probe-lab/ashby}-${ANTS_COMMIT:-839c2f242c67825ff7884515e929c7ba5b16e84e}" | sha256sum | cut -c -8` + "-" + `git describe --always --tag --dirty`)
image_name := env_var_or_default("IMAGE_NAME", "probelab:ants-" + tag)

# Start a local single-node clickhouse for development
local-clickhouse:
    docker run --name ants-clickhouse --rm -p 9000:9000 -p 8123:8123 -e CLICKHOUSE_DB=ants_local -e CLICKHOUSE_USER=ants_local -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -e CLICKHOUSE_PASSWORD=password clickhouse/clickhouse-server

# Strip the `Replicated` engine prefix into db/migrations/local (needed for single-node clickhouse)
non-cluster-migrations:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p db/migrations/local
    for file in $(find db/migrations -maxdepth 1 -name "*.sql"); do
        sed 's/Replicated//' "$file" > "db/migrations/local/$(basename "$file")"
    done

# Roll back local migrations (the queen applies up-migrations on startup)
local-migrate-down: non-cluster-migrations
    migrate -database 'clickhouse://localhost:9000?username=ants_local&database=ants_local&password=password&x-multi-statement=true' -path db/migrations/local down

# Format, vet, and lint the codebase
check:
    gofmt -w .
    go vet ./...
    golangci-lint run

# Build the linux/amd64 docker image
build:
    @echo "{{ tag }}"
    docker build --platform="linux/amd64" -t "{{ image_name }}" .

# Push the docker image to the ECR registry
push:
    aws ecr get-login-password --region {{ repo_region }} | docker login --username {{ repo_user }} --password-stdin {{ repo }}
    docker tag "{{ image_name }}" "{{ repo }}/{{ image_name }}"
    docker push "{{ repo }}/{{ image_name }}"
