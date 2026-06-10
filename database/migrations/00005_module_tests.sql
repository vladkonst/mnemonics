-- +goose Up
-- +goose StatementBegin

CREATE TABLE module_tests (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    module_id         INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    questions_json    TEXT NOT NULL DEFAULT '[]',
    difficulty        INTEGER NOT NULL DEFAULT 1,
    passing_score     INTEGER NOT NULL DEFAULT 70
                          CHECK(passing_score >= 0 AND passing_score <= 100),
    shuffle_questions INTEGER NOT NULL DEFAULT 1,
    shuffle_answers   INTEGER NOT NULL DEFAULT 1,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_module_tests_module ON module_tests(module_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS module_tests;

-- +goose StatementEnd
