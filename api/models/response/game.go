package response

import (
	"time"
)

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
	AvailableLetters string       `json:"available_letters,omitempty"`
	WinnerUsername   string       `json:"winner_username,omitempty"`
	EndedAt          *time.Time   `json:"ended_at,omitempty"`
	IsYourGame       bool         `json:"is_your_game"`
	BlankTiles       []BoardBlank `json:"blank_tiles,omitempty"`
	PassCount        int          `json:"pass_count"`
}

type PlayerInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	Position int    `json:"position"`
	Rack     string `json:"rack,omitempty"`
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
	WinnerUsername      string    `json:"winner_username,omitempty"`
}

type GamesListResponse struct {
	Games []GameSummary `json:"games"`
}

type BoardBlank struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// AdminGameSummary is a compact summary for admin listing of games
type AdminGameSummary struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	CurrentTurnUserID   *int64     `json:"current_turn_user_id,omitempty"`
	CurrentTurnUsername string     `json:"current_turn_username,omitempty"`
	WinnerUsername      string     `json:"winner_username,omitempty"`
	EndedAt             *time.Time `json:"ended_at,omitempty"`
	PassCount           int        `json:"pass_count"`
	PlayersCount        int        `json:"players_count"`
	MovesCount          int        `json:"moves_count"`
	LastPlayTime        time.Time  `json:"last_play_time"`
	CreatedByUsername   string     `json:"created_by_username,omitempty"`
}

type AdminGamesListResponse struct {
	Games []AdminGameSummary `json:"games"`
}
