# Iteration 2: Domain Layer

**Branch**: `iter-2-domain-layer`
**Status**: ✅ Completed
**Tests**: 12 passing

---

## What Was Done

Pure business logic layer — **zero external dependencies** (only `pkg/apperrors`).

### Domain Packages

#### `internal/domain/user/`
| File | Content |
|------|---------|
| `user.go` | `User` aggregate root: `IsTeacher()`, `HasActiveSubscription()`, `SetRole()`, `ActivateSubscription()`, `SetPendingPayment()`, `ClearPendingPayment()` |
| `role.go` | `Role` value object: `student` / `teacher`, `NewRole()` with validation |
| `subscription_status.go` | `SubscriptionStatus` value object: `active` / `inactive` / `expired` |

#### `internal/domain/content/`
| File | Content |
|------|---------|
| `module.go` | `Module` aggregate root (ID, Name, OrderNum, IsLocked, IconEmoji) |
| `theme.go` | `Theme` entity (ModuleID FK, IsIntroduction flag, IsLocked, EstimatedTime) |
| `mnemonic.go` | `Mnemonic` entity, `MnemonicType` (text/image), `Validate()` enforces content rules |
| `test.go` | `Test` aggregate root with `[]Question` value objects, `Grade(answers map[int]string)`, `Passed(score int)`, `Validate()` |

#### `internal/domain/progress/`
| File | Content |
|------|---------|
| `score.go` | `Score` value object (0–100), `Passed(passingScore)`, `Grade()` (5/4/3/2 scale) |
| `user_progress.go` | `UserProgress` aggregate: `MarkStarted()`, `StartTest()`, `Complete(score)`, `Fail(score)`, `IsCompleted()` |
| `test_attempt.go` | `TestAttempt` entity with `AttemptID` (UUID, idempotency key), `IsSubmitted()` |

#### `internal/domain/subscription/`
| File | Content |
|------|---------|
| `promo_code.go` | `PromoCode` aggregate root, lifecycle `pending→active→expired/deactivated`: `Activate(teacherID)`, `IsValidForStudent()`, `Consume()`, `Deactivate()` |
| `subscription.go` | `Subscription` entity (personal/university types), `IsActive()` with expiry check |

### Repository & Service Interfaces (`internal/domain/interfaces/`)

#### `repositories.go`
All repository contracts with no implementation details:
- `UserRepository` — Create, GetByID, Update, Exists
- `ModuleRepository` — GetAll, GetByID, Create, Update
- `ThemeRepository` — GetByModuleID, GetByID, Create, GetPreviousTheme
- `MnemonicRepository` — GetByThemeID, Create
- `TestRepository` — GetByThemeID, GetByID, Create
- `ProgressRepository` — Upsert, GetByUserAndTheme, GetByUser, GetByUserAndModule, CountCompletedByUser
- `TestAttemptRepository` — Create, GetByAttemptID, GetByUserAndTheme
- `PromoCodeRepository` — GetByCode, Update, Create, Deactivate, GetByTeacherID
- `SubscriptionRepository` — Create, GetActiveByUserID, GetByPaymentID
- `TeacherStudentRepository` — AddStudent, GetStudentsByTeacher, IsTeacherStudent

#### `services.go`
External service contracts:
- `StorageService` — `PresignURL(ctx, s3Key)` → URL
- `PaymentService` — `CreateInvoice(...)`, `VerifyWebhookSignature(...)`
- `NotificationService` — `Send(ctx, telegramID, message)`

### Tests
| File | Tests |
|------|-------|
| `domain/content/test_test.go` | `TestTest_Grade` (all/half/none correct), `TestTest_Passed` |
| `domain/progress/score_test.go` | `TestNewScore_Valid`, `TestNewScore_Invalid`, `TestScore_Grade` |
| `domain/subscription/promo_code_test.go` | `TestPromoCode_Activate`, `TestPromoCode_Activate_AlreadyActivated`, `TestPromoCode_Consume`, `TestPromoCode_Consume_Exhausted`, `TestPromoCode_Expired` |

```
ok  github.com/vladkonst/mnemonics/internal/domain/content       0.493s
ok  github.com/vladkonst/mnemonics/internal/domain/progress      0.550s
ok  github.com/vladkonst/mnemonics/internal/domain/subscription  0.517s
```

---

## Key Design Decisions

1. **PromoCode is Aggregate Root** (not Value Object) — has identity (code PK), mutable lifecycle, business methods that enforce consistency between `remaining` and `status`
2. **Test.Grade()** takes `map[int]string` (question_id → answer) — decoupled from HTTP layer
3. **Score** is immutable value object — created via `NewScore()` constructor with validation
4. **UserProgress** lifecycle enforced via methods — external code cannot directly mutate status

---

## Next: Iteration 3 — Repository Layer (SQLite)
