# План разработки Backend

## Контекст

Проект: система подготовки к экзаменам по анатомии. Архитектура состоит из трёх компонентов (Telegram Bot, Backend Go, GoAdmin Panel). Данный план охватывает исключительно **backend-трек**.

Вся архитектурная документация завершена:
- 29 endpoints задокументированы в `docs/api-endpoints.md`
- Схема БД и улучшения описаны в `database/schema-improvements.md`
- SQL-миграции описаны в `database/migrations.md` (пока только документация, не файлы)
- Clean Architecture структура определена в `architecture.md` и `CLAUDE.md`

Подход: **Spec-First** — OpenAPI spec → кодогенерация типов → схема БД → реализация по слоям Clean Architecture.

---

## Шаг 1: OpenAPI / Swagger спецификация

**Файл**: `backend/api/openapi.yaml`

Формализовать документ `docs/api-endpoints.md` в стандарт OpenAPI 3.0.

### Что включить:

**Info & Servers**
- title, version (1.0.0), description
- servers: `http://localhost:8080` (dev)

**Security Schemes**
- `TelegramAuth` — X-Telegram-User-Id header (для всех user/teacher endpoints)
- `AdminAuth` — X-Admin-Token header (только для /admin/* endpoints)

**Components / Schemas** — переиспользуемые модели:
- `User`, `Module`, `Theme`, `Mnemonic`, `Test`, `Question`
- `Subscription`, `PromoCode`, `PaymentInvoice`, `TestAttempt`
- `UserProgress`, `ModuleProgress`, `ThemeProgress`
- `GroupStatistics`, `StudentStatistics`
- `Error` (code, message, details)

**Все 29 endpoints** с полными схемами запросов/ответов и кодами ошибок:
- 3 endpoints: Управление пользователями
- 6 endpoints: Контент (модули, темы, сессии, тесты)
- 2 endpoints: Прогресс пользователей
- 3 endpoints: Подписки и промокоды
- 3 endpoints: Платежи
- 3 endpoints: Функционал преподавателя
- 9 endpoints: Администрирование

### Кодогенерация из спецификации

Использовать `github.com/oapi-codegen/oapi-codegen` для генерации Go-типов и server stubs:

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen -generate types,server -package api -o backend/internal/api/generated.go backend/api/openapi.yaml
```

Конфигурационный файл `backend/api/oapi-codegen.yaml`:
```yaml
package: api
generate:
  - types
  - server          # net/http compatible stubs (не Gin/Chi)
output: internal/api/generated.go
```

Генерируемые артефакты:
- `internal/api/generated.go` — все типы запросов/ответов и интерфейс `StrictServerInterface`
- Переименовывать/изменять нельзя — файл перегенерируется при изменении spec

### Swagger UI

`docker-compose.yml` — **только один сервис**: `swaggerapi/swagger-ui` с монтированием `backend/api/openapi.yaml`:
```yaml
services:
  swagger-ui:
    image: swaggerapi/swagger-ui
    ports:
      - "8081:8080"
    volumes:
      - ./api/openapi.yaml:/usr/share/nginx/html/openapi.yaml
    environment:
      SWAGGER_JSON: /usr/share/nginx/html/openapi.yaml
```

**Критерий готовности**: spec валидируется (spectral или redocly), `oapi-codegen` генерирует типы без ошибок, Swagger UI открывается на `http://localhost:8081`.

---

## Шаг 2: Схема базы данных и миграция

**Инструмент**: `github.com/pressly/goose/v3` (формат `NNNNNN_name.sql` с `-- +goose Up` / `-- +goose Down` секциями)

**Директория**: `backend/database/migrations/`

### Одна начальная миграция (всё в одном файле)

`00001_initial_schema.sql` — полная схема со всеми таблицами, полями из `database/schema-improvements.md`, constraints и индексами:

```sql
-- +goose Up

CREATE TABLE users (
    telegram_id   BIGINT PRIMARY KEY,
    role          TEXT NOT NULL DEFAULT 'student' CHECK(role IN ('student','teacher')),
    subscription_status TEXT NOT NULL DEFAULT 'inactive' CHECK(subscription_status IN ('active','inactive','expired')),
    university_code TEXT,
    pending_payment_id TEXT,
    first_name    TEXT NOT NULL,
    last_name     TEXT,
    username      TEXT,
    language      TEXT NOT NULL DEFAULT 'ru',
    timezone      TEXT NOT NULL DEFAULT 'UTC',
    notifications_enabled INTEGER NOT NULL DEFAULT 1,
    last_activity_at DATETIME,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE modules (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT,
    order_num   INTEGER NOT NULL,
    is_locked   INTEGER NOT NULL DEFAULT 0,
    icon_emoji  TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE themes (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    module_id             INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    name                  TEXT NOT NULL,
    description           TEXT,
    order_num             INTEGER NOT NULL,
    is_introduction       INTEGER NOT NULL DEFAULT 0,
    is_locked             INTEGER NOT NULL DEFAULT 0,
    estimated_time_minutes INTEGER,
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mnemonics (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    theme_id      INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    type          TEXT NOT NULL CHECK(type IN ('text','image')),
    content_text  TEXT,
    s3_image_key  TEXT,
    order_num     INTEGER NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tests (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    theme_id        INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    questions_json  TEXT NOT NULL,
    difficulty      INTEGER NOT NULL DEFAULT 1,
    passing_score   INTEGER NOT NULL DEFAULT 70 CHECK(passing_score >= 0 AND passing_score <= 100),
    shuffle_questions INTEGER NOT NULL DEFAULT 1,
    shuffle_answers   INTEGER NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE promo_codes (
    code                TEXT PRIMARY KEY,
    university_name     TEXT NOT NULL,
    teacher_id          BIGINT REFERENCES users(telegram_id) ON DELETE SET NULL,
    max_activations     INTEGER NOT NULL CHECK(max_activations > 0),
    remaining           INTEGER NOT NULL CHECK(remaining >= 0),
    status              TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','active','expired','deactivated')),
    expires_at          DATETIME,
    created_by_admin_id BIGINT,
    activated_at        DATETIME,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE teacher_promo_students (
    teacher_id  BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    student_id  BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    promo_code  TEXT REFERENCES promo_codes(code),
    joined_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (teacher_id, student_id)
);

CREATE TABLE subscriptions (
    payment_id          TEXT PRIMARY KEY,
    user_id             BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    type                TEXT NOT NULL CHECK(type IN ('personal','university')),
    status              TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','expired','cancelled')),
    plan                TEXT,
    expires_at          DATETIME,
    auto_renew          INTEGER NOT NULL DEFAULT 0,
    cancelled_at        DATETIME,
    cancellation_reason TEXT,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(payment_id)
);

CREATE TABLE user_progress (
    user_id         BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    theme_id        INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'started' CHECK(status IN ('started','completed','failed')),
    score           INTEGER CHECK(score >= 0 AND score <= 100),
    current_attempt INTEGER NOT NULL DEFAULT 0 CHECK(current_attempt >= 0),
    test_started_at DATETIME,
    started_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at    DATETIME,
    time_spent_seconds INTEGER NOT NULL DEFAULT 0,
    last_viewed_at  DATETIME,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, theme_id)
);

CREATE TABLE test_attempts (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id          BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    theme_id         INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    test_id          INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    answers_json     TEXT NOT NULL,
    score            INTEGER NOT NULL,
    passed           INTEGER NOT NULL,
    started_at       DATETIME NOT NULL,
    submitted_at     DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL
);

CREATE TABLE notifications (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    type          TEXT NOT NULL,
    title         TEXT NOT NULL,
    message       TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','sent','failed')),
    sent_at       DATETIME,
    error_message TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_log (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id      BIGINT NOT NULL,
    admin_username TEXT,
    action        TEXT NOT NULL,
    entity_type   TEXT NOT NULL,
    entity_id     TEXT NOT NULL,
    old_value_json TEXT,
    new_value_json TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_themes_module_order ON themes(module_id, order_num);
CREATE INDEX idx_themes_introduction ON themes(is_introduction) WHERE is_introduction = 1;
CREATE INDEX idx_user_progress_user ON user_progress(user_id, completed_at DESC);
CREATE INDEX idx_user_progress_theme ON user_progress(theme_id, status);
CREATE INDEX idx_test_attempts_user_theme ON test_attempts(user_id, theme_id, submitted_at DESC);
CREATE INDEX idx_promo_codes_teacher ON promo_codes(teacher_id);
CREATE INDEX idx_promo_codes_status ON promo_codes(status);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status, expires_at);
CREATE INDEX idx_teacher_promo_student ON teacher_promo_students(student_id);
CREATE INDEX idx_mnemonics_theme_order ON mnemonics(theme_id, order_num);
CREATE INDEX idx_users_pending_payment ON users(pending_payment_id) WHERE pending_payment_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS test_attempts;
DROP TABLE IF EXISTS user_progress;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS teacher_promo_students;
DROP TABLE IF EXISTS promo_codes;
DROP TABLE IF EXISTS tests;
DROP TABLE IF EXISTS mnemonics;
DROP TABLE IF EXISTS themes;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS users;
```

Последующие миграции добавляются новыми файлами `00002_*.sql`, `00003_*.sql` по мере итерации разработки.

### Seed-данные для разработки

`backend/database/seeds/dev_seed.sql`:
- 3 модуля, 3 темы в первом модуле (is_introduction=1 для первой)
- Мнемоники текстовые + заглушка s3_image_key
- Тест с 5 вопросами
- 2 промокода (TEST2025, DEMO2025)

**Критерий готовности**: `goose up` проходит без ошибок, `goose down` откатывает чисто.

---

## Шаг 3: Инициализация Go-проекта

**Директория**: `backend/`

### 3.1 Структура директорий (Clean Architecture)

```
backend/
├── api/
│   ├── openapi.yaml               # OpenAPI spec (Шаг 1)
│   └── oapi-codegen.yaml          # Конфиг кодогенерации
├── cmd/
│   └── server/
│       └── main.go                # Точка входа
├── internal/
│   ├── api/
│   │   └── generated.go           # Авто-генерация из openapi.yaml (не трогать вручную)
│   ├── domain/                    # Слой: Rich domain model
│   │   ├── user/
│   │   │   ├── user.go            # Aggregate Root
│   │   │   ├── role.go            # Value Object
│   │   │   └── subscription_status.go  # Value Object
│   │   ├── content/
│   │   │   ├── module.go          # Aggregate Root
│   │   │   ├── theme.go           # Entity
│   │   │   ├── mnemonic.go        # Entity
│   │   │   └── test.go            # Aggregate Root (с Question Value Objects)
│   │   ├── progress/
│   │   │   ├── user_progress.go   # Aggregate Root
│   │   │   ├── score.go           # Value Object
│   │   │   └── test_attempt.go    # Entity
│   │   ├── subscription/
│   │   │   ├── promo_code.go      # Aggregate Root
│   │   │   └── subscription.go    # Entity
│   │   └── interfaces/
│   │       ├── repositories.go    # Интерфейсы репозиториев
│   │       └── services.go        # Интерфейсы внешних сервисов
│   ├── usecase/                   # Слой: бизнес-логика (оркестрация)
│   │   ├── user/
│   │   ├── content/
│   │   ├── progress/
│   │   ├── subscription/
│   │   ├── payment/
│   │   ├── teacher/
│   │   └── admin/
│   ├── repository/                # Слой: доступ к данным (raw SQL)
│   │   └── sqlite/
│   │       ├── db.go              # *sql.DB + Goose миграции
│   │       ├── user_repository.go
│   │       ├── module_repository.go
│   │       └── ...
│   ├── delivery/                  # Слой: HTTP-обработчики
│   │   └── http/
│   │       ├── middleware/
│   │       │   ├── logger.go
│   │       │   ├── recovery.go
│   │       │   ├── request_id.go
│   │       │   ├── telegram_auth.go
│   │       │   ├── admin_auth.go
│   │       │   ├── resource_ownership.go
│   │       │   ├── content_type.go
│   │       │   ├── rate_limit.go
│   │       │   └── cors.go
│   │       ├── handlers/
│   │       │   ├── user_handler.go
│   │       │   ├── content_handler.go
│   │       │   ├── progress_handler.go
│   │       │   ├── subscription_handler.go
│   │       │   ├── payment_handler.go
│   │       │   ├── teacher_handler.go
│   │       │   └── admin_handler.go
│   │       ├── server.go          # Реализует сгенерированный StrictServerInterface
│   │       └── router.go          # http.ServeMux routing
│   └── infrastructure/            # Внешние зависимости
│       ├── s3/
│       │   └── client.go          # AWS S3 (aws-sdk-go-v2)
│       └── payment/
│           └── client.go          # YooKassa или Stripe
├── database/
│   ├── migrations/
│   │   └── 00001_initial_schema.sql
│   └── seeds/
│       └── dev_seed.sql
├── pkg/
│   ├── apperrors/                 # Типизированные ошибки домена
│   └── logger/                    # zerolog обёртка
├── .env.example
├── docker-compose.yml             # Только Swagger UI
├── Makefile
└── go.mod
```

### 3.2 Выбор библиотек

| Назначение | Библиотека |
|-----------|-----------|
| HTTP | `net/http` (стандартная библиотека) |
| Router | `net/http.ServeMux` (Go 1.22+, поддерживает паттерны `GET /api/v1/users/{id}`) |
| DB Driver | `modernc.org/sqlite` (pure Go, без CGO) |
| Migrations | `github.com/pressly/goose/v3` |
| API Codegen | `github.com/oapi-codegen/oapi-codegen/v2` |
| Env Config | `github.com/caarlos0/env/v11` |
| .env файл | `github.com/joho/godotenv` |
| Logging | `github.com/rs/zerolog` |
| AWS S3 | `github.com/aws/aws-sdk-go-v2` |
| Testing | `github.com/stretchr/testify` |
| Mocks | `github.com/vektra/mockery/v2` |
| UUID | `github.com/google/uuid` |
| Validation | `github.com/go-playground/validator/v10` |

**Критерий готовности**: `go build ./...` проходит, Swagger UI доступен через `docker compose up`.

---

## Шаг 4: Domain Layer (Rich Domain Model)

**Файлы**: `backend/internal/domain/`

Domain-слой содержит **бизнес-логику** непосредственно в сущностях — не просто data transfer objects. Никаких зависимостей от внешних библиотек.

### 4.1 Value Objects

Неизменяемые объекты, определяемые своим значением:

```go
// domain/user/role.go
type Role string
const (
    RoleStudent Role = "student"
    RoleTeacher Role = "teacher"
)
func NewRole(s string) (Role, error) { ... } // валидация

// domain/progress/score.go
type Score struct{ value int }
func NewScore(v int) (Score, error) { // 0-100
    if v < 0 || v > 100 { return Score{}, ErrInvalidScore }
    return Score{value: v}, nil
}
func (s Score) Passed(passingScore int) bool { return s.value >= passingScore }
func (s Score) Grade() string { // "5", "4", "3", "2"
    switch {
    case s.value >= 90: return "5"
    case s.value >= 75: return "4"
    case s.value >= 60: return "3"
    default:            return "2"
    }
}

// domain/subscription/promo_code_status.go
type PromoCodeStatus string
const (
    PromoCodeStatusPending     PromoCodeStatus = "pending"
    PromoCodeStatusActive      PromoCodeStatus = "active"
    PromoCodeStatusExpired     PromoCodeStatus = "expired"
    PromoCodeStatusDeactivated PromoCodeStatus = "deactivated"
)
```

### 4.2 Aggregate Roots с бизнес-методами

**Почему PromoCode — Aggregate Root, а не Value Object?**

Value Object неизменяем и определяется своим значением без идентичности (например, `Score` — просто число 0–100, без PK, без lifecycle). PromoCode — Aggregate Root, потому что:
- Имеет собственную идентичность (`code` — PRIMARY KEY в БД)
- Имеет изменяемое состояние с lifecycle: `pending → active → expired/deactivated`
- Содержит бизнес-методы, мутирующие состояние (`Activate`, `Consume`)
- Является точкой консистентности: изменение `remaining` и `status` всегда через методы агрегата

```go
// domain/subscription/promo_code.go
type PromoCode struct {
    Code           string
    UniversityName string
    TeacherID      *int64
    MaxActivations int
    Remaining      int
    Status         PromoCodeStatus
    ExpiresAt      *time.Time
}

// Бизнес-методы на агрегате
func (p *PromoCode) Activate(teacherID int64) error {
    if p.TeacherID != nil { return ErrAlreadyActivated }
    if p.Status != PromoCodeStatusPending { return ErrInvalidStatus }
    p.TeacherID = &teacherID
    p.Status = PromoCodeStatusActive
    return nil
}

func (p *PromoCode) IsValidForStudent() error {
    if p.Status != PromoCodeStatusActive { return ErrPromoCodeNotActive }
    if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) { return ErrPromoCodeExpired }
    if p.Remaining <= 0 { return ErrPromoCodeExhausted }
    return nil
}

func (p *PromoCode) Consume() error {
    if err := p.IsValidForStudent(); err != nil { return err }
    p.Remaining--
    return nil
}

// domain/progress/user_progress.go
type UserProgress struct {
    UserID         int64
    ThemeID        int
    Status         ProgressStatus
    Score          *Score
    CurrentAttempt int
    TestStartedAt  *time.Time
    StartedAt      time.Time
    CompletedAt    *time.Time
}

func (up *UserProgress) RecordTestResult(score Score, passed bool) {
    up.Score = &score
    up.CurrentAttempt++
    if passed {
        up.Status = ProgressStatusCompleted
        now := time.Now()
        up.CompletedAt = &now
    } else {
        up.Status = ProgressStatusFailed
    }
}

func (up *UserProgress) StartTest() {
    now := time.Now()
    up.TestStartedAt = &now
    up.CurrentAttempt++
}

// domain/content/test.go — Aggregate Root
type Test struct {
    ID            int
    ThemeID       int
    Questions     []Question   // Value Objects внутри агрегата
    Difficulty    int
    PassingScore  int
    ShuffleQ      bool
    ShuffleA      bool
}

func (t *Test) Evaluate(answers map[int]string) (Score, bool) {
    correct := 0
    for _, q := range t.Questions {
        if answers[q.ID] == q.CorrectAnswer { correct++ }
    }
    pct := correct * 100 / len(t.Questions)
    score, _ := NewScore(pct)
    return score, score.Passed(t.PassingScore)
}

// Question — Value Object внутри Test Aggregate
type Question struct {
    ID            int
    Text          string
    Type          QuestionType
    Options       []string
    CorrectAnswer string
    OrderNum      int
}
```

### 4.3 Repository Interfaces (`domain/interfaces/repositories.go`)

```go
type UserRepository interface {
    Create(ctx context.Context, user *user.User) error
    GetByTelegramID(ctx context.Context, id int64) (*user.User, error)
    Update(ctx context.Context, user *user.User) error
}

type ModuleRepository interface {
    GetAll(ctx context.Context) ([]*content.Module, error)
    GetByID(ctx context.Context, id int) (*content.Module, error)
    Create(ctx context.Context, m *content.Module) (*content.Module, error)
    Update(ctx context.Context, m *content.Module) error
}

type ThemeRepository interface {
    GetByModuleID(ctx context.Context, moduleID int) ([]*content.Theme, error)
    GetByID(ctx context.Context, id int) (*content.Theme, error)
    GetPrevious(ctx context.Context, moduleID, orderNum int) (*content.Theme, error)
    Create(ctx context.Context, t *content.Theme) (*content.Theme, error)
}

// ... MnemonicRepository, TestRepository, PromoCodeRepository,
//     SubscriptionRepository, UserProgressRepository,
//     TestAttemptRepository, TeacherStudentRepository
```

### 4.4 External Service Interfaces (`domain/interfaces/services.go`)

```go
type S3Service interface {
    GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
    UploadFile(ctx context.Context, key string, data io.Reader, contentType string) error
}

type PaymentService interface {
    CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error)
    GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
    VerifyWebhookSignature(payload []byte, signature string) bool
}

type NotificationService interface {
    SendMessage(ctx context.Context, telegramID int64, message string) error
}
```

**Критерий готовности**: `go build ./internal/domain/...` без единого внешнего импорта.

---

## Шаг 5: Repository Layer (Raw SQL)

**Файлы**: `backend/internal/repository/sqlite/*.go`

Реализовать интерфейсы через `database/sql` с raw SQL запросами. Никакого ORM.

### 5.1 DB Connection + Migrations (`repository/sqlite/db.go`)

```go
func NewSQLiteDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dsn)
    // WAL mode для конкурентных чтений
    db.Exec("PRAGMA journal_mode=WAL")
    db.Exec("PRAGMA foreign_keys=ON")
    // Применить goose миграции
    goose.SetDialect("sqlite3")
    goose.Up(db, "database/migrations")
    return db, err
}
```

### 5.2 Репозитории (raw SQL)

```go
// repository/sqlite/user_repository.go
func (r *UserRepository) GetByTelegramID(ctx context.Context, id int64) (*user.User, error) {
    row := r.db.QueryRowContext(ctx,
        `SELECT telegram_id, role, subscription_status, university_code,
                pending_payment_id, first_name, last_name, username,
                language, notifications_enabled, last_activity_at, created_at
         FROM users WHERE telegram_id = ?`, id)
    return scanUser(row)
}

// repository/sqlite/user_progress_repository.go
func (r *UserProgressRepo) GetAggregatedByUser(ctx context.Context, userID int64) (*progress.Summary, error) {
    // Сложный запрос: агрегация прогресса по всем модулям и темам
    rows, _ := r.db.QueryContext(ctx, `
        SELECT m.id, m.name, t.id, t.name, t.is_introduction,
               up.status, up.score, up.current_attempt, up.completed_at
        FROM modules m
        JOIN themes t ON t.module_id = m.id
        LEFT JOIN user_progress up ON up.theme_id = t.id AND up.user_id = ?
        ORDER BY m.order_num, t.order_num`, userID)
    // ...scan и сборка агрегата...
}
```

Ключевые запросы для каждого репозитория (не исчерпывающий список):
- `UserProgressRepository.GetByUserAndTheme` — WHERE user_id=? AND theme_id=?
- `TeacherStudentRepository.GetStudentsWithProgress` — JOIN teacher_promo_students + user_progress с агрегацией
- `PromoCodeRepository.FindActive` — WHERE code=? AND status='active' AND (expires_at IS NULL OR expires_at > datetime('now'))
- `SubscriptionRepository.FindActiveByUser` — WHERE user_id=? AND status='active' AND (expires_at IS NULL OR expires_at > datetime('now'))

### 5.3 Тесты репозиториев

`backend/internal/repository/sqlite/*_test.go` — интеграционные тесты с SQLite in-memory + goose:
```go
func setupTestDB(t *testing.T) *sql.DB {
    db, _ := sql.Open("sqlite", ":memory:")
    db.Exec("PRAGMA foreign_keys=ON")
    goose.Up(db, "../../database/migrations")
    return db
}
```

**Критерий готовности**: интеграционные тесты проходят, coverage > 80%.

---

## Шаг 6: Use Case Layer

**Файлы**: `backend/internal/usecase/*/`

Оркестрирует domain-объекты и репозитории. Не содержит бизнес-правил — они в domain.

### 6.1 Модули use cases

| Модуль | Use Cases |
|--------|----------|
| `usecase/user/` | RegisterUser, UpdateUser, GetSubscription |
| `usecase/content/` | GetModules, GetThemesByModule, CreateStudySession, StartTest, SubmitTest, CheckThemeAccess |
| `usecase/progress/` | GetUserProgress, GetModuleProgress |
| `usecase/subscription/` | ActivatePromoCode, CreateSubscription, GetTeacherPromoCodes |
| `usecase/payment/` | CreatePaymentInvoice, HandlePaymentWebhook, GetPendingInvoice |
| `usecase/teacher/` | GetStudents, GetStudentProgress, GetGroupStatistics |
| `usecase/admin/` | CreatePromoCode, DeactivatePromoCode, CreateModule, CreateTheme, CreateMnemonic, CreateTest, GetUsers, GetAnalytics |

### 6.2 Пример use case

```go
// usecase/content/submit_test.go
type SubmitTestUseCase struct {
    progressRepo interfaces.UserProgressRepository
    attemptRepo  interfaces.TestAttemptRepository
    testRepo     interfaces.TestRepository
    themeRepo    interfaces.ThemeRepository
    notify       interfaces.NotificationService
}

func (uc *SubmitTestUseCase) Execute(ctx context.Context, in SubmitTestInput) (*SubmitTestOutput, error) {
    // Загрузить тест (агрегат с бизнес-методом Evaluate)
    t, err := uc.testRepo.GetByThemeID(ctx, in.ThemeID)

    // Бизнес-логика Evaluate — в domain
    score, passed := t.Evaluate(in.Answers)

    // Загрузить/создать прогресс
    prog, _ := uc.progressRepo.GetOrCreate(ctx, in.UserID, in.ThemeID)

    // Бизнес-метод на агрегате
    prog.RecordTestResult(score, passed)

    // Сохранить
    uc.progressRepo.Update(ctx, prog)
    uc.attemptRepo.Create(ctx, &progress.TestAttempt{...})

    // Определить следующую тему (если нет подписки)
    nextTheme := uc.resolveNextTheme(ctx, in.UserID, in.ThemeID, passed)

    // Уведомление
    uc.notify.SendMessage(ctx, in.UserID, buildMotivationMessage(score, passed))

    return &SubmitTestOutput{Score: score, Passed: passed, NextTheme: nextTheme}, nil
}
```

### 6.3 Ключевая бизнес-логика оркестрации

- **CheckThemeAccess**: проверить активную подписку → если есть, доступ открыт; если нет → проверить user_progress предыдущей темы (status='completed')
- **HandlePaymentWebhook**: идемпотентность — сначала SELECT по payment_id; всегда возвращать 200 OK payment gateway
- **ActivatePromoCode**: вызвать `promoCode.Activate(teacherID)` (бизнес-метод на агрегате)
- **SubmitTest**: вызвать `test.Evaluate(answers)` и `progress.RecordTestResult(score, passed)`

### 6.4 Тесты use cases

`backend/internal/usecase/*/*_test.go` — unit-тесты с моками (mockery):
```go
progressRepo := mocks.NewUserProgressRepository(t)
progressRepo.On("GetOrCreate", mock.Anything, int64(1), 5).Return(&progress.UserProgress{...}, nil)
uc := content.NewSubmitTestUseCase(progressRepo, ...)
```

**Критерий готовности**: unit-тесты всех use cases, coverage > 80%.

---

## Шаг 7: Delivery Layer (net/http)

**Файлы**: `backend/internal/delivery/http/`

Используется **только стандартная библиотека** `net/http` + `http.ServeMux` (Go 1.22+).

### 7.1 Middleware

Middleware реализуется как `func(http.Handler) http.Handler` и выстраивается в цепочку.

| Middleware | Назначение |
|-----------|-----------|
| `RequestID` | Генерирует X-Request-Id (uuid) для каждого запроса, добавляет в контекст и response header |
| `Logger` | zerolog: логирует method, path, status, duration, request_id |
| `Recovery` | Перехватывает панику → 500 + лог stacktrace |
| `ContentType` | Проверяет Content-Type: application/json для POST/PUT/PATCH; выставляет для ответов |
| `CORS` | Access-Control заголовки (настраиваемый список origins) |
| `RateLimit` | In-memory rate limit по IP (token bucket); усиленный для `/webhooks/*` |
| `TelegramAuth` | Извлекает X-Telegram-User-Id, валидирует (число > 0), кладёт userID в ctx |
| `AdminAuth` | Проверяет X-Admin-Token против значения из env; 401 если не совпадает |
| `ResourceOwnership` | Проверяет, что `{user_id}` в URL совпадает с userID из контекста (TelegramAuth); исключения: admin |
| `TeacherOnly` | Проверяет роль пользователя из БД — должна быть `teacher`; 403 если нет |

Применение middleware по группам:

```go
// router.go
func NewRouter(h *Handlers, env *Config) http.Handler {
    mux := http.NewServeMux()

    // Глобальные: RequestID → Logger → Recovery → ContentType → CORS
    global := chain(mux, RequestID, Logger, Recovery, ContentType, CORS)

    // User endpoints: + TelegramAuth + ResourceOwnership
    userMW := chain(TelegramAuth, ResourceOwnership)

    // Teacher endpoints: + TelegramAuth + TeacherOnly
    teacherMW := chain(TelegramAuth, TeacherOnly)

    // Admin endpoints: + AdminAuth
    adminMW := chain(AdminAuth)

    // Webhook: + RateLimit (строгий)
    webhookMW := chain(RateLimit(10, time.Minute))

    // Регистрация маршрутов (Go 1.22 pattern matching)
    mux.Handle("POST /api/v1/users", userMW(h.RegisterUser))
    mux.Handle("PATCH /api/v1/users/{telegram_id}", userMW(h.UpdateUser))
    mux.Handle("GET /api/v1/users/{user_id}/subscription", userMW(h.GetSubscription))
    mux.Handle("GET /api/v1/content/modules", userMW(h.GetModules))
    mux.Handle("GET /api/v1/content/modules/{module_id}/themes", userMW(h.GetThemes))
    mux.Handle("POST /api/v1/users/{user_id}/study-sessions", userMW(h.CreateStudySession))
    mux.Handle("POST /api/v1/users/{user_id}/test-attempts", userMW(h.StartTest))
    mux.Handle("PUT /api/v1/users/{user_id}/test-attempts/{attempt_id}", userMW(h.SubmitTest))
    mux.Handle("GET /api/v1/users/{user_id}/theme/{theme_id}/access", userMW(h.CheckAccess))
    mux.Handle("GET /api/v1/users/{user_id}/progress", userMW(h.GetProgress))
    mux.Handle("GET /api/v1/users/{user_id}/progress/modules/{module_id}", userMW(h.GetModuleProgress))
    mux.Handle("POST /api/v1/users/{user_id}/subscriptions", userMW(h.CreateSubscription))
    mux.Handle("POST /api/v1/users/{user_id}/payment-invoices", userMW(h.CreatePaymentInvoice))
    mux.Handle("GET /api/v1/users/{user_id}/payment-invoices/pending", userMW(h.GetPendingInvoice))
    mux.Handle("POST /api/v1/teachers/{teacher_id}/promo-codes", teacherMW(h.ActivatePromoCode))
    mux.Handle("GET /api/v1/teachers/{teacher_id}/promo-codes", teacherMW(h.GetPromoCodes))
    mux.Handle("GET /api/v1/teachers/{teacher_id}/students", teacherMW(h.GetStudents))
    mux.Handle("GET /api/v1/teachers/{teacher_id}/students/{student_id}/progress", teacherMW(h.GetStudentProgress))
    mux.Handle("GET /api/v1/teachers/{teacher_id}/statistics", teacherMW(h.GetGroupStatistics))
    mux.Handle("POST /api/v1/admin/promo-codes", adminMW(h.AdminCreatePromoCode))
    // ... остальные admin endpoints
    mux.Handle("POST /api/v1/webhooks/payment-gateway", webhookMW(h.HandlePaymentWebhook))

    return global
}
```

### 7.2 Handler structure

Каждый handler: decode request → вызвать use case → encode response.

```go
// handlers/user_handler.go
func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
    var req api.RegisterUserRequest        // тип из generated.go
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
        return
    }
    if err := validate(req); err != nil {
        writeError(w, http.StatusBadRequest, "validation_error", err.Error())
        return
    }
    result, err := h.userUC.RegisterUser(r.Context(), usecase.RegisterUserInput{...})
    if err != nil {
        writeAppError(w, err)
        return
    }
    writeJSON(w, http.StatusCreated, result)
}
```

### 7.3 Обработка ошибок

```go
// pkg/apperrors/errors.go
type AppError struct {
    Code       int    // HTTP status
    ErrCode    string // machine-readable: "not_found", "forbidden", "conflict"
    Message    string // human-readable для пользователя
}

var (
    ErrNotFound   = &AppError{Code: 404, ErrCode: "not_found"}
    ErrForbidden  = &AppError{Code: 403, ErrCode: "forbidden"}
    ErrConflict   = &AppError{Code: 409, ErrCode: "conflict"}
    ErrBadRequest = &AppError{Code: 400, ErrCode: "bad_request"}
    ErrInternal   = &AppError{Code: 500, ErrCode: "internal_error"}
)
```

**Критерий готовности**: все 29 endpoints зарегистрированы, `go test ./internal/delivery/...` проходит.

---

## Шаг 8: Infrastructure Layer

**Файлы**: `backend/internal/infrastructure/`

### 8.1 S3 Client (`infrastructure/s3/client.go`)

Реализует `domain/interfaces/S3Service` через AWS SDK v2:
- `GeneratePresignedURL` — presigned GET URL для изображений мнемоник (TTL: 1 час)
- `UploadFile` — PUT объекта в bucket
- Конфигурируется через env: `AWS_REGION`, `AWS_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`

### 8.2 Payment Client (`infrastructure/payment/client.go`)

Реализует `domain/interfaces/PaymentService`:
- Провайдер: YooKassa (приоритет)
- `CreateInvoice` — создание счёта через YooKassa API
- `GetPaymentStatus` — проверка статуса платежа
- `VerifyWebhookSignature` — проверка HMAC SHA256 подписи

### 8.3 Notification Client (`infrastructure/telegram/client.go`)

Реализует `domain/interfaces/NotificationService`:
- Отправка сообщений через Telegram Bot API
- Используется только для уведомлений из use cases (не для routing)

### 8.4 Config через env (`cmd/server/main.go`)

Переменные окружения парсятся через `caarlos0/env`. Для удобства локальной разработки поддерживается загрузка `.env` файла через `joho/godotenv`. В production `.env` файла нет — используются настоящие env vars (Docker, systemd, k8s).

```go
import (
    "github.com/caarlos0/env/v11"
    "github.com/joho/godotenv"
)

type Config struct {
    Server struct {
        Port int    `env:"PORT" envDefault:"8080"`
        Host string `env:"HOST" envDefault:"0.0.0.0"`
    }
    Database struct {
        DSN string `env:"DATABASE_DSN,required"`
    }
    S3 struct {
        Region    string `env:"AWS_REGION,required"`
        Bucket    string `env:"AWS_BUCKET,required"`
        AccessKey string `env:"AWS_ACCESS_KEY_ID,required"`
        SecretKey string `env:"AWS_SECRET_ACCESS_KEY,required"`
    }
    Payment struct {
        Provider string `env:"PAYMENT_PROVIDER" envDefault:"yookassa"`
        APIKey   string `env:"PAYMENT_API_KEY,required"`
        Secret   string `env:"PAYMENT_WEBHOOK_SECRET,required"`
    }
    Telegram struct {
        BotToken string `env:"TELEGRAM_BOT_TOKEN,required"`
    }
    Admin struct {
        Token string `env:"ADMIN_TOKEN,required"`
    }
}

func main() {
    // Загружает .env если файл существует; игнорирует ошибку если нет
    // В production .env нет — переменные задаются через Docker/systemd/k8s
    _ = godotenv.Load()

    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        log.Fatal().Err(err).Msg("failed to parse config")
    }
    // ...wire всё вместе...
}
```

Файл `.env` добавляется в `.gitignore`. Пример — `.env.example`.

### 8.5 Dockerfile (multi-stage)

**Файл**: `backend/Dockerfile`

Двухэтапная сборка: на первом этапе компилируется бинарь, на втором — минимальный образ без Go toolchain:

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server/

# Stage 2: Run
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /bin/server /bin/server

# Директория для SQLite базы данных
VOLUME ["/app/data"]

EXPOSE 8080
ENTRYPOINT ["/bin/server"]
```

Примечания:
- `CGO_ENABLED=0` — pure Go сборка (modernc.org/sqlite не требует CGO)
- `ca-certificates` — нужен для HTTPS запросов к AWS S3, YooKassa, Telegram
- `tzdata` — нужен для корректной работы часовых поясов
- SQLite файл хранится в volume `/app/data`, DSN: `DATABASE_DSN=/app/data/mnemo.db`

### 8.6 Docker Compose (Swagger UI + Backend)

```yaml
# docker-compose.yml
services:
  backend:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    volumes:
      - sqlite_data:/app/data
    restart: unless-stopped

  swagger-ui:
    image: swaggerapi/swagger-ui
    ports:
      - "8081:8080"
    volumes:
      - ./api/openapi.yaml:/usr/share/nginx/html/openapi.yaml
    environment:
      SWAGGER_JSON: /usr/share/nginx/html/openapi.yaml
      BASE_URL: /

volumes:
  sqlite_data:
```

`docker compose up` поднимает оба сервиса. `docker compose up swagger-ui` — только документацию без backend.

---

## Шаг 9: Тестирование

### 9.1 Unit тесты (domain)

- `domain/*/` — тесты Value Objects и бизнес-методов агрегатов:
  - `Score.Grade()`, `Score.Passed()`
  - `PromoCode.Activate()`, `PromoCode.Consume()`, `PromoCode.IsValidForStudent()`
  - `Test.Evaluate(answers)`
  - `UserProgress.RecordTestResult()`

### 9.2 Unit тесты (use cases)

- `usecase/*/` — с mockery-моками всех зависимостей
- Цель: coverage > 80%

### 9.3 Интеграционные тесты (repository)

- `repository/sqlite/*_test.go` — с SQLite in-memory + goose миграции
- Тестировать CRUD + сложные JOIN/агрегации

### 9.4 E2E тесты (критические flows)

`tests/e2e/` — HTTP-тесты через `httptest.NewServer`:
1. Study Flow: POST /users → GET /content/modules → POST /study-sessions → POST /test-attempts → PUT /test-attempts/{id}
2. Promo Code Flow: POST /teachers/{id}/promo-codes → POST /users/{id}/subscriptions
3. Payment Flow: POST /payment-invoices → POST /webhooks/payment-gateway

---

## Шаг 10: CI/CD и финализация

### 10.1 Makefile

```makefile
run:          go run ./cmd/server/
build:        go build -o bin/server ./cmd/server/
test:         go test ./...
test-cover:   go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
test-e2e:     go test ./tests/e2e/...
lint:         golangci-lint run
generate:     oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml
migrate-up:   goose -dir database/migrations sqlite3 ${DATABASE_DSN} up
migrate-down: goose -dir database/migrations sqlite3 ${DATABASE_DSN} down
swagger:      docker compose up swagger-ui -d
docker-build: docker compose build
docker-up:    docker compose up -d
docker-down:  docker compose down
```

### 10.2 GitHub Actions (`.github/workflows/backend.yml`)

```yaml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - go test ./...
      - golangci-lint run
      - go vet ./...
  build:
    runs-on: ubuntu-latest
    steps:
      - go build ./...
```

### 10.3 .env.example

```
PORT=8080
HOST=0.0.0.0
DATABASE_DSN=./mnemo.db
AWS_REGION=us-east-1
AWS_BUCKET=mnemo-images
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
PAYMENT_PROVIDER=yookassa
PAYMENT_API_KEY=your_yookassa_key
PAYMENT_WEBHOOK_SECRET=your_webhook_secret
TELEGRAM_BOT_TOKEN=your_bot_token
ADMIN_TOKEN=your_admin_secret_token
```

---

## Порядок выполнения

| Шаг | Результат |
|-----|----------|
| 1. OpenAPI spec + oapi-codegen | Формальный контракт, Go-типы сгенерированы |
| 2. Миграция БД (goose) | Полная схема в одном файле |
| 3. Go-проект, структура, go.mod | `go build ./...` проходит |
| 4. Domain layer (rich entities + VO + interfaces) | Zero external deps в domain |
| 5. Repository layer (raw SQL) + интеграционные тесты | Данные из БД |
| 6. Use case layer + unit тесты | Бизнес-логика |
| 7. Delivery layer (net/http + middleware) | Все 29 endpoints работают |
| 8. Infrastructure (S3, Payment, env config) | Полный сервис |
| 9. E2E тесты | Критические flows проверены |
| 10. CI/CD, Makefile | Готово к деплою |

**MVP-минимум** (Шаги 1–7 + заглушки инфраструктуры): 20 user-facing endpoints работают локально, S3 возвращает фиктивные URL, Payment — фиктивный invoice ID.

---

## Критические файлы для создания

- `backend/api/openapi.yaml`
- `backend/api/oapi-codegen.yaml`
- `backend/internal/api/generated.go` (авто)
- `backend/database/migrations/00001_initial_schema.sql`
- `backend/go.mod`
- `backend/internal/domain/**/*.go`
- `backend/internal/usecase/**/*.go`
- `backend/internal/repository/sqlite/*.go`
- `backend/internal/delivery/http/middleware/*.go`
- `backend/internal/delivery/http/handlers/*.go`
- `backend/internal/delivery/http/router.go`
- `backend/internal/infrastructure/**/*.go`
- `backend/cmd/server/main.go`
- `backend/Dockerfile`
- `backend/docker-compose.yml`
- `backend/Makefile`
- `backend/.env.example`
- `backend/.gitignore` (содержит `.env`, `bin/`, `*.db`)

## Верификация

1. `cp .env.example .env` → заполнить значения → `make docker-up` → backend на :8080, Swagger UI на :8081
2. `make migrate-up` → goose применяет `00001_initial_schema.sql` без ошибок
3. `make test` → все тесты зелёные, coverage > 80%
4. `make lint` → 0 предупреждений
5. `curl -X POST http://localhost:8080/api/v1/users -H "Content-Type: application/json" -d '{"telegram_id": 123, "first_name": "Test"}'` → 201 Created
6. `go test ./tests/e2e/...` → Study Flow, Promo Code Flow, Payment Flow — все проходят
7. `docker compose logs backend` → структурированные zerolog-логи без ошибок
