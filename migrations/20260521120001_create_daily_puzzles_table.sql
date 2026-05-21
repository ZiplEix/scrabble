-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS daily_puzzles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    puzzle_date DATE NOT NULL UNIQUE,
    level INT NOT NULL DEFAULT 1,
    board JSONB NOT NULL,
    available_letters TEXT NOT NULL,
    seed TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_daily_puzzles_date ON daily_puzzles(puzzle_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS daily_puzzles;
-- +goose StatementEnd
