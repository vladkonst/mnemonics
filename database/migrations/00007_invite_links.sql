-- +goose Up
CREATE TABLE IF NOT EXISTS invite_links (
    id         TEXT    PRIMARY KEY,
    teacher_id INTEGER NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    promo_code TEXT    NOT NULL REFERENCES promo_codes(code)  ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS invite_link_activations (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    invite_link_id TEXT    NOT NULL REFERENCES invite_links(id) ON DELETE CASCADE,
    user_id        INTEGER NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    activated_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE (invite_link_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS invite_link_activations;
DROP TABLE IF EXISTS invite_links;
