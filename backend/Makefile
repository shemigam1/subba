# Load local .env so DATABASE_URL / ADMIN_DATABASE_URL are available to targets.
ifneq (,$(wildcard ./.env))
include .env
export
endif

# Tooling pinned via `go run` so no global installs are needed.
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@latest
SQLC  := go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Migrations run as the privileged role (they create the subba_app role + RLS).
MIGRATE_DSN ?= $(ADMIN_DATABASE_URL)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: up
up: ## Start infra (postgres, redis, rabbitmq)
	docker compose -f ../docker-compose.yml up -d

.PHONY: down
down: ## Stop infra
	docker compose -f ../docker-compose.yml down

.PHONY: migrate
migrate: ## Apply all up migrations
	$(GOOSE) -dir migrations postgres "$(MIGRATE_DSN)" up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	$(GOOSE) -dir migrations postgres "$(MIGRATE_DSN)" down

.PHONY: migrate-status
migrate-status: ## Show migration status
	$(GOOSE) -dir migrations postgres "$(MIGRATE_DSN)" status

.PHONY: sqlc
sqlc: ## Generate typed query code from SQL
	$(SQLC) generate

.PHONY: run
run: ## Run the API server
	go run ./cmd/api

.PHONY: run-worker
run-worker: ## Run the async worker (RabbitMQ consumers + idempotency)
	go run ./cmd/worker

.PHONY: tidy
tidy: ## Sync go.mod/go.sum
	go mod tidy

.PHONY: build
build: ## Compile all binaries
	go build ./...

.PHONY: build-api
build-api: ## Compile only the API binary
	go build -o bin/api ./cmd/api

.PHONY: build-worker
build-worker: ## Compile only the worker binary
	go build -o bin/worker ./cmd/worker
