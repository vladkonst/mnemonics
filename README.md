# Mnemonics Backend

Backend service for the anatomy exam preparation system. Built with Go using Clean Architecture principles.

## Technology Stack

- **Language**: Go 1.22+
- **HTTP**: `net/http` standard library + `net/http.ServeMux`
- **Database**: SQLite (`modernc.org/sqlite` — pure Go, no CGO)
- **Migrations**: `github.com/pressly/goose/v3`
- **API Spec**: OpenAPI 3.0 (`github.com/oapi-codegen/oapi-codegen/v2`)
- **Config**: `github.com/caarlos0/env/v11` + `github.com/joho/godotenv`
- **Logging**: `github.com/rs/zerolog`

## Architecture

Clean Architecture with strict layer separation:

```
internal/
├── domain/          # Business logic, entities, value objects (no external deps)
├── usecase/         # Business rules orchestration
├── repository/      # Data access (SQLite)
├── delivery/        # HTTP handlers
└── infrastructure/  # External services (S3, Payment)
```

## Getting Started

### Prerequisites

- Go 1.22+
- Docker (for Swagger UI only)

### Run

```bash
cp .env.example .env
make migrate-up
make run
```

### Swagger UI

```bash
docker compose up
# open http://localhost:8081
```

### Development

```bash
make build      # compile
make test       # run tests
make lint       # run golangci-lint
make generate   # regenerate API types from openapi.yaml
make migrate-up # apply DB migrations
```

## Branch Strategy

- `master` — production-ready, protected. Merged from `dev` when CI passes.
- `dev` — integration branch. All feature branches merge here.
- `iter/*` — individual iteration branches.

## API

Full API specification: [`api/openapi.yaml`](api/openapi.yaml)
Interactive docs (Swagger UI): `http://localhost:8081` after `docker compose up`
