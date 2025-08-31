package services

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/stats"
)

func GetMeInfo(userID int64) (response.MeResponse, error) {
	var res response.MeResponse

	// Basic user info
	var createdAt time.Time
	if err := database.QueryRow("SELECT id, username, role, created_at FROM users WHERE id = $1", userID).Scan(&res.ID, &res.Username, &res.Role, &createdAt); err != nil {
		return res, err
	}
	res.CreatedAt = createdAt

	// Notifications enabled
	if err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM push_subscriptions WHERE user_id = $1)", userID).Scan(&res.NotificationsEnabled); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Stats via helpers
	if v, p, err := stats.GetGamesCountAndTop(userID); err == nil {
		res.GamesCount = v
		if p > 0 {
			f := float64(p)
			res.GamesCountTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := stats.GetBestScoreAndTop(userID); err == nil {
		res.BestScore = v
		if p > 0 {
			f := float64(p)
			res.BestScoreTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := stats.GetVictoriesAndTop(userID); err == nil {
		res.Victories = v
		if p > 0 {
			f := float64(p)
			res.VictoriesTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := stats.GetAvgScoreAndTop(userID); err == nil {
		res.AvgScore = float64(v)
		if p > 0 {
			f := float64(p)
			res.AvgScoreTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := stats.GetAvgPointsPerMoveAndTop(userID); err == nil {
		res.AvgPointsPerMove = float64(v)
		if p > 0 {
			f := float64(p)
			res.AvgPointsPerMoveTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := stats.GetBestMoveScoreAndTop(userID); err == nil {
		res.BestMoveScore = v
		if p > 0 {
			f := float64(p)
			res.BestMoveScoreTopPercent = &f
		}
	} else {
		return res, err
	}

	// Notification prefs
	var prefsJSON sql.NullString
	if err := database.QueryRow("SELECT notification_prefs FROM users WHERE id = $1", userID).Scan(&prefsJSON); err != nil && err != sql.ErrNoRows {
		return res, err
	}
	if prefsJSON.Valid {
		var prefs struct {
			Turn     *bool `json:"turn"`
			Messages *bool `json:"messages"`
		}
		if err := json.Unmarshal([]byte(prefsJSON.String), &prefs); err == nil {
			if prefs.Turn != nil {
				res.TurnNotificationsEnabled = *prefs.Turn
			}
			if prefs.Messages != nil {
				res.MessagesNotificationsEnabled = *prefs.Messages
			}
		}
	}

	return res, nil
}

func UpdateUserNotificationPrefs(userID int64, prefs map[string]bool) error {
	b, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	_, err = database.Exec(`UPDATE users SET notification_prefs = $1 WHERE id = $2`, string(b), userID)
	return err
}
