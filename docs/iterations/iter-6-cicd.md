# Iteration 6: CI/CD Pipeline (GitHub Actions)

**Branch**: `iter-6-cicd`
**Status**: ✅ Completed

---

## What Was Done

### `.github/workflows/ci.yml` — Continuous Integration

Triggers on:
- Push to `dev` or any `iter-*` branch
- Pull requests targeting `dev` or `main`

Jobs:
1. **build-and-test** (ubuntu-latest, Go 1.26):
   - `go mod download` + `go mod verify`
   - `go build ./...`
   - `go test ./... -race -count=1` (with 120s timeout)
   - Coverage report uploaded as artifact

2. **lint** (ubuntu-latest, Go 1.26):
   - `golangci/golangci-lint-action@v6` with `.golangci.yml` config

### `.github/workflows/cd.yml` — Continuous Deployment

Triggers on: push to `dev`

Steps:
1. Build + test (same as CI)
2. Auto-merge `dev → main` via `git merge dev --no-ff`

Workflow:
```
iter-N branch → PR to dev → CI passes → merge to dev → CD: auto-merge to main
```

### `.golangci.yml` — Linter Config

Enabled linters:
- `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`
- `gofmt`, `goimports` (local prefix: `github.com/vladkonst/mnemonics`)
- `revive`, `misspell`

Test files excluded from `errcheck`.

---

## Branch Strategy (Final)

```
main ← (CD: auto-merge when dev CI passes)
  ↑
dev  ← (merge iter branches via PR)
  ↑
iter-1-project-setup-openapi-migrations  ✅
iter-2-domain-layer                      ✅
iter-3-repository-layer                  ✅
iter-4-use-case-layer                    ✅
iter-5-delivery-layer                    ✅
iter-6-cicd                              ✅
```

---

## Project Complete — Summary

All 6 iterations delivered:

| Layer | Status | Tests |
|-------|--------|-------|
| Go project + OpenAPI + Migrations | ✅ | — |
| Domain (entities, value objects, interfaces) | ✅ | 12 |
| Repository (SQLite, raw SQL) | ✅ | 22 |
| Use Cases (business logic) | ✅ | 13 |
| Delivery (HTTP handlers + middleware) | ✅ | — |
| CI/CD (GitHub Actions) | ✅ | — |

**Total: 47 tests, `go build ./...` clean**
