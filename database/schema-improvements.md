# Database Schema Improvements - Улучшения схемы БД

**Версия**: 2.1
**Дата**: 19 января 2026
**Статус**: Актуально (включает поле is_introduction из SYSTEM_SPECIFICATION v2.1)

> Дополнительные поля, таблицы и constraints для полной функциональности системы

## Changelog

- **v2.1** (19.01.2026):
  - ✅ Добавлено поле `themes.is_introduction` для поддержки сущности "Введение"
  - 📊 Всего новых полей: 19 (было 18)

- **v1.0** (28.12.2025):
  - Базовая спецификация улучшений БД

## Содержание

- [1. Обновления существующих таблиц](#1-обновления-существующих-таблиц)
- [2. Новые таблицы](#2-новые-таблицы)
- [3. Constraints и индексы](#3-constraints-и-индексы)
- [4. Полная миграция](#4-полная-миграция)
- [5. Seed данные для тестирования](#5-seed-данные-для-тестирования)

---

## 1. Обновления существующих таблиц

### 1.1 Таблица `users`

**Назначение**: Добавление полей для отслеживания платежей и персонализации

```sql
-- Критические поля (Фаза 1)
ALTER TABLE users ADD COLUMN pending_payment_id VARCHAR(255);

-- Поля персонализации (Фаза 4)
ALTER TABLE users ADD COLUMN language VARCHAR(5) DEFAULT 'ru';
ALTER TABLE users ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN notifications_enabled BOOLEAN DEFAULT true;

-- Временные метки
ALTER TABLE users ADD COLUMN last_activity_at TIMESTAMP;

-- Индексы
CREATE INDEX idx_users_pending_payment ON users(pending_payment_id) WHERE pending_payment_id IS NOT NULL;
CREATE INDEX idx_users_last_activity ON users(last_activity_at DESC);
```

**Обоснование**:
- `pending_payment_id` - отслеживание ожидаемых платежей (критично для /check_payment)
- `language`, `timezone`, `notifications_enabled` - персонализация (низкий приоритет)
- `last_activity_at` - для аналитики активности пользователей

---

### 1.2 Таблица `promo_codes`

**Назначение**: Добавление срока действия и статуса промокодов

```sql
-- Критические поля (Фаза 1)
ALTER TABLE promo_codes ADD COLUMN expires_at TIMESTAMP;
ALTER TABLE promo_codes ADD COLUMN status ENUM('pending', 'active', 'expired', 'deactivated') DEFAULT 'pending';

-- Дополнительные поля (Фаза 2)
ALTER TABLE promo_codes ADD COLUMN created_by_admin_id BIGINT;

-- Индексы
CREATE INDEX idx_promo_codes_status ON promo_codes(status);
CREATE INDEX idx_promo_codes_teacher ON promo_codes(teacher_id) WHERE teacher_id IS NOT NULL;
CREATE INDEX idx_promo_codes_expires ON promo_codes(expires_at);

-- Constraint для автоматического истечения
-- (можно использовать trigger или cronjob)
```

**Обоснование**:
- `expires_at` - срок действия промокодов (требование бизнес-логики)
- `status` - явное отслеживание состояния (pending/active/expired/deactivated)
- `created_by_admin_id` - аудит создания промокодов

**Trigger для автоматического истечения** (опционально):
```sql
CREATE TRIGGER update_promo_code_status
BEFORE SELECT ON promo_codes
FOR EACH ROW
BEGIN
    IF NEW.expires_at < NOW() AND NEW.status = 'active' THEN
        UPDATE promo_codes SET status = 'expired' WHERE code = NEW.code;
    END IF;
END;
```

---

### 1.3 Таблица `tests`

**Назначение**: Добавление ограничений попыток и времени

```sql
-- Критические поля (Фаза 2)
ALTER TABLE tests ADD COLUMN max_attempts INT DEFAULT 3;
ALTER TABLE tests ADD COLUMN time_limit_minutes INT;

-- Дополнительные поля
ALTER TABLE tests ADD COLUMN shuffle_questions BOOLEAN DEFAULT true;
ALTER TABLE tests ADD COLUMN shuffle_answers BOOLEAN DEFAULT true;

-- Индексы
CREATE INDEX idx_tests_theme ON tests(theme_id);
CREATE INDEX idx_tests_difficulty ON tests(difficulty);
```

**Обоснование**:
- `max_attempts` - ограничение количества попыток (защита от подбора)
- `time_limit_minutes` - ограничение времени (защита от поиска ответов)
- `shuffle_questions`, `shuffle_answers` - рандомизация для уменьшения списывания

---

### 1.4 Таблица `user_progress`

**Назначение**: Расширенное отслеживание прогресса

```sql
-- Критические поля (Фаза 2)
ALTER TABLE user_progress ADD COLUMN current_attempt INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN test_started_at TIMESTAMP;
ALTER TABLE user_progress ADD COLUMN started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Дополнительные поля
ALTER TABLE user_progress ADD COLUMN time_spent_seconds INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN last_viewed_at TIMESTAMP;

-- Обновление индексов
CREATE INDEX idx_user_progress_user ON user_progress(user_id, completed_at DESC);
CREATE INDEX idx_user_progress_theme ON user_progress(theme_id, status);
CREATE INDEX idx_user_progress_attempts ON user_progress(current_attempt);
```

**Обоснование**:
- `current_attempt` - текущая попытка теста (для проверки max_attempts)
- `test_started_at` - время начала теста (для проверки time_limit)
- `started_at` - время начала изучения темы (аналитика)
- `time_spent_seconds` - общее время на тему (аналитика)

---

### 1.5 Таблица `subscriptions`

**Назначение**: Добавление автопродления и отмены

```sql
-- Поля управления подпиской (Фаза 3)
ALTER TABLE subscriptions ADD COLUMN auto_renew BOOLEAN DEFAULT false;
ALTER TABLE subscriptions ADD COLUMN cancelled_at TIMESTAMP;
ALTER TABLE subscriptions ADD COLUMN cancellation_reason TEXT;

-- Индексы
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_expires ON subscriptions(expires_at);
CREATE INDEX idx_subscriptions_payment ON subscriptions(payment_id);
```

**Обоснование**:
- `auto_renew` - управление автоматическим продлением
- `cancelled_at`, `cancellation_reason` - аудит отмен (для аналитики)

---

### 1.6 Таблица `teacher_promo_students`

**Назначение**: Добавление информации о промокоде

```sql
-- Дополнительные поля (Фаза 2)
ALTER TABLE teacher_promo_students ADD COLUMN promo_code VARCHAR(255);

-- Индексы
CREATE INDEX idx_teacher_promo_teacher ON teacher_promo_students(teacher_id);
CREATE INDEX idx_teacher_promo_student ON teacher_promo_students(student_id);
CREATE INDEX idx_teacher_promo_code ON teacher_promo_students(promo_code);
```

**Обоснование**:
- `promo_code` - какой код использовал студент (для аналитики)

---

### 1.7 Таблицы `modules` и `themes`

**Назначение**: Добавление блокировки контента и поддержки "Введения"

```sql
-- Поля блокировки (Фаза 2)
ALTER TABLE modules ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE themes ADD COLUMN is_locked BOOLEAN DEFAULT false;

-- Поле для "Введения" (Фаза 1) 🟡 HIGH PRIORITY (v2.1)
ALTER TABLE themes ADD COLUMN is_introduction BOOLEAN DEFAULT false;

-- Дополнительные поля
ALTER TABLE modules ADD COLUMN icon_emoji VARCHAR(10);
ALTER TABLE themes ADD COLUMN estimated_time_minutes INT;

-- Индексы
CREATE INDEX idx_modules_locked ON modules(is_locked);
CREATE INDEX idx_themes_module ON themes(module_id, order_num);
CREATE INDEX idx_themes_locked ON themes(is_locked);
CREATE INDEX idx_themes_introduction ON themes(is_introduction) WHERE is_introduction = true;
```

**Обоснование**:
- `is_introduction` - маркер первой темы модуля (Введение с терминами) 🟡 **HIGH PRIORITY**
- `is_locked` - гибкое управление доступом к контенту
- `icon_emoji` - визуализация в боте
- `estimated_time_minutes` - оценка времени изучения

**Логика "Введения"**:
- Первая тема каждого модуля (`order_num = 1`, `is_introduction = true`)
- Для пользователей С подпиской: полный доступ (введение не обязательно)
- Для пользователей БЕЗ подписки: последовательный доступ (начиная с введения)

---

## 2. Новые таблицы

### 2.1 Таблица `test_attempts`

**Назначение**: История всех попыток прохождения тестов

**Приоритет**: 🔴 ВЫСОКИЙ (Фаза 2)

```sql
CREATE TABLE test_attempts (
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

    -- Метаданные
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
```

**Обоснование**:
- Хранение истории всех попыток (не только последней)
- Аналитика: какие вопросы вызывают трудности
- Защита от мошенничества (IP, user agent)
- Построение графиков прогресса

**Использование**:
- Построение графика динамики баллов
- Определение проблемных вопросов
- Аудит прохождения тестов

---

### 2.2 Таблица `notifications`

**Назначение**: История отправленных уведомлений

**Приоритет**: 🟢 НИЗКИЙ (Фаза 4)

```sql
CREATE TABLE notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,

    -- Тип уведомления
    type ENUM('test_result', 'subscription_expiring', 'subscription_activated', 'promo_activated', 'student_joined') NOT NULL,

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
```

**Обоснование**:
- История всех уведомлений пользователя
- Повторная отправка при ошибке
- Аналитика открываемости

---

### 2.3 Таблица `audit_log`

**Назначение**: Аудит действий администраторов

**Приоритет**: 🟡 СРЕДНИЙ (Фаза 3)

```sql
CREATE TABLE audit_log (
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
```

**Обоснование**:
- Отслеживание действий администраторов
- Compliance и безопасность
- Восстановление данных

---

## 3. Constraints и индексы

### 3.1 Уникальные constraints

```sql
-- Защита от дубликатов промокодов студентов
ALTER TABLE teacher_promo_students
ADD CONSTRAINT unique_teacher_student UNIQUE (teacher_id, student_id);

-- Защита от дублирования платежей
ALTER TABLE subscriptions
ADD CONSTRAINT unique_payment_id UNIQUE (payment_id);

-- Составной первичный ключ в user_progress (уже есть)
-- PRIMARY KEY (user_id, theme_id)
```

---

### 3.2 Check constraints

```sql
-- Валидация баллов
ALTER TABLE user_progress
ADD CONSTRAINT check_score CHECK (score >= 0 AND score <= 100);

ALTER TABLE tests
ADD CONSTRAINT check_passing_score CHECK (passing_score >= 0 AND passing_score <= 100);

-- Валидация попыток
ALTER TABLE tests
ADD CONSTRAINT check_max_attempts CHECK (max_attempts > 0);

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
```

---

### 3.3 Foreign key constraints

```sql
-- Убедиться, что все FK имеют ON DELETE и ON UPDATE правила

-- user_progress
ALTER TABLE user_progress
MODIFY CONSTRAINT fk_user_progress_user
FOREIGN KEY (user_id) REFERENCES users(telegram_id)
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE user_progress
MODIFY CONSTRAINT fk_user_progress_theme
FOREIGN KEY (theme_id) REFERENCES themes(id)
ON DELETE CASCADE ON UPDATE CASCADE;

-- subscriptions
ALTER TABLE subscriptions
MODIFY CONSTRAINT fk_subscriptions_user
FOREIGN KEY (user_id) REFERENCES users(telegram_id)
ON DELETE CASCADE ON UPDATE CASCADE;

-- promo_codes
ALTER TABLE promo_codes
ADD CONSTRAINT fk_promo_codes_teacher
FOREIGN KEY (teacher_id) REFERENCES users(telegram_id)
ON DELETE SET NULL ON UPDATE CASCADE;

-- teacher_promo_students
ALTER TABLE teacher_promo_students
MODIFY CONSTRAINT fk_teacher_promo_teacher
FOREIGN KEY (teacher_id) REFERENCES users(telegram_id)
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE teacher_promo_students
MODIFY CONSTRAINT fk_teacher_promo_student
FOREIGN KEY (student_id) REFERENCES users(telegram_id)
ON DELETE CASCADE ON UPDATE CASCADE;
```

---

### 3.4 Дополнительные индексы для производительности

```sql
-- Частые запросы по подпискам
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status, expires_at);

-- Поиск промокодов
CREATE INDEX idx_promo_codes_code_status ON promo_codes(code, status);

-- Прогресс по темам
CREATE INDEX idx_user_progress_status_updated ON user_progress(status, updated_at DESC);

-- Тесты по темам и сложности
CREATE INDEX idx_tests_theme_difficulty ON tests(theme_id, difficulty);

-- Мнемоники по темам
CREATE INDEX idx_mnemonics_theme_order ON mnemonics(theme_id, order_num);
```

---

## 4. Полная миграция

### 4.1 Скрипт миграции (SQL)

```sql
-- ==========================================
-- Migration Script: Add all improvements
-- Version: 1.0
-- Date: 2025-12-28
-- ==========================================

BEGIN TRANSACTION;

-- ========== 1. UPDATE EXISTING TABLES ==========

-- 1.1 users
ALTER TABLE users ADD COLUMN pending_payment_id VARCHAR(255);
ALTER TABLE users ADD COLUMN language VARCHAR(5) DEFAULT 'ru';
ALTER TABLE users ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN notifications_enabled BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN last_activity_at TIMESTAMP;

-- 1.2 promo_codes
ALTER TABLE promo_codes ADD COLUMN expires_at TIMESTAMP;
ALTER TABLE promo_codes ADD COLUMN status ENUM('pending', 'active', 'expired', 'deactivated') DEFAULT 'pending';
ALTER TABLE promo_codes ADD COLUMN created_by_admin_id BIGINT;

-- 1.3 tests
ALTER TABLE tests ADD COLUMN max_attempts INT DEFAULT 3;
ALTER TABLE tests ADD COLUMN time_limit_minutes INT;
ALTER TABLE tests ADD COLUMN shuffle_questions BOOLEAN DEFAULT true;
ALTER TABLE tests ADD COLUMN shuffle_answers BOOLEAN DEFAULT true;

-- 1.4 user_progress
ALTER TABLE user_progress ADD COLUMN current_attempt INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN test_started_at TIMESTAMP;
ALTER TABLE user_progress ADD COLUMN started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE user_progress ADD COLUMN time_spent_seconds INT DEFAULT 0;
ALTER TABLE user_progress ADD COLUMN last_viewed_at TIMESTAMP;

-- 1.5 subscriptions
ALTER TABLE subscriptions ADD COLUMN auto_renew BOOLEAN DEFAULT false;
ALTER TABLE subscriptions ADD COLUMN cancelled_at TIMESTAMP;
ALTER TABLE subscriptions ADD COLUMN cancellation_reason TEXT;

-- 1.6 teacher_promo_students
ALTER TABLE teacher_promo_students ADD COLUMN promo_code VARCHAR(255);

-- 1.7 modules and themes
ALTER TABLE modules ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE modules ADD COLUMN icon_emoji VARCHAR(10);

ALTER TABLE themes ADD COLUMN is_locked BOOLEAN DEFAULT false;
ALTER TABLE themes ADD COLUMN is_introduction BOOLEAN DEFAULT false; -- v2.1 🟡 HIGH
ALTER TABLE themes ADD COLUMN estimated_time_minutes INT;

-- ========== 2. CREATE NEW TABLES ==========

-- 2.1 test_attempts
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
    FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE
);

-- 2.2 notifications (optional)
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
    FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
);

-- 2.3 audit_log (optional)
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
    user_agent TEXT
);

-- ========== 3. ADD CONSTRAINTS ==========

-- 3.1 Unique constraints
ALTER TABLE teacher_promo_students
ADD CONSTRAINT unique_teacher_student UNIQUE (teacher_id, student_id);

ALTER TABLE subscriptions
ADD CONSTRAINT unique_payment_id UNIQUE (payment_id);

-- 3.2 Check constraints
ALTER TABLE user_progress
ADD CONSTRAINT check_score CHECK (score >= 0 AND score <= 100);

ALTER TABLE tests
ADD CONSTRAINT check_passing_score CHECK (passing_score >= 0 AND passing_score <= 100);

ALTER TABLE tests
ADD CONSTRAINT check_max_attempts CHECK (max_attempts > 0);

ALTER TABLE user_progress
ADD CONSTRAINT check_current_attempt CHECK (current_attempt >= 0);

ALTER TABLE promo_codes
ADD CONSTRAINT check_remaining CHECK (remaining >= 0);

ALTER TABLE promo_codes
ADD CONSTRAINT check_max_activations CHECK (max_activations > 0);

ALTER TABLE subscriptions
ADD CONSTRAINT check_amount CHECK (amount >= 0);

-- ========== 4. CREATE INDEXES ==========

-- users
CREATE INDEX idx_users_pending_payment ON users(pending_payment_id);
CREATE INDEX idx_users_last_activity ON users(last_activity_at DESC);

-- promo_codes
CREATE INDEX idx_promo_codes_status ON promo_codes(status);
CREATE INDEX idx_promo_codes_teacher ON promo_codes(teacher_id);
CREATE INDEX idx_promo_codes_expires ON promo_codes(expires_at);
CREATE INDEX idx_promo_codes_code_status ON promo_codes(code, status);

-- tests
CREATE INDEX idx_tests_theme ON tests(theme_id);
CREATE INDEX idx_tests_difficulty ON tests(difficulty);
CREATE INDEX idx_tests_theme_difficulty ON tests(theme_id, difficulty);

-- user_progress
CREATE INDEX idx_user_progress_user ON user_progress(user_id, completed_at DESC);
CREATE INDEX idx_user_progress_theme ON user_progress(theme_id, status);
CREATE INDEX idx_user_progress_attempts ON user_progress(current_attempt);
CREATE INDEX idx_user_progress_status_updated ON user_progress(status, updated_at DESC);

-- subscriptions
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_expires ON subscriptions(expires_at);
CREATE INDEX idx_subscriptions_payment ON subscriptions(payment_id);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status, expires_at);

-- teacher_promo_students
CREATE INDEX idx_teacher_promo_teacher ON teacher_promo_students(teacher_id);
CREATE INDEX idx_teacher_promo_student ON teacher_promo_students(student_id);
CREATE INDEX idx_teacher_promo_code ON teacher_promo_students(promo_code);

-- modules and themes
CREATE INDEX idx_modules_locked ON modules(is_locked);
CREATE INDEX idx_themes_module ON themes(module_id, order_num);
CREATE INDEX idx_themes_locked ON themes(is_locked);
CREATE INDEX idx_themes_introduction ON themes(is_introduction) WHERE is_introduction = true; -- v2.1

-- mnemonics
CREATE INDEX idx_mnemonics_theme_order ON mnemonics(theme_id, order_num);

-- test_attempts
CREATE INDEX idx_test_attempts_user_theme ON test_attempts(user_id, theme_id, submitted_at DESC);
CREATE INDEX idx_test_attempts_test ON test_attempts(test_id, passed);
CREATE INDEX idx_test_attempts_user ON test_attempts(user_id, submitted_at DESC);

-- notifications
CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_status ON notifications(status, created_at);

-- audit_log
CREATE INDEX idx_audit_admin ON audit_log(admin_id, created_at DESC);
CREATE INDEX idx_audit_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_action ON audit_log(action, created_at DESC);

COMMIT;
```

---

### 4.2 Rollback скрипт

```sql
-- ==========================================
-- Rollback Script: Revert all improvements
-- ==========================================

BEGIN TRANSACTION;

-- Drop new tables
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS test_attempts;

-- Drop indexes
-- (list all created indexes here)

-- Drop constraints
-- (list all added constraints here)

-- Drop columns
ALTER TABLE themes DROP COLUMN estimated_time_minutes;
ALTER TABLE themes DROP COLUMN is_locked;
ALTER TABLE modules DROP COLUMN icon_emoji;
ALTER TABLE modules DROP COLUMN is_locked;
ALTER TABLE teacher_promo_students DROP COLUMN promo_code;
ALTER TABLE subscriptions DROP COLUMN cancellation_reason;
ALTER TABLE subscriptions DROP COLUMN cancelled_at;
ALTER TABLE subscriptions DROP COLUMN auto_renew;
ALTER TABLE user_progress DROP COLUMN last_viewed_at;
ALTER TABLE user_progress DROP COLUMN time_spent_seconds;
ALTER TABLE user_progress DROP COLUMN started_at;
ALTER TABLE user_progress DROP COLUMN test_started_at;
ALTER TABLE user_progress DROP COLUMN current_attempt;
ALTER TABLE tests DROP COLUMN shuffle_answers;
ALTER TABLE tests DROP COLUMN shuffle_questions;
ALTER TABLE tests DROP COLUMN time_limit_minutes;
ALTER TABLE tests DROP COLUMN max_attempts;
ALTER TABLE promo_codes DROP COLUMN created_by_admin_id;
ALTER TABLE promo_codes DROP COLUMN status;
ALTER TABLE promo_codes DROP COLUMN expires_at;
ALTER TABLE users DROP COLUMN last_activity_at;
ALTER TABLE users DROP COLUMN notifications_enabled;
ALTER TABLE users DROP COLUMN timezone;
ALTER TABLE users DROP COLUMN language;
ALTER TABLE users DROP COLUMN pending_payment_id;

COMMIT;
```

---

## 5. Seed данные для тестирования

```sql
-- ==========================================
-- Seed Data: Test data for development
-- ==========================================

BEGIN TRANSACTION;

-- 1. Test admin user
INSERT INTO users (telegram_id, role, subscription_status, university_code, created_at)
VALUES (999999999, 'teacher', 'active', 'ADMIN', NOW());

-- 2. Test promo codes
INSERT INTO promo_codes (code, university_name, teacher_id, max_activations, remaining, expires_at, status, created_at)
VALUES
    ('TEST2025', 'Test University', NULL, 100, 100, DATE_ADD(NOW(), INTERVAL 1 YEAR), 'pending', NOW()),
    ('DEMO2025', 'Demo University', NULL, 50, 50, DATE_ADD(NOW(), INTERVAL 6 MONTH), 'pending', NOW());

-- 3. Test modules
INSERT INTO modules (name, order_num, description, is_locked, icon_emoji, created_at)
VALUES
    ('Анатомия костей', 1, 'Изучение костной системы человека', false, '🦴', NOW()),
    ('Мышечная система', 2, 'Изучение мышечной системы', false, '💪', NOW()),
    ('Сердечно-сосудистая система', 3, 'Изучение сердца и сосудов', false, '❤️', NOW());

-- 4. Test themes for module 1
INSERT INTO themes (module_id, name, order_num, description, is_locked, estimated_time_minutes, created_at)
VALUES
    (1, 'Строение черепа', 1, 'Кости черепа и их соединения', false, 30, NOW()),
    (1, 'Позвоночник', 2, 'Строение позвоночного столба', false, 45, NOW()),
    (1, 'Грудная клетка', 3, 'Рёбра и грудина', false, 30, NOW());

-- 5. Test mnemonics
INSERT INTO mnemonics (theme_id, type, content_text, s3_image_key, order_num, created_at)
VALUES
    (1, 'text', 'Мнемоника для запоминания костей черепа...', NULL, 1, NOW()),
    (1, 'image', 'Схема строения черепа', 'mnemonics/skull_diagram.jpg', 2, NOW());

-- 6. Test tests
INSERT INTO tests (theme_id, questions_json, difficulty, passing_score, max_attempts, time_limit_minutes, created_at)
VALUES
    (1, '{"questions": [{"id": 1, "question": "Сколько костей в черепе человека?", "type": "multiple_choice", "options": ["22", "25", "29", "32"], "correct_answer": "29"}]}', 2, 70, 3, 30, NOW());

COMMIT;
```

---

**Дата создания**: 2025-12-28
**Версия**: 1.0
**Статус**: Готово к применению
