# Iteration 3: Repository Layer (SQLite)

**Branch**: `iter-3-repository-layer`
**Status**: ✅ Completed
**Tests**: 22 passing (integration, in-memory SQLite)

---

## What Was Done

Implemented all 10 SQLite repositories in `internal/repository/sqlite/`, each satisfying the corresponding interface from `internal/domain/interfaces/repositories.go`.

### Repositories

| File | Interface | Methods |
|------|-----------|---------|
| `user_repository.go` | `UserRepository` | Create, GetByID, Update, Exists |
| `module_repository.go` | `ModuleRepository` | Create, GetByID, GetAll, Update |
| `theme_repository.go` | `ThemeRepository` | Create, GetByID, GetByModuleID, GetPreviousTheme |
| `mnemonic_repository.go` | `MnemonicRepository` | Create, GetByThemeID |
| `test_repository.go` | `TestRepository` | Create, GetByID, GetByThemeID |
| `progress_repository.go` | `ProgressRepository` | Upsert, GetByUserAndTheme, GetByUser, GetByUserAndModule, CountCompletedByUser |
| `test_attempt_repository.go` | `TestAttemptRepository` | Create, GetByAttemptID, GetByUserAndTheme |
| `promo_code_repository.go` | `PromoCodeRepository` | Create, GetByCode, Update, Deactivate, GetByTeacherID |
| `subscription_repository.go` | `SubscriptionRepository` | Create, GetActiveByUserID, GetByPaymentID |
| `teacher_student_repository.go` | `TeacherStudentRepository` | AddStudent, GetStudentsByTeacher, IsTeacherStudent |

### Key Implementation Details

**No ORM** — raw `database/sql` throughout.

**Error mapping**:
```go
if errors.Is(err, sql.ErrNoRows) {
    return nil, apperrors.ErrNotFound
}
```

**Boolean handling** — SQLite stores as INTEGER:
```go
// write: boolToInt(value) → 0 or 1
// read:  scanned int != 0
```

**JSON fields**:
- `tests.questions_json` ↔ `[]content.Question` via `encoding/json`
- `test_attempts.answers_json` ↔ `[]progress.AnswerItem` via `encoding/json`

**UserProgress UPSERT** (SQLite ON CONFLICT):
```sql
INSERT INTO user_progress (...) VALUES (...)
ON CONFLICT(user_id, theme_id) DO UPDATE SET
  status = excluded.status,
  score = excluded.score,
  updated_at = CURRENT_TIMESTAMP, ...
```

**TeacherStudentRepository.AddStudent** uses `INSERT OR IGNORE` to be idempotent.

**ThemeRepository.GetPreviousTheme** — JOINs theme with itself to find `order_num = current - 1` in same module:
```sql
SELECT t2.* FROM themes t1
JOIN themes t2 ON t2.module_id = t1.module_id AND t2.order_num = t1.order_num - 1
WHERE t1.id = ?
```

### Integration Tests (`repository_test.go`)

Tests use in-memory SQLite (`:memory:`) — migrations auto-run via `sqlite.Open()`:

| Test | Sub-cases |
|------|-----------|
| `TestUserRepository` | Create+GetByID, not found, Exists, Update |
| `TestModuleRepository` | Create+GetByID, GetAll, not found, Update |
| `TestThemeRepository` | Create+GetByID, GetByModuleID, GetPreviousTheme, first-theme has no prev |
| `TestProgressRepository` | Upsert+Get, not found, GetByUser, CountCompleted |
| `TestPromoCodeRepository` | Create+Get, Activate+Update, GetByTeacherID, Deactivate, not found |

### Validation Results

```
go1.26.1 build ./...   → ✅ zero errors
go1.26.1 test ./...    → ✅ 22 tests pass

ok  github.com/vladkonst/mnemonics/internal/domain/content       0.504s
ok  github.com/vladkonst/mnemonics/internal/domain/progress      0.503s
ok  github.com/vladkonst/mnemonics/internal/domain/subscription  0.503s
ok  github.com/vladkonst/mnemonics/internal/repository/sqlite    1.147s
```

---

## Next: Iteration 4 — Use Case Layer
