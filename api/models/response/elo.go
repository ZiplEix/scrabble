package response

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Games    int    `json:"games"`
}

type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
	Total   int                `json:"total"`
}

type UserStatsResponse struct {
	UserID   int64   `json:"user_id"`
	Username string  `json:"username"`
	Rating   int     `json:"rating"`
	Games    int     `json:"games"`
	Wins     int     `json:"wins"`
	Losses   int     `json:"losses"`
	WinRate  float64 `json:"win_rate"`
}
