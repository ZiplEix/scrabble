package response

import "time"

type GameInfo struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Board           any          `json:"board"`
	YourRack        string       `json:"your_rack"`
	Players         []PlayerInfo `json:"players"`
	Moves           []MoveInfo   `json:"moves"`
	CurrentTurn     int64        `json:"current_turn"`
	CurrentTurnName string       `json:"current_turn_username"`
	Status          string       `json:"status"`
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
