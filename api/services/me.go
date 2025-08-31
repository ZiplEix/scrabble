package services

import (
	"database/sql"
	"encoding/json"
	"math"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
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
	if v, p, err := getGamesCountAndTop(userID); err == nil {
		res.GamesCount = v
		if p > 0 {
			f := float64(p)
			res.GamesCountTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := getBestScoreAndTop(userID); err == nil {
		res.BestScore = v
		if p > 0 {
			f := float64(p)
			res.BestScoreTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := getVictoriesAndTop(userID); err == nil {
		res.Victories = v
		if p > 0 {
			f := float64(p)
			res.VictoriesTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := getAvgScoreAndTop(userID); err == nil {
		res.AvgScore = float64(v)
		if p > 0 {
			f := float64(p)
			res.AvgScoreTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := getAvgPointsPerMoveAndTop(userID); err == nil {
		res.AvgPointsPerMove = float64(v)
		if p > 0 {
			f := float64(p)
			res.AvgPointsPerMoveTopPercent = &f
		}
	} else {
		return res, err
	}

	if v, p, err := getBestMoveScoreAndTop(userID); err == nil {
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

// getGamesCountAndTop returns (games_count, top_percent, error)
func getGamesCountAndTop(userID int64) (int, int, error) {
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM game_players WHERE player_id = $1", userID).Scan(&count); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	var nf sql.NullFloat64
	err := database.QueryRow(`
		WITH per_user AS (
			SELECT player_id AS user_id, COUNT(*) AS games
			FROM game_players
			GROUP BY player_id
		), ranked AS (
			SELECT user_id, games,
				RANK() OVER (ORDER BY games DESC) AS rnk,
				COUNT(*) OVER () AS total
			FROM per_user
		)
		SELECT ROUND((100.0 * rnk / total)::numeric, 2)
		FROM ranked WHERE user_id = $1
	`, userID).Scan(&nf)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	top := 0
	if err == nil && nf.Valid {
		top = int(math.Round(nf.Float64))
	}
	return count, top, nil
}

// getBestScoreAndTop returns (best_score, top_percent, error)
func getBestScoreAndTop(userID int64) (int, int, error) {
	var best int
	if err := database.QueryRow("SELECT COALESCE(MAX(score), 0) FROM game_players WHERE player_id = $1", userID).Scan(&best); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	var nf sql.NullFloat64
	err := database.QueryRow(`
		WITH per_user AS (
			SELECT player_id AS user_id, MAX(score) AS best
			FROM game_players
			GROUP BY player_id
		), ranked AS (
			SELECT user_id, best,
				RANK() OVER (ORDER BY best DESC) AS rnk,
				COUNT(*) OVER () AS total
			FROM per_user
		)
		SELECT ROUND((100.0 * rnk / total)::numeric, 2)
		FROM ranked WHERE user_id = $1
	`, userID).Scan(&nf)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	top := 0
	if err == nil && nf.Valid {
		top = int(math.Round(nf.Float64))
	}
	return best, top, nil
}

// getVictoriesAndTop returns (victories, top_percent, error)
func getVictoriesAndTop(userID int64) (int, int, error) {
	var wins int
	if err := database.QueryRow("SELECT COUNT(*) FROM games WHERE winner_username = (SELECT username FROM users WHERE id = $1)", userID).Scan(&wins); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	var nf sql.NullFloat64
	err := database.QueryRow(`
		WITH victories AS (
			SELECT u.id AS user_id, COUNT(*) AS wins
			FROM users u
			JOIN games g ON g.winner_username = u.username
			GROUP BY u.id
		), ranked AS (
			SELECT user_id, wins,
				RANK() OVER (ORDER BY wins DESC) AS rnk,
				COUNT(*) OVER () AS total
			FROM victories
		)
		SELECT ROUND((100.0 * rnk / total)::numeric, 2)
		FROM ranked WHERE user_id = $1
	`, userID).Scan(&nf)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	top := 0
	if err == nil && nf.Valid {
		top = int(math.Round(nf.Float64))
	}
	return wins, top, nil
}

// getAvgScoreAndTop returns (avg_score_rounded, top_percent, error)
func getAvgScoreAndTop(userID int64) (int, int, error) {
	var avg sql.NullFloat64
	if err := database.QueryRow(`
		SELECT COALESCE(AVG(gp.score), 0)
		FROM game_players gp
		JOIN games g ON gp.game_id = g.id
		WHERE gp.player_id = $1
		  AND g.ended_at IS NOT NULL
	`, userID).Scan(&avg); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	avgInt := 0
	if avg.Valid {
		avgInt = int(math.Round(avg.Float64))
	}
	var nf sql.NullFloat64
	err := database.QueryRow(`
		WITH per_user AS (
			SELECT gp.player_id AS user_id, AVG(gp.score) AS avg_score
			FROM game_players gp
			JOIN games g ON gp.game_id = g.id
			WHERE g.ended_at IS NOT NULL
			GROUP BY gp.player_id
		), ranked AS (
			SELECT user_id, avg_score,
				RANK() OVER (ORDER BY avg_score DESC) AS rnk,
				COUNT(*) OVER () AS total
			FROM per_user
		)
		SELECT ROUND((100.0 * rnk / total)::numeric, 2)
		FROM ranked WHERE user_id = $1
	`, userID).Scan(&nf)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	top := 0
	if err == nil && nf.Valid {
		top = int(math.Round(nf.Float64))
	}
	return avgInt, top, nil
}

// getAvgPointsPerMoveAndTop returns (avg_points_per_move_rounded, top_percent, error)
func getAvgPointsPerMoveAndTop(userID int64) (int, int, error) {
	var avg sql.NullFloat64
	if err := database.QueryRow(`
		SELECT AVG((move->>'score')::INT)
		FROM game_moves
		WHERE player_id = $1 AND (move ? 'score')
	`, userID).Scan(&avg); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	avgInt := 0
	if avg.Valid {
		avgInt = int(math.Round(avg.Float64))
	}
	var nf sql.NullFloat64
	err := database.QueryRow(`
		WITH per_user AS (
			SELECT player_id AS user_id, AVG((move->>'score')::INT) AS avg_pm
			FROM game_moves
			WHERE (move ? 'score')
			GROUP BY player_id
		), ranked AS (
			SELECT user_id, avg_pm,
				RANK() OVER (ORDER BY avg_pm DESC) AS rnk,
				COUNT(*) OVER () AS total
			FROM per_user
		)
		SELECT ROUND((100.0 * rnk / total)::numeric, 2)
		FROM ranked WHERE user_id = $1
	`, userID).Scan(&nf)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	top := 0
	if err == nil && nf.Valid {
		top = int(math.Round(nf.Float64))
	}
	return avgInt, top, nil
}

// getBestMoveScoreAndTop returns (best_move_score, top_percent, error)
func getBestMoveScoreAndTop(userID int64) (int, int, error) {
	var best int
	if err := database.QueryRow(`
		SELECT COALESCE(MAX((move->>'score')::INT), 0)
		FROM game_moves
		WHERE player_id = $1
	`, userID).Scan(&best); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	var nf sql.NullFloat64
	err := database.QueryRow(`
		WITH per_user AS (
			SELECT player_id AS user_id, MAX((move->>'score')::INT) AS best_move
			FROM game_moves
			WHERE (move ? 'score')
			GROUP BY player_id
		), ranked AS (
			SELECT user_id, best_move,
				RANK() OVER (ORDER BY best_move DESC) AS rnk,
				COUNT(*) OVER () AS total
			FROM per_user
		)
		SELECT ROUND((100.0 * rnk / total)::numeric, 2)
		FROM ranked WHERE user_id = $1
	`, userID).Scan(&nf)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	top := 0
	if err == nil && nf.Valid {
		top = int(math.Round(nf.Float64))
	}
	return best, top, nil
}

func UpdateUserNotificationPrefs(userID int64, prefs map[string]bool) error {
	b, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	_, err = database.Exec(`UPDATE users SET notification_prefs = $1 WHERE id = $2`, string(b), userID)
	return err
}
