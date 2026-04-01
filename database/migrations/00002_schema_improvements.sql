-- +goose Up

ALTER TABLE modules   ADD COLUMN updated_at DATETIME;
ALTER TABLE themes    ADD COLUMN updated_at DATETIME;
ALTER TABLE mnemonics ADD COLUMN updated_at DATETIME;
ALTER TABLE tests     ADD COLUMN updated_at DATETIME;

CREATE INDEX idx_tests_theme_id       ON tests(theme_id);
CREATE INDEX idx_notifications_user   ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status) WHERE status = 'pending';

-- +goose Down

DROP INDEX IF EXISTS idx_tests_theme_id;
DROP INDEX IF EXISTS idx_notifications_user;
DROP INDEX IF EXISTS idx_notifications_status;
-- SQLite does not support DROP COLUMN for older versions; column removal is not implemented here.
