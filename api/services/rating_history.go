package services

import (
	"database/sql"
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
)

// GetRatingHistoryByUserID récupère l'historique de classement d'un utilisateur enrichi avec les infos des parties
func GetRatingHistoryByUserID(userID int64, limit int) ([]response.RatingHistoryResponse, error) {
	rows, err := database.Query(`
		SELECT game_id, rating, created_at 
		FROM user_rating_history 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []response.RatingHistoryResponse
	for rows.Next() {
		var r response.RatingHistoryResponse
		var gameID sql.NullString
		if err := rows.Scan(&gameID, &r.Rating, &r.CreatedAt); err != nil {
			return nil, err
		}
		if gameID.Valid && gameID.String != "" {
			r.GameInfo = &response.RatingHistoryGameInfo{
				GameID: gameID.String,
			}
		}
		history = append(history, r)
	}

	if len(history) == 0 {
		return []response.RatingHistoryResponse{}, nil
	}

	// Récupérer le pseudo de l'utilisateur principal pour comparer avec winner_username
	var myUsername string
	_ = database.QueryRow(`SELECT username FROM users WHERE id = $1`, userID).Scan(&myUsername)

	// Pour chaque entrée liée à une partie, récupérer les scores et l'adversaire
	for i, h := range history {
		if h.GameInfo != nil {
			var endedAt sql.NullTime
			err := database.QueryRow(`
				SELECT ended_at FROM games WHERE id = $1
			`, h.GameInfo.GameID).Scan(&endedAt)
			if err == nil && endedAt.Valid {
				h.GameInfo.EndedAt = endedAt.Time
			}

			// Score du joueur principal
			var userScore int
			_ = database.QueryRow(`
				SELECT score FROM game_players WHERE game_id = $1 AND player_id = $2
			`, h.GameInfo.GameID, userID).Scan(&userScore)
			h.GameInfo.UserScore = userScore

			// Score et pseudo de l'adversaire
			var oppUsername string
			var oppScore int
			err = database.QueryRow(`
				SELECT u.username, gp.score
				FROM game_players gp
				JOIN users u ON gp.player_id = u.id
				WHERE gp.game_id = $1 AND gp.player_id != $2
				LIMIT 1
			`, h.GameInfo.GameID, userID).Scan(&oppUsername, &oppScore)
			if err == nil {
				h.GameInfo.OpponentUsername = oppUsername
				h.GameInfo.OpponentScore = oppScore
			}

			// Résultat Victoire / Défaite
			var winnerUsername sql.NullString
			_ = database.QueryRow(`
				SELECT winner_username FROM games WHERE id = $1
			`, h.GameInfo.GameID).Scan(&winnerUsername)
			h.GameInfo.Won = winnerUsername.Valid && winnerUsername.String == myUsername

			history[i] = h
		}
	}

	// Inverser l'historique pour le renvoyer par ordre chronologique (du plus ancien au plus récent)
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}
