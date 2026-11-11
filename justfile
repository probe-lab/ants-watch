# Task runner. `just` lists the recipes; CI runs `just ci`.

# The lint tools, pinned here so a person and CI run the same versions.
# staticcheck skips the generated code in proto/, as in the api repo.
gofumpt := "go run mvdan.cc/gofumpt@v0.12.0"
staticcheck := "go run honnef.co/go/tools/cmd/staticcheck@2026.2.2"

# The image CI pushes for a commit: ants-sha<commit>. A dirty tree gets -dirty.
registry := "019120760881.dkr.ecr.us-east-1.amazonaws.com"
image := registry + "/probelab:ants-sha" + `git rev-parse --short=7 HEAD` + `git diff --quiet HEAD || echo -dirty`

# List the recipes
default:
    @just --list

# Build every package and command
build:
    go build ./...

# Run the tests
test *args:
    go test ./... {{ args }}

# Check formatting, go fix, vet, and staticcheck without changing files
lint:
    @test -z "$({{ gofumpt }} -l .)" || { echo "not formatted, run just format:"; {{ gofumpt }} -l .; exit 1; }
    @d=$(go fix -diff ./...) && test -z "$d" || { echo "$d"; echo "go fix changes pending, run just format"; exit 1; }
    go vet ./...
    {{ staticcheck }} $(go list ./... | grep -v /proto/)

# Format the code and apply go fix in place
format:
    {{ gofumpt }} -w .
    go fix ./...

# Run everything CI runs
ci: lint build test

# Build the linux/amd64 image with the tag CI uses
docker:
    docker build --platform linux/amd64 -t "{{ image }}" .

# Build the image and push it to ECR
docker-push: docker
    aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin {{ registry }}
    docker push "{{ image }}"

# Start a local ClickHouse server for the queen
local-clickhouse:
    docker run --name ants-clickhouse --rm -p 9000:9000 -p 8123:8123 -e CLICKHOUSE_DB=ants_local -e CLICKHOUSE_USER=ants_local -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -e CLICKHOUSE_PASSWORD=password clickhouse/clickhouse-server
