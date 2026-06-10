-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS module_test_attempts (
    id               INTEGER  PRIMARY KEY AUTOINCREMENT,
    attempt_id       TEXT     NOT NULL UNIQUE,
    user_id          INTEGER  NOT NULL,
    module_id        INTEGER  NOT NULL,
    questions_json   TEXT     NOT NULL DEFAULT '[]',
    answers_json     TEXT     NOT NULL DEFAULT '[]',
    score            INTEGER  NOT NULL DEFAULT 0,
    passed           INTEGER  NOT NULL DEFAULT 0,
    started_at       DATETIME NOT NULL,
    submitted_at     DATETIME,
    duration_seconds INTEGER  NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mta_user_module ON module_test_attempts(user_id, module_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS module_test_attempts;
-- +goose StatementEnd
