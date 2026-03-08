# Database Migrations - Миграции базы данных

**Версия**: 1.0
**Дата**: 20 января 2026
**Статус**: Готово к применению

> Пошаговые миграции базы данных для системы подготовки к экзаменам по анатомии

## Содержание

- [Обзор миграций](#обзор-миграций)
- [Migration 001: User Management Fields](#migration-001-user-management-fields)
- [Migration 002: Promo Codes Enhancement](#migration-002-promo-codes-enhancement)
- [Migration 003: User Progress Tracking](#migration-003-user-progress-tracking)
- [Migration 004: Subscriptions Management](#migration-004-subscriptions-management)
- [Migration 005: Teacher-Student Relations](#migration-005-teacher-student-relations)
- [Migration 006: Content Management](#migration-006-content-management)
- [Migration 007: Test Attempts History](#migration-007-test-attempts-history)
- [Migration 008: Constraints and Indexes](#migration-008-constraints-and-indexes)
- [Migration 009: Notifications System](#migration-009-notifications-system)
- [Migration 010: Audit Log](#migration-010-audit-log)
- [Полная миграция](#полная-миграция)
- [Порядок применения](#порядок-применения)

---

## Обзор миграций

### Статистика изменений

| Категория | Количество |
|-----------|-----------|
| Новые поля | 17 |
| Новые таблицы | 3 |
| Новые индексы | 25+ |
| Новые constraints | 10+ |

### Приоритеты миграций

- 🔴 **Критические** (Фаза 1): Migration 001, 002, 003, 004, 005, 006
- 🟡 **Высокие** (Фаза 2): Migration 007, 008
- 🟠 **Средние** (Фаза 3): Migration 009
- 🟢 **Низкие** (Фаза 4): Migration 010

---

## Migration 001: User Management Fields

**Приоритет**: 🔴 КРИТИЧЕСКИЙ (Фаза 1)

**Описание**: Добавление полей для управления платежами и персонализации пользователей

**Изменяемая таблица**: `users`

### Новые поля

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `pending_payment_id` | VARCHAR(255) | ID ожидаемого платежа для отслеживания статуса | 🔴 КРИТИЧЕСКИЙ |
| `last_activity_at` | TIMESTAMP | Последняя активность пользователя (аналитика) | 🟡 ВЫСОКИЙ |
| `language` | VARCHAR(5) | Язык интерфейса (ru/en) | 🟢 НИЗКИЙ |
| `timezone` | VARCHAR(50) | Часовой пояс пользователя | 🟢 НИЗКИЙ |
| `notifications_enabled` | BOOLEAN | Включены ли уведомления | 🟢 НИЗКИЙ |

### SQL Migration (Up)

```sql
-- Migration 001: User Management Fields
-- Date: 2026-01-20
-- Priority: CRITICAL

BEGIN;

-- Критические поля (Фаза 1)
ALTER TABLE users ADD COLUMN pending_payment_id VARCHAR(255);
ALTER TABLE users ADD COLUMN last_activity_at TIMESTAMP;

-- Поля персонализации (Фаза 4)
ALTER TABLE users ADD COLUMN language VARCHAR(5) DEFAULT 'ru';
ALTER TABLE users ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN notifications_enabled BOOLEAN DEFAULT true;

-- Индексы
CREATE INDEX idx_users_pending_payment ON users(pending_payment_id)
WHERE pending_payment_id IS NOT NULL;

CREATE INDEX idx_users_last_activity ON users(last_activity_at DESC);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 001
BEGIN;

DROP INDEX IF EXISTS idx_users_last_activity;
DROP INDEX IF EXISTS idx_users_pending_payment;

ALTER TABLE users DROP COLUMN notifications_enabled;
ALTER TABLE users DROP COLUMN timezone;
ALTER TABLE users DROP COLUMN language;
ALTER TABLE users DROP COLUMN last_activity_at;
ALTER TABLE users DROP COLUMN pending_payment_id;

COMMIT;
```

**Зависимости**: Нет

**Влияние**:
- Разблокирует endpoint `POST /api/v1/users/{user_id}/payment-invoices`
- Разблокирует endpoint `GET /api/v1/users/{user_id}/payment-invoices/pending`

---

## Migration 002: Promo Codes Enhancement

**Приоритет**: 🔴 КРИТИЧЕСКИЙ (Фаза 1)

**Описание**: Добавление срока действия, статуса и аудита промокодов

**Изменяемая таблица**: `promo_codes`

### Новые поля

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `expires_at` | TIMESTAMP | Срок действия промокода | 🔴 КРИТИЧЕСКИЙ |
| `status` | ENUM | Статус промокода (pending/active/expired/deactivated) | 🔴 КРИТИЧЕСКИЙ |
| `created_by_admin_id` | BIGINT | Кто создал промокод (аудит) | 🟡 ВЫСОКИЙ |

### SQL Migration (Up)

```sql
-- Migration 002: Promo Codes Enhancement
-- Date: 2026-01-20
-- Priority: CRITICAL

BEGIN;

-- Критические поля
ALTER TABLE promo_codes ADD COLUMN expires_at TIMESTAMP;
ALTER TABLE promo_codes ADD COLUMN status ENUM('pending', 'active', 'expired', 'deactivated')
DEFAULT 'pending';

-- Поле аудита (Фаза 2)
ALTER TABLE promo_codes ADD COLUMN created_by_admin_id BIGINT;

-- Индексы
CREATE INDEX idx_promo_codes_status ON promo_codes(status);
CREATE INDEX idx_promo_codes_teacher ON promo_codes(teacher_id)
WHERE teacher_id IS NOT NULL;
CREATE INDEX idx_promo_codes_expires ON promo_codes(expires_at);
CREATE INDEX idx_promo_codes_code_status ON promo_codes(code, status);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 002
BEGIN;

DROP INDEX IF EXISTS idx_promo_codes_code_status;
DROP INDEX IF EXISTS idx_promo_codes_expires;
DROP INDEX IF EXISTS idx_promo_codes_teacher;
DROP INDEX IF EXISTS idx_promo_codes_status;

ALTER TABLE promo_codes DROP COLUMN created_by_admin_id;
ALTER TABLE promo_codes DROP COLUMN status;
ALTER TABLE promo_codes DROP COLUMN expires_at;

COMMIT;
```

**Зависимости**: Нет

**Влияние**:
- Разблокирует валидацию промокодов по сроку действия
- Разблокирует endpoint `POST /api/v1/teachers/{teacher_id}/promo-codes`
- Разблокирует endpoint `POST /api/v1/users/{user_id}/subscriptions` (promo flow)

---

## Migration 003: User Progress Tracking

**Приоритет**: 🔴 КРИТИЧЕСКИЙ (Фаза 1)

**Описание**: Расширенное отслеживание прогресса пользователей (статистика попыток и времени)

**Изменяемая таблица**: `user_progress`

### Новые поля

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `current_attempt` | INT | Текущая попытка теста (для статистики) | 🟡 ВЫСОКИЙ |
| `test_started_at` | TIMESTAMP | Время начала теста (для статистики) | 🟡 ВЫСОКИЙ |
| `started_at` | TIMESTAMP | Время начала изучения темы | 🟡 ВЫСОКИЙ |
| `time_spent_seconds` | INT | Общее время на тему (аналитика) | 🟢 НИЗКИЙ |
| `last_viewed_at` | TIMESTAMP | Последний просмотр темы | 🟢 НИЗКИЙ |

### SQL Migration (Up)

```sql
-- Migration 003: User Progress Tracking
-- Date: 2026-01-20
-- Priority: CRITICAL
-- NOTE: Поля для статистики, без ограничений попыток/времени

BEGIN;

-- Поля для статистики (Фаза 2)
ALTER TABLE user_progress ADD COLUMN current_attempt INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN test_started_at TIMESTAMP;
ALTER TABLE user_progress ADD COLUMN started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Дополнительные поля аналитики (Фаза 4)
ALTER TABLE user_progress ADD COLUMN time_spent_seconds INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN last_viewed_at TIMESTAMP;

-- Индексы
CREATE INDEX idx_user_progress_user ON user_progress(user_id, completed_at DESC);
CREATE INDEX idx_user_progress_theme ON user_progress(theme_id, status);
CREATE INDEX idx_user_progress_attempts ON user_progress(current_attempt);
CREATE INDEX idx_user_progress_status_updated ON user_progress(status, updated_at DESC);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 003
BEGIN;

DROP INDEX IF EXISTS idx_user_progress_status_updated;
DROP INDEX IF EXISTS idx_user_progress_attempts;
DROP INDEX IF EXISTS idx_user_progress_theme;
DROP INDEX IF EXISTS idx_user_progress_user;

ALTER TABLE user_progress DROP COLUMN last_viewed_at;
ALTER TABLE user_progress DROP COLUMN time_spent_seconds;
ALTER TABLE user_progress DROP COLUMN started_at;
ALTER TABLE user_progress DROP COLUMN test_started_at;
ALTER TABLE user_progress DROP COLUMN current_attempt;

COMMIT;
```

**Зависимости**: Нет

**Влияние**:
- Разблокирует статистику попыток для аналитики
- Разблокирует отслеживание времени изучения
- Разблокирует endpoint `GET /api/v1/users/{user_id}/progress/modules/{module_id}`

**Важно**: Эти поля используются только для статистики. Ограничения по количеству попыток и времени прохождения тестов НЕ применяются.

---

## Migration 004: Subscriptions Management

**Приоритет**: 🔴 КРИТИЧЕСКИЙ (Фаза 1)

**Описание**: Добавление управления подписками (автопродление, отмена)

**Изменяемая таблица**: `subscriptions`

### Новые поля

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `auto_renew` | BOOLEAN | Автоматическое продление подписки | 🟠 СРЕДНИЙ |
| `cancelled_at` | TIMESTAMP | Дата отмены подписки | 🟢 НИЗКИЙ |
| `cancellation_reason` | TEXT | Причина отмены (аналитика) | 🟢 НИЗКИЙ |

### SQL Migration (Up)

```sql
-- Migration 004: Subscriptions Management
-- Date: 2026-01-20
-- Priority: CRITICAL

BEGIN;

-- Поля управления подпиской (Фаза 3)
ALTER TABLE subscriptions ADD COLUMN auto_renew BOOLEAN DEFAULT false;
ALTER TABLE subscriptions ADD COLUMN cancelled_at TIMESTAMP;
ALTER TABLE subscriptions ADD COLUMN cancellation_reason TEXT;

-- Индексы
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_expires ON subscriptions(expires_at);
CREATE INDEX idx_subscriptions_payment ON subscriptions(payment_id);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status, expires_at);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 004
BEGIN;

DROP INDEX IF EXISTS idx_subscriptions_user_status;
DROP INDEX IF EXISTS idx_subscriptions_payment;
DROP INDEX IF EXISTS idx_subscriptions_expires;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user;

ALTER TABLE subscriptions DROP COLUMN cancellation_reason;
ALTER TABLE subscriptions DROP COLUMN cancelled_at;
ALTER TABLE subscriptions DROP COLUMN auto_renew;

COMMIT;
```

**Зависимости**: Нет

**Влияние**:
- Разблокирует управление автопродлением
- Разблокирует аналитику отмен подписок

---

## Migration 005: Teacher-Student Relations

**Приоритет**: 🔴 КРИТИЧЕСКИЙ (Фаза 1)

**Описание**: Добавление информации о промокоде в связи преподаватель-студент

**Изменяемая таблица**: `teacher_promo_students`

### Новые поля

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `promo_code` | VARCHAR(255) | Код, по которому присоединился студент | 🟡 ВЫСОКИЙ |

### SQL Migration (Up)

```sql
-- Migration 005: Teacher-Student Relations
-- Date: 2026-01-20
-- Priority: CRITICAL

BEGIN;

-- Дополнительные поля (Фаза 2)
ALTER TABLE teacher_promo_students ADD COLUMN promo_code VARCHAR(255);

-- Индексы
CREATE INDEX idx_teacher_promo_teacher ON teacher_promo_students(teacher_id);
CREATE INDEX idx_teacher_promo_student ON teacher_promo_students(student_id);
CREATE INDEX idx_teacher_promo_code ON teacher_promo_students(promo_code);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 005
BEGIN;

DROP INDEX IF EXISTS idx_teacher_promo_code;
DROP INDEX IF EXISTS idx_teacher_promo_student;
DROP INDEX IF EXISTS idx_teacher_promo_teacher;

ALTER TABLE teacher_promo_students DROP COLUMN promo_code;

COMMIT;
```

**Зависимости**: Migration 002 (promo_codes)

**Влияние**:
- Разблокирует аналитику использования промокодов
- Разблокирует endpoint `GET /api/v1/teachers/{teacher_id}/promo-codes`

---

## Migration 006: Content Management

**Приоритет**: 🔴 КРИТИЧЕСКИЙ (Фаза 1)

**Описание**: Добавление блокировки контента, поддержки "Введения" и метаданных

**Изменяемые таблицы**: `modules`, `themes`

### Новые поля для `modules`

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `is_locked` | BOOLEAN | Блокировка модуля | 🟡 ВЫСОКИЙ |
| `icon_emoji` | VARCHAR(10) | Иконка для визуализации в боте | 🟢 НИЗКИЙ |

### Новые поля для `themes`

| Поле | Тип | Назначение | Приоритет |
|------|-----|-----------|-----------|
| `is_introduction` | BOOLEAN | Маркер темы "Введение" | 🔴 КРИТИЧЕСКИЙ |
| `is_locked` | BOOLEAN | Блокировка темы | 🟡 ВЫСОКИЙ |
| `estimated_time_minutes` | INT | Оценка времени изучения | 🟢 НИЗКИЙ |

### SQL Migration (Up)

```sql
-- Migration 006: Content Management
-- Date: 2026-01-20
-- Priority: CRITICAL

BEGIN;

-- Modules: поля блокировки и визуализации
ALTER TABLE modules ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE modules ADD COLUMN icon_emoji VARCHAR(10);

-- Themes: поле "Введение" (HIGH PRIORITY)
ALTER TABLE themes ADD COLUMN is_introduction BOOLEAN DEFAULT false;
ALTER TABLE themes ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE themes ADD COLUMN estimated_time_minutes INT;

-- Индексы для modules
CREATE INDEX idx_modules_locked ON modules(is_locked);

-- Индексы для themes
CREATE INDEX idx_themes_module ON themes(module_id, order_num);
CREATE INDEX idx_themes_locked ON themes(is_locked);
CREATE INDEX idx_themes_introduction ON themes(is_introduction)
WHERE is_introduction = true;

-- Индексы для mnemonics (дополнительно)
CREATE INDEX idx_mnemonics_theme_order ON mnemonics(theme_id, order_num);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 006
BEGIN;

DROP INDEX IF EXISTS idx_mnemonics_theme_order;
DROP INDEX IF EXISTS idx_themes_introduction;
DROP INDEX IF EXISTS idx_themes_locked;
DROP INDEX IF EXISTS idx_themes_module;
DROP INDEX IF EXISTS idx_modules_locked;

ALTER TABLE themes DROP COLUMN estimated_time_minutes;
ALTER TABLE themes DROP COLUMN is_locked;
ALTER TABLE themes DROP COLUMN is_introduction;

ALTER TABLE modules DROP COLUMN icon_emoji;
ALTER TABLE modules DROP COLUMN is_locked;

COMMIT;
```

**Зависимости**: Нет

**Влияние**:
- Разблокирует поддержку сущности "Введение"
- Разблокирует endpoint `GET /api/v1/content/modules/{id}/themes` (с is_introduction)
- Разблокирует гибкое управление доступом к контенту

**Важная логика "Введения"**:
- Первая тема каждого модуля (`order_num = 1`, `is_introduction = true`)
- Для пользователей С подпиской: полный доступ (введение не обязательно)
- Для пользователей БЕЗ подписки: последовательный доступ (начиная с введения)

---

## Migration 007: Test Attempts History

**Приоритет**: 🟡 ВЫСОКИЙ (Фаза 2)

**Описание**: Создание таблицы для хранения истории всех попыток прохождения тестов

**Новая таблица**: `test_attempts`

### Структура таблицы

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INT | Первичный ключ |
| `user_id` | BIGINT | ID пользователя (FK) |
| `theme_id` | INT | ID темы (FK) |
| `test_id` | INT | ID теста (FK) |
| `answers` | JSONB | Ответы пользователя |
| `score` | INT | Полученный балл |
| `passed` | BOOLEAN | Пройден ли тест |
| `started_at` | TIMESTAMP | Время начала |
| `submitted_at` | TIMESTAMP | Время отправки |
| `duration_seconds` | INT | Длительность попытки |
| `ip_address` | VARCHAR(45) | IP адрес (защита от мошенничества) |
| `user_agent` | TEXT | User agent (аудит) |

### SQL Migration (Up)

```sql
-- Migration 007: Test Attempts History
-- Date: 2026-01-20
-- Priority: HIGH

BEGIN;

CREATE TABLE IF NOT EXISTS test_attempts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    theme_id INT NOT NULL,
    test_id INT NOT NULL,

    -- Данные попытки
    answers JSONB NOT NULL,
    score INT NOT NULL,
    passed BOOLEAN NOT NULL,

    -- Временные метки
    started_at TIMESTAMP NOT NULL,
    submitted_at TIMESTAMP NOT NULL,
    duration_seconds INT NOT NULL,

    -- Метаданные (защита от мошенничества)
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
    FOREIGN KEY (theme_id) REFERENCES themes(id) ON DELETE CASCADE,
    FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE,

    -- Индексы
    INDEX idx_test_attempts_user_theme (user_id, theme_id, submitted_at DESC),
    INDEX idx_test_attempts_test (test_id, passed),
    INDEX idx_test_attempts_user (user_id, submitted_at DESC)
);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 007
BEGIN;

DROP TABLE IF EXISTS test_attempts;

COMMIT;
```

**Зависимости**: Migration 006 (themes)

**Влияние**:
- Разблокирует историю всех попыток тестов
- Разблокирует аналитику: какие вопросы вызывают трудности
- Разблокирует построение графиков прогресса
- Разблокирует защиту от мошенничества (IP, user agent)

**Использование**:
- Построение графика динамики баллов
- Определение проблемных вопросов
- Аудит прохождения тестов

---

## Migration 008: Constraints and Indexes

**Приоритет**: 🟡 ВЫСОКИЙ (Фаза 2)

**Описание**: Добавление constraints для обеспечения целостности данных

### SQL Migration (Up)

```sql
-- Migration 008: Constraints and Indexes
-- Date: 2026-01-20
-- Priority: HIGH

BEGIN;

-- ========== UNIQUE CONSTRAINTS ==========

-- Защита от дубликатов промокодов студентов
ALTER TABLE teacher_promo_students
ADD CONSTRAINT unique_teacher_student UNIQUE (teacher_id, student_id);

-- Защита от дублирования платежей
ALTER TABLE subscriptions
ADD CONSTRAINT unique_payment_id UNIQUE (payment_id);

-- ========== CHECK CONSTRAINTS ==========

-- Валидация баллов
ALTER TABLE user_progress
ADD CONSTRAINT check_score CHECK (score >= 0 AND score <= 100);

ALTER TABLE tests
ADD CONSTRAINT check_passing_score CHECK (passing_score >= 0 AND passing_score <= 100);

-- Валидация попыток (только для статистики, без ограничений)
ALTER TABLE user_progress
ADD CONSTRAINT check_current_attempt CHECK (current_attempt >= 0);

-- Валидация промокодов
ALTER TABLE promo_codes
ADD CONSTRAINT check_remaining CHECK (remaining >= 0);

ALTER TABLE promo_codes
ADD CONSTRAINT check_max_activations CHECK (max_activations > 0);

-- Валидация subscriptions
ALTER TABLE subscriptions
ADD CONSTRAINT check_amount CHECK (amount >= 0);

-- ========== ADDITIONAL INDEXES ==========

-- Частые запросы по тестам
CREATE INDEX idx_tests_theme ON tests(theme_id);
CREATE INDEX idx_tests_difficulty ON tests(difficulty);
CREATE INDEX idx_tests_theme_difficulty ON tests(theme_id, difficulty);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 008
BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS idx_tests_theme_difficulty;
DROP INDEX IF EXISTS idx_tests_difficulty;
DROP INDEX IF EXISTS idx_tests_theme;

-- Drop check constraints
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS check_amount;
ALTER TABLE promo_codes DROP CONSTRAINT IF EXISTS check_max_activations;
ALTER TABLE promo_codes DROP CONSTRAINT IF EXISTS check_remaining;
ALTER TABLE user_progress DROP CONSTRAINT IF EXISTS check_current_attempt;
ALTER TABLE tests DROP CONSTRAINT IF EXISTS check_passing_score;
ALTER TABLE user_progress DROP CONSTRAINT IF EXISTS check_score;

-- Drop unique constraints
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS unique_payment_id;
ALTER TABLE teacher_promo_students DROP CONSTRAINT IF EXISTS unique_teacher_student;

COMMIT;
```

**Зависимости**: Migration 001, 002, 003, 004, 005

**Влияние**:
- Обеспечивает целостность данных
- Предотвращает дубликаты
- Ускоряет запросы

---

## Migration 009: Notifications System

**Приоритет**: 🟢 НИЗКИЙ (Фаза 4)

**Описание**: Создание таблицы для хранения истории уведомлений

**Новая таблица**: `notifications`

### SQL Migration (Up)

```sql
-- Migration 009: Notifications System
-- Date: 2026-01-20
-- Priority: LOW

BEGIN;

CREATE TABLE IF NOT EXISTS notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,

    -- Тип уведомления
    type ENUM('test_result', 'subscription_expiring', 'subscription_activated',
              'promo_activated', 'student_joined') NOT NULL,

    -- Содержимое
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,

    -- Статус
    status ENUM('pending', 'sent', 'failed') DEFAULT 'pending',
    sent_at TIMESTAMP,
    error_message TEXT,

    -- Метаданные
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key
    FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE,

    -- Индексы
    INDEX idx_notifications_user (user_id, created_at DESC),
    INDEX idx_notifications_status (status, created_at)
);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 009
BEGIN;

DROP TABLE IF EXISTS notifications;

COMMIT;
```

**Зависимости**: Migration 001 (users)

**Влияние**:
- История всех уведомлений пользователя
- Повторная отправка при ошибке
- Аналитика открываемости

---

## Migration 010: Audit Log

**Приоритет**: 🟢 НИЗКИЙ (Фаза 4)

**Описание**: Создание таблицы для аудита действий администраторов

**Новая таблица**: `audit_log`

### SQL Migration (Up)

```sql
-- Migration 010: Audit Log
-- Date: 2026-01-20
-- Priority: LOW

BEGIN;

CREATE TABLE IF NOT EXISTS audit_log (
    id INT PRIMARY KEY AUTO_INCREMENT,

    -- Кто
    admin_id BIGINT NOT NULL,
    admin_username VARCHAR(255),

    -- Что
    action VARCHAR(100) NOT NULL, -- 'create_promo', 'deactivate_promo', 'delete_user', etc.
    entity_type VARCHAR(50) NOT NULL, -- 'promo_code', 'user', 'module', etc.
    entity_id VARCHAR(255) NOT NULL,

    -- Детали
    old_value JSONB,
    new_value JSONB,

    -- Когда
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Метаданные
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Индексы
    INDEX idx_audit_admin (admin_id, created_at DESC),
    INDEX idx_audit_entity (entity_type, entity_id),
    INDEX idx_audit_action (action, created_at DESC)
);

COMMIT;
```

### SQL Rollback (Down)

```sql
-- Rollback Migration 010
BEGIN;

DROP TABLE IF EXISTS audit_log;

COMMIT;
```

**Зависимости**: Нет

**Влияние**:
- Отслеживание действий администраторов
- Compliance и безопасность
- Восстановление данных

---

## Полная миграция

### Скрипт полной миграции (все миграции за один раз)

```sql
-- ==========================================
-- Full Migration Script: All improvements
-- Version: 1.0
-- Date: 2026-01-20
-- ==========================================

BEGIN;

-- ========== Migration 001: User Management Fields ==========
ALTER TABLE users ADD COLUMN pending_payment_id VARCHAR(255);
ALTER TABLE users ADD COLUMN last_activity_at TIMESTAMP;
ALTER TABLE users ADD COLUMN language VARCHAR(5) DEFAULT 'ru';
ALTER TABLE users ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN notifications_enabled BOOLEAN DEFAULT true;

CREATE INDEX idx_users_pending_payment ON users(pending_payment_id) WHERE pending_payment_id IS NOT NULL;
CREATE INDEX idx_users_last_activity ON users(last_activity_at DESC);

-- ========== Migration 002: Promo Codes Enhancement ==========
ALTER TABLE promo_codes ADD COLUMN expires_at TIMESTAMP;
ALTER TABLE promo_codes ADD COLUMN status ENUM('pending', 'active', 'expired', 'deactivated') DEFAULT 'pending';
ALTER TABLE promo_codes ADD COLUMN created_by_admin_id BIGINT;

CREATE INDEX idx_promo_codes_status ON promo_codes(status);
CREATE INDEX idx_promo_codes_teacher ON promo_codes(teacher_id) WHERE teacher_id IS NOT NULL;
CREATE INDEX idx_promo_codes_expires ON promo_codes(expires_at);
CREATE INDEX idx_promo_codes_code_status ON promo_codes(code, status);

-- ========== Migration 003: User Progress Tracking ==========
ALTER TABLE user_progress ADD COLUMN current_attempt INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN test_started_at TIMESTAMP;
ALTER TABLE user_progress ADD COLUMN started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE user_progress ADD COLUMN time_spent_seconds INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN last_viewed_at TIMESTAMP;

CREATE INDEX idx_user_progress_user ON user_progress(user_id, completed_at DESC);
CREATE INDEX idx_user_progress_theme ON user_progress(theme_id, status);
CREATE INDEX idx_user_progress_attempts ON user_progress(current_attempt);
CREATE INDEX idx_user_progress_status_updated ON user_progress(status, updated_at DESC);

-- ========== Migration 004: Subscriptions Management ==========
ALTER TABLE subscriptions ADD COLUMN auto_renew BOOLEAN DEFAULT false;
ALTER TABLE subscriptions ADD COLUMN cancelled_at TIMESTAMP;
ALTER TABLE subscriptions ADD COLUMN cancellation_reason TEXT;

CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_expires ON subscriptions(expires_at);
CREATE INDEX idx_subscriptions_payment ON subscriptions(payment_id);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status, expires_at);

-- ========== Migration 005: Teacher-Student Relations ==========
ALTER TABLE teacher_promo_students ADD COLUMN promo_code VARCHAR(255);

CREATE INDEX idx_teacher_promo_teacher ON teacher_promo_students(teacher_id);
CREATE INDEX idx_teacher_promo_student ON teacher_promo_students(student_id);
CREATE INDEX idx_teacher_promo_code ON teacher_promo_students(promo_code);

-- ========== Migration 006: Content Management ==========
ALTER TABLE modules ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE modules ADD COLUMN icon_emoji VARCHAR(10);

ALTER TABLE themes ADD COLUMN is_introduction BOOLEAN DEFAULT false;
ALTER TABLE themes ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE themes ADD COLUMN estimated_time_minutes INT;

CREATE INDEX idx_modules_locked ON modules(is_locked);
CREATE INDEX idx_themes_module ON themes(module_id, order_num);
CREATE INDEX idx_themes_locked ON themes(is_locked);
CREATE INDEX idx_themes_introduction ON themes(is_introduction) WHERE is_introduction = true;
CREATE INDEX idx_mnemonics_theme_order ON mnemonics(theme_id, order_num);

-- ========== Migration 007: Test Attempts History ==========
CREATE TABLE IF NOT EXISTS test_attempts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    theme_id INT NOT NULL,
    test_id INT NOT NULL,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    passed BOOLEAN NOT NULL,
    started_at TIMESTAMP NOT NULL,
    submitted_at TIMESTAMP NOT NULL,
    duration_seconds INT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
    FOREIGN KEY (theme_id) REFERENCES themes(id) ON DELETE CASCADE,
    FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE,
    INDEX idx_test_attempts_user_theme (user_id, theme_id, submitted_at DESC),
    INDEX idx_test_attempts_test (test_id, passed),
    INDEX idx_test_attempts_user (user_id, submitted_at DESC)
);

-- ========== Migration 008: Constraints and Indexes ==========
ALTER TABLE teacher_promo_students ADD CONSTRAINT unique_teacher_student UNIQUE (teacher_id, student_id);
ALTER TABLE subscriptions ADD CONSTRAINT unique_payment_id UNIQUE (payment_id);

ALTER TABLE user_progress ADD CONSTRAINT check_score CHECK (score >= 0 AND score <= 100);
ALTER TABLE tests ADD CONSTRAINT check_passing_score CHECK (passing_score >= 0 AND passing_score <= 100);
ALTER TABLE user_progress ADD CONSTRAINT check_current_attempt CHECK (current_attempt >= 0);
ALTER TABLE promo_codes ADD CONSTRAINT check_remaining CHECK (remaining >= 0);
ALTER TABLE promo_codes ADD CONSTRAINT check_max_activations CHECK (max_activations > 0);
ALTER TABLE subscriptions ADD CONSTRAINT check_amount CHECK (amount >= 0);

CREATE INDEX idx_tests_theme ON tests(theme_id);
CREATE INDEX idx_tests_difficulty ON tests(difficulty);
CREATE INDEX idx_tests_theme_difficulty ON tests(theme_id, difficulty);

-- ========== Migration 009: Notifications System (Optional - Фаза 4) ==========
CREATE TABLE IF NOT EXISTS notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    type ENUM('test_result', 'subscription_expiring', 'subscription_activated', 'promo_activated', 'student_joined') NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status ENUM('pending', 'sent', 'failed') DEFAULT 'pending',
    sent_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
    INDEX idx_notifications_user (user_id, created_at DESC),
    INDEX idx_notifications_status (status, created_at)
);

-- ========== Migration 010: Audit Log (Optional - Фаза 4) ==========
CREATE TABLE IF NOT EXISTS audit_log (
    id INT PRIMARY KEY AUTO_INCREMENT,
    admin_id BIGINT NOT NULL,
    admin_username VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    INDEX idx_audit_admin (admin_id, created_at DESC),
    INDEX idx_audit_entity (entity_type, entity_id),
    INDEX idx_audit_action (action, created_at DESC)
);

COMMIT;
```

### Скрипт полного отката (rollback)

```sql
-- ==========================================
-- Full Rollback Script: Revert all migrations
-- ==========================================

BEGIN;

-- Drop new tables (reverse order)
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS test_attempts;

-- Drop constraints
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS check_amount;
ALTER TABLE promo_codes DROP CONSTRAINT IF EXISTS check_max_activations;
ALTER TABLE promo_codes DROP CONSTRAINT IF EXISTS check_remaining;
ALTER TABLE user_progress DROP CONSTRAINT IF EXISTS check_current_attempt;
ALTER TABLE tests DROP CONSTRAINT IF EXISTS check_passing_score;
ALTER TABLE user_progress DROP CONSTRAINT IF EXISTS check_score;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS unique_payment_id;
ALTER TABLE teacher_promo_students DROP CONSTRAINT IF EXISTS unique_teacher_student;

-- Drop all indexes (reverse order)
DROP INDEX IF EXISTS idx_tests_theme_difficulty;
DROP INDEX IF EXISTS idx_tests_difficulty;
DROP INDEX IF EXISTS idx_tests_theme;
DROP INDEX IF EXISTS idx_mnemonics_theme_order;
DROP INDEX IF EXISTS idx_themes_introduction;
DROP INDEX IF EXISTS idx_themes_locked;
DROP INDEX IF EXISTS idx_themes_module;
DROP INDEX IF EXISTS idx_modules_locked;
DROP INDEX IF EXISTS idx_teacher_promo_code;
DROP INDEX IF EXISTS idx_teacher_promo_student;
DROP INDEX IF EXISTS idx_teacher_promo_teacher;
DROP INDEX IF EXISTS idx_subscriptions_user_status;
DROP INDEX IF EXISTS idx_subscriptions_payment;
DROP INDEX IF EXISTS idx_subscriptions_expires;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user;
DROP INDEX IF EXISTS idx_user_progress_status_updated;
DROP INDEX IF EXISTS idx_user_progress_attempts;
DROP INDEX IF EXISTS idx_user_progress_theme;
DROP INDEX IF EXISTS idx_user_progress_user;
DROP INDEX IF EXISTS idx_promo_codes_code_status;
DROP INDEX IF EXISTS idx_promo_codes_expires;
DROP INDEX IF EXISTS idx_promo_codes_teacher;
DROP INDEX IF EXISTS idx_promo_codes_status;
DROP INDEX IF EXISTS idx_users_last_activity;
DROP INDEX IF EXISTS idx_users_pending_payment;

-- Drop columns (reverse order)
ALTER TABLE themes DROP COLUMN IF EXISTS estimated_time_minutes;
ALTER TABLE themes DROP COLUMN IF EXISTS is_locked;
ALTER TABLE themes DROP COLUMN IF EXISTS is_introduction;
ALTER TABLE modules DROP COLUMN IF EXISTS icon_emoji;
ALTER TABLE modules DROP COLUMN IF EXISTS is_locked;
ALTER TABLE teacher_promo_students DROP COLUMN IF EXISTS promo_code;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS cancellation_reason;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS auto_renew;
ALTER TABLE user_progress DROP COLUMN IF EXISTS last_viewed_at;
ALTER TABLE user_progress DROP COLUMN IF EXISTS time_spent_seconds;
ALTER TABLE user_progress DROP COLUMN IF EXISTS started_at;
ALTER TABLE user_progress DROP COLUMN IF EXISTS test_started_at;
ALTER TABLE user_progress DROP COLUMN IF EXISTS current_attempt;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS created_by_admin_id;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS status;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS notifications_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS timezone;
ALTER TABLE users DROP COLUMN IF EXISTS language;
ALTER TABLE users DROP COLUMN IF EXISTS last_activity_at;
ALTER TABLE users DROP COLUMN IF EXISTS pending_payment_id;

COMMIT;
```

---

## Порядок применения

### Рекомендуемая последовательность

#### Фаза 1: MVP (Критические миграции)

Применить в следующем порядке:

1. **Migration 001** (User Management Fields)
2. **Migration 002** (Promo Codes Enhancement)
3. **Migration 003** (User Progress Tracking)
4. **Migration 004** (Subscriptions Management)
5. **Migration 005** (Teacher-Student Relations)
6. **Migration 006** (Content Management) ⚠️ **ОБЯЗАТЕЛЬНО** для поддержки "Введения"

**Результат**: MVP готов к запуску

#### Фаза 2: Надежность (Высокоприоритетные миграции)

7. **Migration 007** (Test Attempts History)
8. **Migration 008** (Constraints and Indexes)

**Результат**: Надежная и защищенная система

#### Фаза 3: Расширенные возможности (Средний приоритет)

9. **Migration 009** (Notifications System) - опционально

**Результат**: Система с уведомлениями

#### Фаза 4: Дополнительный функционал (Низкий приоритет)

10. **Migration 010** (Audit Log) - опционально

**Результат**: Полнофункциональная система с аудитом

### Проверка после каждой миграции

```sql
-- Проверка структуры таблицы
DESCRIBE table_name;

-- Проверка индексов
SHOW INDEXES FROM table_name;

-- Проверка constraints
SELECT * FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
WHERE TABLE_NAME = 'table_name';
```

### Важные замечания

⚠️ **Критические изменения**:
- Migration 006 обязательна для поддержки сущности "Введение" (is_introduction)
- Migration 001 обязательна для платежного функционала
- Migration 002 обязательна для промокодов с истечением срока

✅ **Безопасность**:
- Все миграции используют транзакции (BEGIN/COMMIT)
- Предусмотрены rollback скрипты для каждой миграции
- Check constraints обеспечивают валидацию данных

📊 **Производительность**:
- Индексы создаются для всех часто используемых запросов
- Композитные индексы для сложных запросов
- Partial индексы для оптимизации памяти

---

**Дата создания**: 2026-01-20
**Версия**: 1.0
**Статус**: ✅ Готово к применению
