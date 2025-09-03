package response

import "time"

type SuggestUsersResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// UserPublicResponse exposes public info about a user
type UserPublicResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	// Stats
	GamesCount                 int      `json:"games_count"`
	GamesCountTopPercent       *float64 `json:"games_count_top_percent,omitempty"`
	BestScore                  int      `json:"best_score"`
	BestScoreTopPercent        *float64 `json:"best_score_top_percent,omitempty"`
	Victories                  int      `json:"victories"`
	VictoriesTopPercent        *float64 `json:"victories_top_percent,omitempty"`
	AvgScore                   float64  `json:"avg_score"`
	AvgScoreTopPercent         *float64 `json:"avg_score_top_percent,omitempty"`
	AvgPointsPerMove           float64  `json:"avg_points_per_move"`
	AvgPointsPerMoveTopPercent *float64 `json:"avg_points_per_move_top_percent,omitempty"`
	BestMoveScore              int      `json:"best_move_score"`
	BestMoveScoreTopPercent    *float64 `json:"best_move_score_top_percent,omitempty"`
}
