package response

import "time"

type MeResponse struct {
	ID                   int64     `json:"id"`
	Username             string    `json:"username"`
	Role                 string    `json:"role"`
	CreatedAt            time.Time `json:"created_at"`
	GamesCount           int       `json:"games_count"`
	BestScore            int       `json:"best_score"`
	Victories            int       `json:"victories"`
	AvgScore             float64   `json:"avg_score"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
}
