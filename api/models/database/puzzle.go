package database

import (
	"time"
)

type DailyPuzzle struct {
	ID               string    `db:"id"`
	PuzzleDate       time.Time `db:"puzzle_date"`
	Level            int       `db:"level"`
	Board            []byte    `db:"board"` // JSONB
	AvailableLetters string    `db:"available_letters"`
	Seed             string    `db:"seed"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type PuzzleAttempt struct {
	ID          string     `db:"id"`
	PuzzleID    string     `db:"puzzle_id"`
	PlayerID    int64      `db:"player_id"`
	StartedAt   time.Time  `db:"started_at"`
	Score       *int       `db:"score"`        // NULL until submitted
	WordsPlayed []byte     `db:"words_played"` // JSONB, NULL until submitted
	TimeUsed    *int       `db:"time_used"`    // computed server-side at submit
	SubmittedAt *time.Time `db:"submitted_at"` // NULL until submitted
	CreatedAt   time.Time  `db:"created_at"`
}

type PuzzleAttemptWithRank struct {
	Attempt *PuzzleAttempt
	Rank    int
}
