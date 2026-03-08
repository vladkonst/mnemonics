.PHONY: build run test lint generate migrate-up migrate-down migrate-create seed tidy

# ── Build ────────────────────────────────────────────────────────────────────
build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

# ── Test ─────────────────────────────────────────────────────────────────────
test:
	go test ./... -count=1 -race

test-cover:
	go test ./... -count=1 -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ── Lint ─────────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ── Code generation ───────────────────────────────────────────────────────────
generate:
	cd api && oapi-codegen -config oapi-codegen.yaml openapi.yaml

# ── Database migrations ───────────────────────────────────────────────────────
DB_PATH ?= ./mnemo_dev.db
MIGRATIONS_DIR = ./database/migrations

migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) status

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=add_some_table"; exit 1; fi
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# ── Seed ─────────────────────────────────────────────────────────────────────
seed:
	sqlite3 $(DB_PATH) < ./database/seeds/dev_seed.sql

# ── Dependencies ─────────────────────────────────────────────────────────────
tidy:
	go mod tidy

# ── Swagger UI ───────────────────────────────────────────────────────────────
swagger:
	docker compose up swagger-ui
