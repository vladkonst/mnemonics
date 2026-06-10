-- +goose Up
-- +goose StatementBegin
ALTER TABLE mnemonics ADD COLUMN term_ru    TEXT;
ALTER TABLE mnemonics ADD COLUMN term_latin TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite does not support DROP COLUMN in older versions; migration is irreversible here.
-- +goose StatementEnd
