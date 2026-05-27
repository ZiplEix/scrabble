-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS game_message_reads (
    user_id BIGINT NOT NULL,
    game_id TEXT NOT NULL,
    last_read_message_id BIGINT,
    last_read_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_gmr_game_id ON game_message_reads (game_id);
CREATE INDEX IF NOT EXISTS idx_gmr_user_id ON game_message_reads (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_message_reads;
DROP INDEX IF EXISTS idx_gmr_game_id;
DROP INDEX IF EXISTS idx_gmr_user_id;
-- +goose StatementEnd
