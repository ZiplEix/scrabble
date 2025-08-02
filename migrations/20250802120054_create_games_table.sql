-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_by INT REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'ongoing',
    current_turn INT REFERENCES users(id),
    board JSONB NOT NULL,
    available_letters TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS games;
-- +goose StatementEnd
