# Iteration 4: Use Case Layer

**Branch**: `iter-4-use-case-layer`
**Status**: ✅ Completed
**Tests**: 13 passing (unit tests with hand-written mocks)

---

## What Was Done

Implemented all use case packages in `internal/usecase/`. Each use case depends only on repository/service **interfaces** — never on concrete implementations (Clean Architecture).

### Packages

#### `internal/usecase/user/usecase.go`
| Method | Description |
|--------|-------------|
| `Register` | Create user, return `ErrAlreadyExists` if exists |
| `UpdateRole` | Change user role (student↔teacher) |
| `UpdateSettings` | Update language, notifications_enabled |
| `GetSubscription` | Fetch active subscription for user |

#### `internal/usecase/content/usecase.go`
| Method | Description |
|--------|-------------|
| `GetModules` | Modules enriched with completed_themes count per user |
| `GetModuleThemes` | Themes with access/completion/score per user |
| `CheckThemeAccess` | Subscription OR sequential access logic |
| `CreateStudySession` | Check access → mark started → return mnemonics with presigned S3 URLs |

**Access logic** (implemented correctly):
1. User has active subscription → `accessible=true, access_type="subscription"`
2. No subscription, first theme (no previous) → `accessible=true, access_type="sequential"`
3. No subscription, previous theme completed → `accessible=true, access_type="sequential"`
4. No subscription, previous theme not completed → `accessible=false, reason="previous_theme_required"`

#### `internal/usecase/content/test_usecase.go`
| Method | Description |
|--------|-------------|
| `StartTestAttempt` | Create TestAttempt with UUID, increment current_attempt |
| `SubmitTestAttempt` | Grade answers, update progress, return result + next action |

**Idempotency**: re-submitting an already-submitted `attemptID` returns cached result without re-grading.

**Next action types**: `next_theme`, `retry_test`, `module_completed`, `all_completed`

#### `internal/usecase/progress/usecase.go`
| Method | Description |
|--------|-------------|
| `GetUserProgress` | Overall stats (total/completed modules+themes, avg score, study days) + recent activity |
| `GetModuleProgress` | Per-theme breakdown with status, score, attempt count |

#### `internal/usecase/subscription/usecase.go`
| Method | Description |
|--------|-------------|
| `ActivatePromoCode` | Teacher claims pending promo code (pending→active) |
| `CreatePromoSubscription` | Student joins via promo: validate → consume → link teacher-student → activate user |
| `CreatePaymentSubscription` | Activate subscription after successful payment (idempotent by paymentID) |
| `GetTeacherPromoCodes` | List promo codes for teacher |

#### `internal/usecase/payment/usecase.go`
| Method | Description |
|--------|-------------|
| `CreateInvoice` | Call PaymentService → save pending_payment_id on user → return invoice |
| `HandleWebhook` | Verify HMAC signature → idempotent by paymentID → activate subscription on success |

#### `internal/usecase/teacher/usecase.go`
| Method | Description |
|--------|-------------|
| `GetStudents` | List students with progress summary + last activity |
| `GetStudentProgress` | Verify `IsTeacherStudent` → per-module/theme progress breakdown |
| `GetStatistics` | Group stats: completion rate, avg score, top students, difficult themes |

#### `internal/usecase/admin/usecase.go`
| Method | Description |
|--------|-------------|
| `CreatePromoCode` | Admin creates a new pending promo code |
| `DeactivatePromoCode` | Deactivate via domain method |
| `CreateModule` | Create module |
| `UpdateModule` | Update module |
| `CreateTheme` | Create theme (validates order_num=1 must be introduction) |
| `CreateMnemonic` | Create mnemonic (validates content via `Mnemonic.Validate()`) |
| `CreateTest` | Create test with questions (validates via `Test.Validate()`) |
| `GetUsers` | List users with optional role/subscription filters |

### Result Types (content package)
```go
type ModuleWithProgress     // Module + TotalThemes, CompletedThemes, IsAccessible
type ThemeWithAccess        // Theme + IsAccessible, IsCompleted, Score, LockedReason
type ModuleThemesResult     // ModuleID, ModuleName, Themes
type StudySessionResult     // SessionID(UUID), Theme, Mnemonics, TestAvailable, TestID
type AccessResult           // Accessible, AccessType, Reason, RequiredThemeID/Name/Action
type SubmitResult           // Score, Passed, CorrectAnswers, TotalQuestions, NextAction
type NextAction             // Type, ThemeID, ThemeName, IsIntroduction, Message
```

### Tests

**`content/usecase_test.go`** (7 tests):
- `TestCheckThemeAccess_WithActiveSubscription`
- `TestCheckThemeAccess_NoSubscription_PrevCompleted`
- `TestCheckThemeAccess_NoSubscription_PrevNotCompleted`
- `TestCheckThemeAccess_FirstTheme_AlwaysAccessible`
- `TestSubmitTestAttempt_Passed`
- `TestSubmitTestAttempt_Failed`
- `TestSubmitTestAttempt_IdempotentResubmit`

**`subscription/usecase_test.go`** (6 tests):
- `TestActivatePromoCode_HappyPath`
- `TestActivatePromoCode_PromoNotFound`
- `TestActivatePromoCode_AlreadyActivated`
- `TestActivatePromoCode_NotTeacher`
- `TestCreatePromoSubscription_HappyPath`
- `TestCreatePromoSubscription_AlreadyHasSubscription`

Mocks: hand-written structs implementing interfaces (no mockery tool).

### Validation Results
```
go1.26.1 build ./...   → ✅ zero errors
go1.26.1 test ./...    → ✅ all 35 tests pass

ok  github.com/vladkonst/mnemonics/internal/domain/content       (cached)
ok  github.com/vladkonst/mnemonics/internal/domain/progress      (cached)
ok  github.com/vladkonst/mnemonics/internal/domain/subscription  (cached)
ok  github.com/vladkonst/mnemonics/internal/repository/sqlite    (cached)
ok  github.com/vladkonst/mnemonics/internal/usecase/content      ✅ 7 tests
ok  github.com/vladkonst/mnemonics/internal/usecase/subscription ✅ 6 tests
```

---

## Next: Iteration 5 — Delivery Layer (HTTP handlers + middleware)
