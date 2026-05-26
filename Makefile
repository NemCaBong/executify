SHELL := /bin/bash

# Source env vars from .env so DSN-style targets (migrate, seed) work standalone.
# If .env is absent the targets that need it will surface the missing variable.
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

MIGRATIONS_DIR := migrations
DB_URL ?= mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

.PHONY: help build run worker-submit worker-run up down logs migrate-up migrate-down seed test lint tidy fmt clean \
        docker-build docker-server docker-worker-submit docker-worker-run docker-seed docker-shell docker-clean

help: ## Show this help.
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

## --- build / run -----------------------------------------------------------

build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/seed   ./cmd/seed

run:
	go run ./cmd/server

worker-submit:
	go run ./cmd/worker submit

worker-run:
	go run ./cmd/worker run

## --- infra -----------------------------------------------------------------

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

## --- migrations ------------------------------------------------------------
# Requires golang-migrate: `go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

## --- seed ------------------------------------------------------------------

seed:
	go run ./cmd/seed

## --- quality ---------------------------------------------------------------

test:
	go test -race -cover ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

clean:
	rm -rf bin

## --- docker ----------------------------------------------------------------
# All docker-* targets expect MySQL + Redis to be running on the compose
# network — bring them up first with `make up`.

IMAGE          ?= executify:latest
DOCKER_NET     ?= executify_default
# Override .env's 127.0.0.1 so the container reaches the compose services by name.
DOCKER_RUN_ENV  = --env-file .env -e MYSQL_HOST=executify-mysql -e REDIS_ADDRESS=executify-redis:6379

docker-build:
	DOCKER_BUILDKIT=1 docker build -t $(IMAGE) .

docker-server: ## Run the HTTP server (foreground, port 8080)
	docker run --rm -d --name executify-server \
		--network $(DOCKER_NET) $(DOCKER_RUN_ENV) \
		-p 8080:8080 \
		$(IMAGE)

docker-worker-submit: ## Run the submit worker (needs --privileged for isolate)
	docker run --rm -d --name executify-worker-submit \
		--network $(DOCKER_NET) $(DOCKER_RUN_ENV) \
		--privileged \
		--entrypoint /usr/local/bin/executify-worker \
		$(IMAGE) submit

docker-worker-run: ## Run the run-mode worker (needs --privileged for isolate)
	docker run --rm -d --name executify-worker-run \
		--network $(DOCKER_NET) $(DOCKER_RUN_ENV) \
		--privileged \
		--entrypoint /usr/local/bin/executify-worker \
		$(IMAGE) run

docker-seed: ## Run the seeder against MySQL on the compose network
	docker run -d --rm \
		--network $(DOCKER_NET) $(DOCKER_RUN_ENV) \
		--entrypoint /usr/local/bin/executify-seed \
		$(IMAGE)
