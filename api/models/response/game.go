package response

import "time"

type GameInfo struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Board            any          `json:"board"`
	YourRack         string       `json:"your_rack"`
	Players          []PlayerInfo `json:"players"`
	Moves            []MoveInfo   `json:"moves"`
	CurrentTurn      int64        `json:"current_turn"`
	CurrentTurnName  string       `json:"current_turn_username"`
	Status           string       `json:"status"`
	RemainingLetters int          `json:"remaining_letters"`
	WinnerUsername   string       `json:"winner_username,omitempty"`
	EndedAt          *time.Time   `json:"ended_at,omitempty"`
}

type PlayerInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	Position int    `json:"position"`
}

type MoveInfo struct {
	PlayerID int64     `json:"player_id"`
	Move     any       `json:"move"` // JSONB brut
	PlayedAt time.Time `json:"played_at"`
}

type GameSummary struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Status              string    `json:"status"`
	CurrentTurnUserID   int       `json:"current_turn_user_id"`
	CurrentTurnUsername string    `json:"current_turn_username"`
	LastPlayTime        time.Time `json:"last_play_time"`
	IsYourGame          bool      `json:"is_your_game"`
}

type GamesListResponse struct {
	Games []GameSummary `json:"games"`
}
