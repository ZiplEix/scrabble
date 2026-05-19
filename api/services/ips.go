package services

import (
	"database/sql"
	"math"
)

// MatchInfo contient les données minimales d'une partie pour le calcul de l'IPS
type MatchInfo struct {
	Score    int
	IsWinner bool
}

// CalculateIPS implémente la formule de l'Indice de Performance Scrabble
func CalculateIPS(recentMatches []MatchInfo) int {
	n := len(recentMatches)
	if n == 0 {
		return 0
	}

	totalScore := 0
	wins := 0
	for _, match := range recentMatches {
		totalScore += match.Score
		if match.IsWinner {
			wins++
		}
	}

	averageScore := float64(totalScore) / float64(n)
	bonus := float64(wins * 15)

	return int(math.Round(averageScore + bonus))
}

// UpdateUserIPS récupère les 10 dernières parties et met à jour l'IPS de l'utilisateur
func UpdateUserIPS(tx *sql.Tx, userID int64) error {
	rows, err := tx.Query(`
		SELECT 
			gp.score, 
			(g.winner_username = u.username) as is_winner
		FROM game_players gp
		JOIN games g ON gp.game_id = g.id
		JOIN users u ON gp.player_id = u.id
		WHERE gp.player_id = $1 AND g.status = 'ended'
		ORDER BY g.ended_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var matches []MatchInfo
	for rows.Next() {
		var m MatchInfo
		if err := rows.Scan(&m.Score, &m.IsWinner); err != nil {
			return err
		}
		matches = append(matches, m)
	}

	ips := CalculateIPS(matches)

	// Note: On utilise pour le moment la colonne 'rating' en attendant la migration vers 'current_ips'
	_, err = tx.Exec(`UPDATE users SET rating = $1 WHERE id = $2`, ips, userID)
	return err
}
