-- +goose Up
-- +goose StatementBegin

-- Recreate puzzle_attempts with started_at column and nullable fields for "started but not submitted" state
ALTER TABLE puzzle_attempts 
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMP,
    ALTER COLUMN score DROP NOT NULL,
    ALTER COLUMN words_played DROP NOT NULL,
    ALTER COLUMN submitted_at DROP NOT NULL;

-- Set started_at = created_at for existing rows
UPDATE puzzle_attempts SET started_at = created_at WHERE started_at IS NULL;

-- Make started_at NOT NULL after backfill
ALTER TABLE puzzle_attempts ALTER COLUMN started_at SET NOT NULL;

-- Add index for lookups by player+puzzle for start checking
CREATE INDEX IF NOT EXISTS idx_puzzle_attempts_started_at ON puzzle_attempts(puzzle_id, player_id, started_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_puzzle_attempts_started_at;

-- Backfill nullable fields introduced by this migration before restoring NOT NULL constraints.
UPDATE puzzle_attempts
SET score = 0
WHERE score IS NULL;

UPDATE puzzle_attempts
SET words_played = '[]'::jsonb
WHERE words_played IS NULL;

UPDATE puzzle_attempts
SET submitted_at = COALESCE(submitted_at, started_at, created_at, now())
WHERE submitted_at IS NULL;

ALTER TABLE puzzle_attempts
    ALTER COLUMN score SET NOT NULL,
    ALTER COLUMN words_played SET NOT NULL,
    ALTER COLUMN submitted_at SET NOT NULL;

ALTER TABLE puzzle_attempts
    DROP COLUMN IF EXISTS started_at;
-- +goose StatementEnd
