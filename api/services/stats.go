package services

import (
	"fmt"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
)

// GetLeaderboard retourne le top N joueurs par rating
func GetLeaderboard(limit int, offset int) (*response.LeaderboardResponse, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Comptage total
	var total int
	if err := database.QueryRow(`SELECT COUNT(*) FROM users WHERE rating > 0`).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Récupération du leaderboard avec nombre de parties
	rows, err := database.Query(`
		SELECT 
			u.id,
			u.username,
			u.rating,
			COUNT(DISTINCT CASE WHEN gp.game_id IS NOT NULL THEN gp.game_id END) as games
		FROM users u
		LEFT JOIN game_players gp ON u.id = gp.player_id
		WHERE u.rating > 0
		GROUP BY u.id, u.username, u.rating
		ORDER BY u.rating DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []response.LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		var userID int64
		var username string
		var rating int
		var games int

		if err := rows.Scan(&userID, &username, &rating, &games); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}

		entries = append(entries, response.LeaderboardEntry{
			Rank:     rank,
			UserID:   userID,
			Username: username,
			Rating:   rating,
			Games:    games,
		})
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard: %w", err)
	}

	return &response.LeaderboardResponse{
		Entries: entries,
		Total:   total,
	}, nil
}

// GetUserStats retourne les stats complètes d'un joueur
func GetUserStats(userID int64) (*response.UserStatsResponse, error) {
	var username string
	var rating int
	var games int
	var wins int

	err := database.QueryRow(`
		SELECT 
			u.username,
			u.rating,
			COUNT(DISTINCT gp.game_id) as games,
			SUM(CASE WHEN g.winner_username = u.username THEN 1 ELSE 0 END) as wins
		FROM users u
		LEFT JOIN game_players gp ON u.id = gp.player_id
		LEFT JOIN games g ON gp.game_id = g.id
		WHERE u.id = $1
		GROUP BY u.id, u.username, u.rating
	`, userID).Scan(&username, &rating, &games, &wins)

	if err != nil {
		return nil, fmt.Errorf("failed to query user stats: %w", err)
	}

	losses := games - wins
	winRate := 0.0
	if games > 0 {
		winRate = float64(wins) / float64(games) * 100
	}

	return &response.UserStatsResponse{
		UserID:   userID,
		Username: username,
		Rating:   rating,
		Games:    games,
		Wins:     wins,
		Losses:   losses,
		WinRate:  winRate,
	}, nil
}
