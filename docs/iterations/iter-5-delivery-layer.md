# Iteration 5: Delivery Layer (HTTP handlers + middleware)

**Branch**: `iter-5-delivery-layer`
**Status**: ✅ Completed

---

## What Was Done

Full HTTP delivery layer using standard `net/http` + `http.ServeMux` (Go 1.22+ patterns).

### Middleware (`internal/delivery/http/middleware/`)

| File | Behaviour |
|------|-----------|
| `logger.go` | zerolog request logger: method, path, status, duration, request-id |
| `recovery.go` | Panic recovery → HTTP 500 JSON |
| `request_id.go` | Generates/propagates `X-Request-Id` UUID header |
| `telegram_auth.go` | Extracts `X-Telegram-User-Id` header → context; exports `TelegramUserID(ctx)` |
| `admin_auth.go` | Validates `X-Admin-Token`; 401 missing, 403 wrong |
| `content_type.go` | Sets `Content-Type: application/json` globally |

### Response Helpers (`internal/delivery/http/respond/respond.go`)
```go
func JSON(w, status, data)    // encode data as JSON
func Error(w, status, code, message) // standard error envelope
func ErrorFrom(w, err)        // maps apperrors sentinels → HTTP codes
```
Mapping: `ErrNotFound` → 404, `IsConflict` → 409, `IsForbidden` → 403, default → 500

### Handlers (`internal/delivery/http/handlers/`)

| File | Routes |
|------|--------|
| `user_handler.go` | `POST /api/v1/users`, `PATCH /users/{id}`, `GET /users/{id}/subscription` |
| `content_handler.go` | `GET /content/modules`, `GET /content/modules/{id}/themes`, `POST /users/{id}/study-sessions`, `POST /users/{id}/test-attempts`, `PUT /users/{id}/test-attempts/{aid}`, `GET /users/{id}/theme/{tid}/access` |
| `progress_handler.go` | `GET /users/{id}/progress`, `GET /users/{id}/progress/modules/{mid}` |
| `subscription_handler.go` | `POST /teachers/{id}/promo-codes`, `GET /teachers/{id}/promo-codes`, `POST /users/{id}/subscriptions` |
| `payment_handler.go` | `POST /users/{id}/payment-invoices`, `GET /users/{id}/payment-invoices/pending`, `POST /webhooks/payment-gateway` |
| `teacher_handler.go` | `GET /teachers/{id}/students`, `GET /teachers/{id}/students/{sid}/progress`, `GET /teachers/{id}/statistics` |
| `admin_handler.go` | 9 admin endpoints (promo codes, content CRUD, users, analytics) |

**Total: 29 endpoints** registered (matching OpenAPI spec).

### Router (`internal/delivery/http/router.go`)
`NewRouter` wires all routes with middleware chains:
- **Public**: `GET /health`, `POST /webhooks/payment-gateway`
- **TelegramAuth**: all user/content/progress/teacher routes
- **AdminAuth**: all `/api/v1/admin/` routes
- Global stack: `RequestID → Logger → Recovery → ContentType`

### Config (`internal/delivery/http/config.go`)
Reads all env vars via `os.Getenv` with defaults:
```
SERVER_ADDR, ADMIN_TOKEN, DB_PATH, LOG_LEVEL, LOG_FORMAT
S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY, S3_REGION
PAYMENT_GATEWAY_URL, PAYMENT_SHOP_ID, PAYMENT_SECRET_KEY, PAYMENT_WEBHOOK_SECRET
```

### Stub Infrastructure (`internal/infrastructure/stub/`)
| File | Behaviour |
|------|-----------|
| `stub_storage.go` | Returns placeholder presigned URL: `https://placeholder/{key}` |
| `stub_payment.go` | Returns fake invoice with UUID invoice_id |
| `stub_notification.go` | Prints to stdout via log |

### Updated `cmd/server/main.go`
Full wiring:
1. Parse config
2. Create zerolog logger
3. `sqlite.Open(ctx, cfg.DBPath)` → auto-migrate
4. Create all repositories
5. Create stub services
6. Create all use cases
7. Create all handlers
8. `NewRouter(...)` with middleware
9. Graceful shutdown on SIGINT/SIGTERM

### Validation
```
go1.26.1 build ./...   → ✅ zero errors
go1.26.1 test ./...    → ✅ all 35 tests pass
GET /health             → ✅ {"status":"ok"}
```

---

## Next: Iteration 6 — CI/CD Pipeline (GitHub Actions)
