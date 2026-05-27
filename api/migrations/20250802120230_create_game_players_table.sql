-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS game_players (
    game_id UUID REFERENCES games(id),
    player_id INT REFERENCES users(id),
    rack TEXT NOT NULL,
    position INT NOT NULL,
    score INT NOT NULL DEFAULT 0,
    PRIMARY KEY (game_id, player_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_players;
-- +goose StatementEnd
