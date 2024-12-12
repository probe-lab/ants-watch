SHELL=/usr/bin/env bash

ANTS_REPO?=github.com/probe-lab/ashby
ANTS_COMMIT?=839c2f242c67825ff7884515e929c7ba5b16e84e

TAG?=$(shell date +%F)-$(shell echo -n ${ANTS_REPO}-${ANTS_COMMIT} | sha256sum - | cut -c -8)-$(shell git describe --always --tag --dirty)
IMAGE_NAME?=probelab:ants-${TAG}
REPO?=019120760881.dkr.ecr.us-east-1.amazonaws.com
REPO_USER?=AWS
REPO_REGION?=us-east-1


tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2

non-cluster-migrations:
	mkdir -p db/migrations/local
	for file in $(shell find db/migrations -maxdepth 1 -name "*.sql"); do \
	  filename=$$(basename $$file); \
	  sed 's/Replicated//' $$file > db/migrations/local/$$filename; \
	done

local-migrate-up: non-cluster-migrations
	migrate -database 'clickhouse://localhost:9000?username=ants_local&database=ants_local&password=password&x-multi-statement=true' -path db/migrations/local up

local-migrate-down: non-cluster-migrations
	migrate -database 'clickhouse://localhost:9000?username=ants_local&database=ants_local&password=password&x-multi-statement=true' -path db/migrations/local down

local-clickhouse:
	docker run --name ants-clickhouse --rm -p 9000:9000 -e CLICKHOUSE_DB=ants_local -e CLICKHOUSE_USER=ants_local -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -e CLICKHOUSE_PASSWORD=password clickhouse/clickhouse-server

.PHONY: build
build:
	echo ${TAG}
	docker build --platform="linux/amd64" -t ${IMAGE_NAME} .

.PHONY: push
push:
	aws ecr get-login-password --region ${REPO_REGION} | docker login --username ${REPO_USER} --password-stdin ${REPO}
	docker tag ${IMAGE_NAME} ${REPO}/${IMAGE_NAME}
	docker push ${REPO}/${IMAGE_NAME}
