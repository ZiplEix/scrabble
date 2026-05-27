-- +goose Up
-- +goose StatementBegin
ALTER TABLE games
    ADD COLUMN IF NOT EXISTS pass_count INT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games
    DROP COLUMN IF EXISTS pass_count;
-- +goose StatementEnd
