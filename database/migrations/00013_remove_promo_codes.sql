-- +goose Up
-- +goose NO TRANSACTION
-- Remove promo_codes entity. Quota tracking moves into invite_links.max_activations.
-- teacher_promo_students.promo_code renamed to join_ref (drops FK to promo_codes).

PRAGMA foreign_keys = OFF;

-- 1. Recreate invite_links: drop promo_code FK, add max_activations.
CREATE TABLE invite_links_new (
    id              TEXT    PRIMARY KEY,
    teacher_id      INTEGER NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    max_activations INTEGER NOT NULL DEFAULT 30,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now'))
);
INSERT INTO invite_links_new (id, teacher_id, max_activations, created_at)
    SELECT id, teacher_id, 30, created_at FROM invite_links;
DROP TABLE invite_links;
ALTER TABLE invite_links_new RENAME TO invite_links;

-- 2. Recreate teacher_promo_students: rename promo_code → join_ref, drop FK.
CREATE TABLE teacher_promo_students_new (
    teacher_id  BIGINT   NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    student_id  BIGINT   NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    join_ref    TEXT,
    joined_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (teacher_id, student_id)
);
INSERT INTO teacher_promo_students_new (teacher_id, student_id, join_ref, joined_at)
    SELECT teacher_id, student_id, promo_code, joined_at FROM teacher_promo_students;
DROP TABLE teacher_promo_students;
ALTER TABLE teacher_promo_students_new RENAME TO teacher_promo_students;

-- 3. Drop promo_codes (invite_links no longer references it).
DROP TABLE IF EXISTS promo_codes;

PRAGMA foreign_keys = ON;

-- +goose Down
-- +goose NO TRANSACTION
PRAGMA foreign_keys = OFF;

-- Restore promo_codes (empty).
CREATE TABLE IF NOT EXISTS promo_codes (
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

-- Restore invite_links with promo_code column.
CREATE TABLE invite_links_old (
    id         TEXT    PRIMARY KEY,
    teacher_id INTEGER NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    promo_code TEXT    NOT NULL DEFAULT '' REFERENCES promo_codes(code) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
INSERT INTO invite_links_old (id, teacher_id, promo_code, created_at)
    SELECT id, teacher_id, '', created_at FROM invite_links;
DROP TABLE invite_links;
ALTER TABLE invite_links_old RENAME TO invite_links;

-- Restore teacher_promo_students with promo_code column.
CREATE TABLE teacher_promo_students_old (
    teacher_id  BIGINT   NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    student_id  BIGINT   NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    promo_code  TEXT     REFERENCES promo_codes(code),
    joined_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (teacher_id, student_id)
);
INSERT INTO teacher_promo_students_old (teacher_id, student_id, promo_code, joined_at)
    SELECT teacher_id, student_id, join_ref, joined_at FROM teacher_promo_students;
DROP TABLE teacher_promo_students;
ALTER TABLE teacher_promo_students_old RENAME TO teacher_promo_students;

PRAGMA foreign_keys = ON;
