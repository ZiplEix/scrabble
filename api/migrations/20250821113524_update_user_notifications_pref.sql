-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN notification_prefs JSONB NOT NULL DEFAULT '{"turn": true, "messages": true}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS notification_prefs;
-- +goose StatementEnd
