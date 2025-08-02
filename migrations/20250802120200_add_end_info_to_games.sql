-- +goose Up
-- +goose StatementBegin
ALTER TABLE games
    ADD COLUMN IF NOT EXISTS winner_username TEXT,
    ADD COLUMN IF NOT EXISTS ended_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games
    DROP COLUMN IF EXISTS ended_at,
    DROP COLUMN IF EXISTS winner_username;
-- +goose StatementEnd
