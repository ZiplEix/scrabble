package services

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
)

func GetMeInfo(userID int64) (response.MeResponse, error) {
	var res response.MeResponse

	var createdAt time.Time
	// Start a transaction so all reads are consistent
	tx, err := database.DB.Begin()
	if err != nil {
		return res, err
	}
	// safe to call rollback in defer; Commit will make subsequent Rollback harmless
	defer func() {
		_ = tx.Rollback()
	}()

	// Get basic user info
	err = tx.QueryRow("SELECT id, username, role, created_at FROM users WHERE id = $1", userID).Scan(&res.ID, &res.Username, &res.Role, &createdAt)
	if err != nil {
		return res, err
	}
	res.CreatedAt = createdAt

	// Games count
	if err := tx.QueryRow("SELECT COUNT(*) FROM game_players WHERE player_id = $1", userID).Scan(&res.GamesCount); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Notifications enabled: check if a push subscription exists for the user
	if err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM push_subscriptions WHERE user_id = $1)", userID).Scan(&res.NotificationsEnabled); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Best score
	if err := tx.QueryRow("SELECT COALESCE(MAX(score), 0) FROM game_players WHERE player_id = $1", userID).Scan(&res.BestScore); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Victories: count in games where winner_username equals this user's username
	if err := tx.QueryRow("SELECT COUNT(*) FROM games WHERE winner_username = (SELECT username FROM users WHERE id = $1)", userID).Scan(&res.Victories); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Avg score on finished games
	if err := tx.QueryRow("SELECT COALESCE(AVG(gp.score), 0) FROM game_players gp JOIN games g ON gp.game_id = g.id WHERE gp.player_id = $1 AND g.ended_at IS NOT NULL", userID).Scan(&res.AvgScore); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Avg points per move (based on stored move.score)
	if err := tx.QueryRow(`
		SELECT COALESCE(AVG((move->>'score')::INT), 0)
		FROM game_moves
		WHERE player_id = $1
	`, userID).Scan(&res.AvgPointsPerMove); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// Best move score
	if err := tx.QueryRow(`
		SELECT COALESCE(MAX((move->>'score')::INT), 0)
		FROM game_moves
		WHERE player_id = $1
	`, userID).Scan(&res.BestMoveScore); err != nil && err != sql.ErrNoRows {
		return res, err
	}

	// notifications pref
	var prefsJSON sql.NullString
	if err := tx.QueryRow("SELECT notification_prefs FROM users WHERE id = $1", userID).Scan(&prefsJSON); err != nil && err != sql.ErrNoRows {
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

	if err := tx.Commit(); err != nil {
		return res, err
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
