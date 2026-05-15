SHELL := /bin/bash

# Source env vars from .env so DSN-style targets (migrate, seed) work standalone.
# If .env is absent the targets that need it will surface the missing variable.
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

MIGRATIONS_DIR := migrations
DB_URL ?= mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

.PHONY: help build run worker-submit worker-run up down logs migrate-up migrate-down seed test lint tidy fmt clean

help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

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