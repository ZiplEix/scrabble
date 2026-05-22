package services

import (
	"database/sql"
	"math"
	"time"
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

// UpdateUserIPS récupère les 10 dernières parties et met à jour l'IPS de l'utilisateur, et ajoute une entrée d'historique.
// Si la partie implique le bot (partie non classée), l'IPS n'est pas mis à jour.
func UpdateUserIPS(tx *sql.Tx, userID int64, gameID string) error {
	// Les parties contre le bot sont hors classement
	if gameID != "" && IsBotGame(gameID) {
		return nil
	}

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
	if err != nil {
		return err
	}

	if gameID != "" {
		_, err = tx.Exec(`
			INSERT INTO user_rating_history (user_id, game_id, rating, created_at)
			VALUES ($1, $2, $3, now())
		`, userID, gameID, ips)
	} else {
		_, err = tx.Exec(`
			INSERT INTO user_rating_history (user_id, rating, created_at)
			VALUES ($1, $2, now())
		`, userID, ips)
	}
	return err
}


// RegenerateUserRatingHistory reconstruit l'historique complet de l'IPS pour un joueur chronologiquement
func RegenerateUserRatingHistory(tx *sql.Tx, userID int64) error {
	// 1. Supprimer l'historique existant
	_, err := tx.Exec(`DELETE FROM user_rating_history WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// 2. Récupérer toutes les parties terminées par ce joueur chronologiquement
	var username string
	err = tx.QueryRow(`SELECT username FROM users WHERE id = $1`, userID).Scan(&username)
	if err != nil {
		return err
	}

	rows, err := tx.Query(`
		SELECT g.id, g.ended_at
		FROM game_players gp
		JOIN games g ON gp.game_id = g.id
		WHERE gp.player_id = $1 AND g.status = 'ended'
		ORDER BY g.ended_at ASC
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type gamePoint struct {
		id      string
		endedAt time.Time
	}
	var games []gamePoint
	for rows.Next() {
		var gp gamePoint
		if err := rows.Scan(&gp.id, &gp.endedAt); err != nil {
			return err
		}
		games = append(games, gp)
	}

	// 3. Pour chaque partie, calculer l'IPS à cet instant précis (les 10 dernières parties se terminant à ou avant cette date)
	for _, gp := range games {
		mRows, err := tx.Query(`
			SELECT 
				gp2.score, 
				(g2.winner_username = $1) as is_winner
			FROM game_players gp2
			JOIN games g2 ON gp2.game_id = g2.id
			WHERE gp2.player_id = $2 AND g2.status = 'ended' AND g2.ended_at <= $3
			ORDER BY g2.ended_at DESC
			LIMIT 10
		`, username, userID, gp.endedAt)
		if err != nil {
			return err
		}

		var matches []MatchInfo
		for mRows.Next() {
			var m MatchInfo
			if err := mRows.Scan(&m.Score, &m.IsWinner); err != nil {
				mRows.Close()
				return err
			}
			matches = append(matches, m)
		}
		mRows.Close()

		ipsVal := CalculateIPS(matches)

		// 4. Insérer dans user_rating_history avec ended_at comme created_at
		_, err = tx.Exec(`
			INSERT INTO user_rating_history (user_id, game_id, rating, created_at)
			VALUES ($1, $2, $3, $4)
		`, userID, gp.id, ipsVal, gp.endedAt)
		if err != nil {
			return err
		}
	}

	return nil
}
