package services

import (
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/stats"
)

func SuggestUsers(userID int64, query string) ([]response.SuggestUsersResponse, error) {
	rows, err := database.DB.Query(`
		SELECT id, username FROM users
		WHERE id != $1 AND LOWER(username) LIKE LOWER($2)
		LIMIT 10
	`, userID, query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []response.SuggestUsersResponse
	for rows.Next() {
		var u response.SuggestUsersResponse
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			continue
		}
		suggestions = append(suggestions, u)
	}

	return suggestions, nil
}

// GetUserPublicByID retourne les informations publiques d'un utilisateur
func GetUserPublicByID(userID int64) (*response.UserPublicResponse, error) {
	var u response.UserPublicResponse
	var createdAt time.Time
	err := database.QueryRow("SELECT id, username, role, created_at FROM users WHERE id = $1", userID).Scan(&u.ID, &u.Username, &u.Role, &createdAt)
	if err != nil {
		return nil, err
	}
	u.CreatedAt = createdAt
	// Populate stats using helpers
	if v, p, err := stats.GetGamesCountAndTop(userID); err == nil {
		u.GamesCount = v
		if p > 0 {
			f := float64(p)
			u.GamesCountTopPercent = &f
		}
	} else {
		return nil, err
	}

	if v, p, err := stats.GetBestScoreAndTop(userID); err == nil {
		u.BestScore = v
		if p > 0 {
			f := float64(p)
			u.BestScoreTopPercent = &f
		}
	} else {
		return nil, err
	}

	if v, p, err := stats.GetVictoriesAndTop(userID); err == nil {
		u.Victories = v
		if p > 0 {
			f := float64(p)
			u.VictoriesTopPercent = &f
		}
	} else {
		return nil, err
	}

	if v, p, err := stats.GetAvgScoreAndTop(userID); err == nil {
		u.AvgScore = float64(v)
		if p > 0 {
			f := float64(p)
			u.AvgScoreTopPercent = &f
		}
	} else {
		return nil, err
	}

	if v, p, err := stats.GetAvgPointsPerMoveAndTop(userID); err == nil {
		u.AvgPointsPerMove = float64(v)
		if p > 0 {
			f := float64(p)
			u.AvgPointsPerMoveTopPercent = &f
		}
	} else {
		return nil, err
	}

	if v, p, err := stats.GetBestMoveScoreAndTop(userID); err == nil {
		u.BestMoveScore = v
		if p > 0 {
			f := float64(p)
			u.BestMoveScoreTopPercent = &f
		}
	} else {
		return nil, err
	}
	return &u, nil
}
