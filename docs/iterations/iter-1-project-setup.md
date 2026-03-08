# Iteration 1: Go Project Setup, OpenAPI Spec, DB Migrations

**Branch**: `iter-1-project-setup-openapi-migrations`
**Status**: ✅ Completed
**Go version**: 1.26.1

---

## What Was Done

### 1. Project Structure
Created the full Clean Architecture directory layout:
```
backend/
├── api/                    # OpenAPI spec + codegen config
├── cmd/server/             # Entry point
├── internal/
│   ├── api/                # Generated types (oapi-codegen output)
│   ├── domain/             # Business logic (no external deps)
│   ├── usecase/            # Business rules orchestration
│   ├── repository/sqlite/  # Data access
│   ├── delivery/http/      # HTTP handlers + middleware
│   └── infrastructure/     # S3, payment gateway clients
├── database/
│   ├── migrations/         # Goose SQL migrations
│   └── seeds/              # Dev seed data
└── pkg/
    ├── apperrors/          # Typed domain error sentinels
    └── logger/             # Zerolog wrapper
```

### 2. Go Module
- Module: `github.com/vladkonst/mnemonics`
- Go: 1.26.1
- Key dependencies:
  - `modernc.org/sqlite` — pure Go SQLite (no CGO)
  - `github.com/pressly/goose/v3` — SQL migrations
  - `github.com/rs/zerolog` — structured logging
  - `github.com/joho/godotenv` — .env loading
  - `github.com/caarlos0/env/v11` — env config binding
  - `github.com/google/uuid` — UUID generation

### 3. OpenAPI Specification (`api/openapi.yaml`)
Full OpenAPI 3.0 spec covering all **29 endpoints**:
- `POST /api/v1/users` — user registration
- `PATCH /api/v1/users/{user_id}` — update user
- `GET /api/v1/users/{user_id}/subscription` — subscription info
- `GET /api/v1/content/modules` — list modules
- `GET /api/v1/content/modules/{module_id}/themes` — module themes
- `POST /api/v1/users/{user_id}/study-sessions` — start study session
- `POST /api/v1/users/{user_id}/test-attempts` — start test
- `PUT /api/v1/users/{user_id}/test-attempts/{attempt_id}` — submit answers
- `GET /api/v1/users/{user_id}/theme/{theme_id}/access` — check access
- `GET /api/v1/users/{user_id}/progress` — overall progress
- `GET /api/v1/users/{user_id}/progress/modules/{module_id}` — module progress
- `POST /api/v1/teachers/{teacher_id}/promo-codes` — activate promo code
- `GET /api/v1/teachers/{teacher_id}/promo-codes` — list promo codes
- `POST /api/v1/users/{user_id}/subscriptions` — create subscription
- `POST /api/v1/users/{user_id}/payment-invoices` — create invoice
- `GET /api/v1/users/{user_id}/payment-invoices/pending` — pending invoice
- `POST /api/v1/webhooks/payment-gateway` — payment webhook
- `GET /api/v1/teachers/{teacher_id}/students` — teacher's students
- `GET /api/v1/teachers/{teacher_id}/students/{student_id}/progress` — student progress
- `GET /api/v1/teachers/{teacher_id}/statistics` — group stats
- 9× admin endpoints (`/api/v1/admin/...`)

Security schemes: `TelegramAuth` (X-Telegram-User-Id header), `AdminAuth` (X-Admin-Token header)

### 4. Database Migrations (`database/migrations/00001_initial_schema.sql`)
11 tables with goose Up/Down sections:
- `users` — PK: telegram_id, role, subscription_status, pending_payment_id
- `modules` — order_num, is_locked, icon_emoji
- `themes` — module_id FK, is_introduction, is_locked, estimated_time_minutes
- `mnemonics` — type (text/image), content_text, s3_image_key
- `tests` — questions_json (JSONB), passing_score, shuffle flags
- `promo_codes` — lifecycle: pending→active→expired/deactivated
- `teacher_promo_students` — junction table (teacher_id, student_id)
- `subscriptions` — payment_id PK, type (personal/university)
- `user_progress` — composite PK (user_id, theme_id)
- `test_attempts` — attempt_id UUID (idempotency key)
- `notifications`, `audit_log`

12 indexes for query performance.

`database/migrations/migrations.go` — embedded FS for goose (no runtime path needed).

### 5. Dev Seed Data (`database/seeds/dev_seed.sql`)
- 3 modules (Остеология, Миология, Спланхнология)
- 6 themes (4 in module 1, 2 in module 2; each with `is_introduction=1` for first)
- 4 mnemonics (text type)
- 2 tests with questions JSON
- 2 promo codes (TEST2025, DEMO2025)
- 3 demo users

### 6. SQLite Database Opener (`internal/repository/sqlite/db.go`)
- Opens SQLite with WAL mode + foreign keys enabled
- `MaxOpenConns(1)` — SQLite single-writer constraint
- Auto-runs goose migrations on startup via embedded FS

### 7. Supporting Files
- `cmd/server/main.go` — HTTP server, graceful shutdown, `GET /health` endpoint
- `pkg/logger/logger.go` — zerolog with console/JSON format
- `pkg/apperrors/errors.go` — typed sentinel errors (ErrNotFound, ErrForbidden, etc.)
- `docker-compose.yml` — Swagger UI on port 8081
- `Makefile` — build/run/test/lint/migrate/generate targets
- `.env.example` — all env variables documented
- `api/oapi-codegen.yaml` — codegen config for type generation

---

## Acceptance Criteria

- [x] `go build ./...` passes with zero errors
- [x] `GET /health` returns `{"status":"ok"}`
- [x] OpenAPI spec covers all 29 endpoints with request/response schemas
- [x] Migration SQL is valid (Up + Down)
- [x] `go mod tidy` clean

---

## Next: Iteration 2 — Domain Layer
