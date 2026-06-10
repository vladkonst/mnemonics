-- +goose Up

CREATE TABLE IF NOT EXISTS corporate_purchases (
    payment_id    TEXT    PRIMARY KEY,
    manager_id    BIGINT  NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    groups_count  INTEGER NOT NULL,
    semesters     INTEGER NOT NULL CHECK(semesters BETWEEN 1 AND 3),
    total_amount  INTEGER NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_corp_purchases_manager ON corporate_purchases(manager_id);

CREATE TABLE IF NOT EXISTS corporate_groups (
    id               TEXT    PRIMARY KEY,
    name             TEXT    NOT NULL,
    purchase_id      TEXT    NOT NULL REFERENCES corporate_purchases(payment_id) ON DELETE CASCADE,
    manager_id       BIGINT  NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    teacher_id       BIGINT  REFERENCES users(telegram_id) ON DELETE SET NULL,
    teacher_join_code TEXT   NOT NULL UNIQUE,
    student_link_id  TEXT    NOT NULL UNIQUE REFERENCES invite_links(id) ON DELETE RESTRICT,
    semesters        INTEGER NOT NULL CHECK(semesters BETWEEN 1 AND 3),
    created_at       DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_corp_groups_purchase  ON corporate_groups(purchase_id);
CREATE INDEX idx_corp_groups_manager   ON corporate_groups(manager_id);
CREATE INDEX idx_corp_groups_teacher   ON corporate_groups(teacher_id) WHERE teacher_id IS NOT NULL;
CREATE INDEX idx_corp_groups_join_code ON corporate_groups(teacher_join_code);

-- +goose Down
DROP TABLE IF EXISTS corporate_groups;
DROP TABLE IF EXISTS corporate_purchases;
