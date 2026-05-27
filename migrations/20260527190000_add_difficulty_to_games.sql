-- +goose Up
-- +goose StatementBegin
ALTER TABLE games ADD COLUMN difficulty VARCHAR(20) DEFAULT 'hard' CHECK (difficulty IN ('easy', 'medium', 'hard'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games DROP COLUMN IF EXISTS difficulty;
-- +goose StatementEnd
