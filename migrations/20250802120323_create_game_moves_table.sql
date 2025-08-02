-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS game_moves (
    id SERIAL PRIMARY KEY,
    game_id UUID REFERENCES games(id),
    player_id INT REFERENCES users(id),
    move JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_moves;
-- +goose StatementEnd
