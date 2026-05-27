-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_rating_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id UUID REFERENCES games(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_rating_history_user ON user_rating_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_rating_history_created ON user_rating_history(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_rating_history;
-- +goose StatementEnd
