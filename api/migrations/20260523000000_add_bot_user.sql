-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_bot BOOLEAN NOT NULL DEFAULT FALSE;

INSERT INTO users (username, password, role, is_bot, created_at)
VALUES ('Scrabby', '', 'ordinateur', TRUE, now())
ON CONFLICT (username) DO UPDATE SET role = 'ordinateur', is_bot = TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE username = 'Scrabby' AND is_bot = TRUE;
ALTER TABLE users DROP COLUMN IF EXISTS is_bot;
-- +goose StatementEnd
