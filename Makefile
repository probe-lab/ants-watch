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
	go install github.com/volatiletech/sqlboiler/v4@v4.13.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.13.0

models:
	sqlboiler --no-tests psql

migrate-up:
	migrate -database 'postgres://ants_watch:password@localhost:5432/ants_watch?sslmode=disable' -path db/migrations up

migrate-down:
	migrate -database 'postgres://ants_watch:password@localhost:5432/ants_watch?sslmode=disable' -path db/migrations down

.PHONY: build
build:
	echo ${TAG}
	docker build --platform="linux/amd64" -t ${IMAGE_NAME} .

.PHONY: push
push:
	aws ecr get-login-password --profile probelab --region ${REPO_REGION} | docker login --username ${REPO_USER} --password-stdin ${REPO}
	docker tag ${IMAGE_NAME} ${REPO}/${IMAGE_NAME}
	docker push ${REPO}/${IMAGE_NAME}
