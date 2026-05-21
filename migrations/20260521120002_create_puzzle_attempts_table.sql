-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS puzzle_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    puzzle_id UUID NOT NULL REFERENCES daily_puzzles(id) ON DELETE CASCADE,
    player_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score INT NOT NULL DEFAULT 0,
    words_played JSONB NOT NULL DEFAULT '[]'::jsonb,
    time_used INT NOT NULL DEFAULT 0,
    submitted_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(puzzle_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_puzzle_attempts_puzzle_id ON puzzle_attempts(puzzle_id);
CREATE INDEX IF NOT EXISTS idx_puzzle_attempts_player_id ON puzzle_attempts(player_id);
CREATE INDEX IF NOT EXISTS idx_puzzle_attempts_score ON puzzle_attempts(puzzle_id, score DESC);
CREATE INDEX IF NOT EXISTS idx_puzzle_attempts_submitted_at ON puzzle_attempts(submitted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS puzzle_attempts;
-- +goose StatementEnd
