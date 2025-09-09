package response

type AdminStatsResponse struct {
	ActiveUsersCount        int     `json:"active_users_count"`
	ActiveUsersPctChange    float64 `json:"active_users_pct_change"`
	CreatedGamesCount       int     `json:"created_games_count"`
	CreatedGamesPctChange   float64 `json:"created_games_pct_change"`
	ActiveGamesCount        int     `json:"active_games_count"`
	ActiveGamesPctChange    float64 `json:"active_games_pct_change"`
	SendedMessagesCount     int     `json:"sent_messages_count"`
	SendedMessagesPctChange float64 `json:"sent_messages_pct_change"`
	TicketsCreatedCount     int     `json:"tickets_created_count"`
	TicketsCreatedPctChange float64 `json:"tickets_created_pct_change"`
}
