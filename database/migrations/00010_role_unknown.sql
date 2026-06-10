-- +goose Up
CREATE TABLE users_new (
    telegram_id             BIGINT PRIMARY KEY,
    role                    TEXT NOT NULL DEFAULT 'unknown'
                                CHECK(role IN ('unknown','student','teacher')),
    subscription_status     TEXT NOT NULL DEFAULT 'inactive'
                                CHECK(subscription_status IN ('active','inactive','expired')),
    university_code         TEXT,
    pending_payment_id      TEXT,
    username                TEXT,
    language                TEXT NOT NULL DEFAULT 'ru',
    timezone                TEXT NOT NULL DEFAULT 'UTC',
    notifications_enabled   INTEGER NOT NULL DEFAULT 1,
    last_activity_at        DATETIME,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_new SELECT * FROM users;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
CREATE INDEX idx_users_pending_payment ON users(pending_payment_id) WHERE pending_payment_id IS NOT NULL;

-- +goose Down
CREATE TABLE users_old (
    telegram_id             BIGINT PRIMARY KEY,
    role                    TEXT NOT NULL DEFAULT 'student'
                                CHECK(role IN ('student','teacher')),
    subscription_status     TEXT NOT NULL DEFAULT 'inactive'
                                CHECK(subscription_status IN ('active','inactive','expired')),
    university_code         TEXT,
    pending_payment_id      TEXT,
    username                TEXT,
    language                TEXT NOT NULL DEFAULT 'ru',
    timezone                TEXT NOT NULL DEFAULT 'UTC',
    notifications_enabled   INTEGER NOT NULL DEFAULT 1,
    last_activity_at        DATETIME,
    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_old SELECT telegram_id, CASE WHEN role='unknown' THEN 'student' ELSE role END,
    subscription_status, university_code, pending_payment_id, username, language, timezone,
    notifications_enabled, last_activity_at, created_at FROM users;
DROP TABLE users;
ALTER TABLE users_old RENAME TO users;
CREATE INDEX idx_users_pending_payment ON users(pending_payment_id) WHERE pending_payment_id IS NOT NULL;
