-- +goose Up
-- +goose StatementBegin

CREATE TABLE users (
    telegram_id             BIGINT PRIMARY KEY,
    role                    TEXT NOT NULL DEFAULT 'student'
                                CHECK(role IN ('student','teacher')),
    subscription_status     TEXT NOT NULL DEFAULT 'inactive'
                                CHECK(subscription_status IN ('active','inactive','expired')),
    university_code         TEXT,
    pending_payment_id      TEXT,
    first_name              TEXT NOT NULL,
    last_name               TEXT,
    username                TEXT,
    language                TEXT NOT NULL DEFAULT 'ru',
    timezone                TEXT NOT NULL DEFAULT 'UTC',
    notifications_enabled   INTEGER NOT NULL DEFAULT 1,
    last_activity_at        DATETIME,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    module_id               INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    name                    TEXT NOT NULL,
    description             TEXT,
    order_num               INTEGER NOT NULL,
    is_introduction         INTEGER NOT NULL DEFAULT 0,
    is_locked               INTEGER NOT NULL DEFAULT 0,
    estimated_time_minutes  INTEGER,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mnemonics (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    theme_id     INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    type         TEXT NOT NULL CHECK(type IN ('text','image')),
    content_text TEXT,
    s3_image_key TEXT,
    order_num    INTEGER NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tests (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    theme_id          INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    questions_json    TEXT NOT NULL,
    difficulty        INTEGER NOT NULL DEFAULT 1,
    passing_score     INTEGER NOT NULL DEFAULT 70
                          CHECK(passing_score >= 0 AND passing_score <= 100),
    shuffle_questions INTEGER NOT NULL DEFAULT 1,
    shuffle_answers   INTEGER NOT NULL DEFAULT 1,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE promo_codes (
    code                TEXT PRIMARY KEY,
    university_name     TEXT NOT NULL,
    teacher_id          BIGINT REFERENCES users(telegram_id) ON DELETE SET NULL,
    max_activations     INTEGER NOT NULL CHECK(max_activations > 0),
    remaining           INTEGER NOT NULL CHECK(remaining >= 0),
    status              TEXT NOT NULL DEFAULT 'pending'
                            CHECK(status IN ('pending','active','expired','deactivated')),
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
    status              TEXT NOT NULL DEFAULT 'active'
                            CHECK(status IN ('active','expired','cancelled')),
    plan                TEXT,
    expires_at          DATETIME,
    auto_renew          INTEGER NOT NULL DEFAULT 0,
    cancelled_at        DATETIME,
    cancellation_reason TEXT,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_progress (
    user_id             BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    theme_id            INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    status              TEXT NOT NULL DEFAULT 'started'
                            CHECK(status IN ('started','completed','failed')),
    score               INTEGER CHECK(score >= 0 AND score <= 100),
    current_attempt     INTEGER NOT NULL DEFAULT 0 CHECK(current_attempt >= 0),
    test_started_at     DATETIME,
    started_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at        DATETIME,
    time_spent_seconds  INTEGER NOT NULL DEFAULT 0,
    last_viewed_at      DATETIME,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, theme_id)
);

CREATE TABLE test_attempts (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id          BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    theme_id         INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    test_id          INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    attempt_id       TEXT NOT NULL UNIQUE,
    answers_json     TEXT NOT NULL,
    score            INTEGER NOT NULL,
    passed           INTEGER NOT NULL,
    started_at       DATETIME NOT NULL,
    submitted_at     DATETIME,
    duration_seconds INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE notifications (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    type          TEXT NOT NULL,
    title         TEXT NOT NULL,
    message       TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending'
                      CHECK(status IN ('pending','sent','failed')),
    sent_at       DATETIME,
    error_message TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id        BIGINT NOT NULL,
    admin_username  TEXT,
    action          TEXT NOT NULL,
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    old_value_json  TEXT,
    new_value_json  TEXT,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_themes_module_order       ON themes(module_id, order_num);
CREATE INDEX idx_themes_introduction       ON themes(is_introduction) WHERE is_introduction = 1;
CREATE INDEX idx_user_progress_user        ON user_progress(user_id, completed_at DESC);
CREATE INDEX idx_user_progress_theme       ON user_progress(theme_id, status);
CREATE INDEX idx_test_attempts_user_theme  ON test_attempts(user_id, theme_id, submitted_at DESC);
CREATE INDEX idx_test_attempts_attempt_id  ON test_attempts(attempt_id);
CREATE INDEX idx_promo_codes_teacher       ON promo_codes(teacher_id);
CREATE INDEX idx_promo_codes_status        ON promo_codes(status);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status, expires_at);
CREATE INDEX idx_teacher_promo_student     ON teacher_promo_students(student_id);
CREATE INDEX idx_mnemonics_theme_order     ON mnemonics(theme_id, order_num);
CREATE INDEX idx_users_pending_payment     ON users(pending_payment_id)
    WHERE pending_payment_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

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

-- +goose StatementEnd
