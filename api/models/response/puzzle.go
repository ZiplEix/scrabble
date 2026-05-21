package response

import (
	"time"
)

type PuzzleInfo struct {
	ID                 string    `json:"id"`
	PuzzleDate         string    `json:"puzzle_date"` // Format: YYYY-MM-DD
	Level              int       `json:"level"`
	Board              any       `json:"board"`
	AvailableLetters   string    `json:"available_letters"`
	TimeoutSeconds     int       `json:"timeout_seconds"`
	HasPlayerAttempted bool      `json:"has_player_attempted"`
	CreatedAt          time.Time `json:"created_at"`
}

type PuzzleAttempt struct {
	ID           string             `json:"id"`
	PuzzleID     string             `json:"puzzle_id"`
	PlayerID     int64              `json:"player_id"`
	StartedAt    time.Time          `json:"started_at"`
	Score        int                `json:"score"`
	WordsPlayed  []PuzzleWordRecord `json:"words_played"`
	TimeUsedSecs int                `json:"time_used_secs"` // computed server-side
	RankToday    int                `json:"rank_today"`
	SubmittedAt  *time.Time         `json:"submitted_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

type PuzzleStarted struct {
	AttemptID      string    `json:"attempt_id"`
	StartedAt      time.Time `json:"started_at"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	AlreadyStarted bool      `json:"already_started"`
}

type PuzzleWordRecord struct {
	Word      string `json:"word"`
	Position  string `json:"position"`  // e.g., "7,7"
	Direction string `json:"direction"` // "horizontal" or "vertical"
	Score     int    `json:"score"`
}

type PuzzleDailyLeaderboard struct {
	Rank        int       `json:"rank"`
	PlayerID    int64     `json:"player_id"`
	Username    string    `json:"username"`
	Score       int       `json:"score"`
	TimeUsed    int       `json:"time_used"`
	Attempts    int       `json:"attempts"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type PuzzleHistory struct {
	ID             string                   `json:"id"`
	PuzzleDate     string                   `json:"puzzle_date"`
	Level          int                      `json:"level"`
	HasAttempted   bool                     `json:"has_attempted"`
	PlayerAttempt  *PuzzleAttempt           `json:"player_attempt,omitempty"`
	DayLeaderboard []PuzzleDailyLeaderboard `json:"day_leaderboard,omitempty"`
}

type PuzzleStats struct {
	TotalAttempts    int `json:"total_attempts"`
	BestScore        int `json:"best_score"`
	AverageScore     int `json:"average_score"`
	CompletedPuzzles int `json:"completed_puzzles"`
}
