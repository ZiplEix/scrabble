package stats

import (
	"database/sql"
	"math"

	"github.com/ZiplEix/scrabble/api/database"
)

func pctChange(curr, prev int) float64 {
	if prev == 0 {
		if curr == 0 {
			return 0
		}
		return 100
	}
	v := (float64(curr) - float64(prev)) / float64(prev) * 100
	return math.Round(v*100) / 100
}

// GetActiveUsersCountAndVariance returns (count_last7days, percent_variation_vs_prev7days, error)
// Active users are defined as distinct users who either sent messages or made game moves in the period.
func GetActiveUsersCountAndVariance() (int, float64, error) {
	var curr int
	var prev int

	// current 7 days
	err := database.QueryRow(`
		SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM messages WHERE created_at >= now() - interval '7 days' AND deleted_at IS NULL
			UNION
			SELECT player_id AS user_id FROM game_moves WHERE created_at >= now() - interval '7 days'
		) t
	`).Scan(&curr)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	// previous 7 days
	err = database.QueryRow(`
		SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM messages WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days' AND deleted_at IS NULL
			UNION
			SELECT player_id AS user_id FROM game_moves WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days'
		) t
	`).Scan(&prev)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	return curr, pctChange(curr, prev), nil
}

// GetCreatedGamesCountAndVariance returns created games count and variance
func GetCreatedGamesCountAndVariance() (int, float64, error) {
	var curr int
	var prev int
	if err := database.QueryRow(`SELECT COUNT(*) FROM games WHERE created_at >= now() - interval '7 days'`).Scan(&curr); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM games WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days'`).Scan(&prev); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	return curr, pctChange(curr, prev), nil
}

// GetActiveGamesCountAndVariance counts distinct games that had activity (messages or moves) in the period
func GetActiveGamesCountAndVariance() (int, float64, error) {
	var curr int
	var prev int
	if err := database.QueryRow(`
		SELECT COUNT(DISTINCT game_id) FROM (
			SELECT game_id FROM messages WHERE created_at >= now() - interval '7 days' AND deleted_at IS NULL
			UNION
			SELECT game_id FROM game_moves WHERE created_at >= now() - interval '7 days'
		) t
	`).Scan(&curr); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err := database.QueryRow(`
		SELECT COUNT(DISTINCT game_id) FROM (
			SELECT game_id FROM messages WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days' AND deleted_at IS NULL
			UNION
			SELECT game_id FROM game_moves WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days'
		) t
	`).Scan(&prev); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	return curr, pctChange(curr, prev), nil
}

// GetMessageSendCountAndVariance returns number of messages sent in last 7 days and variance
func GetMessageSendCountAndVariance() (int, float64, error) {
	var curr int
	var prev int
	if err := database.QueryRow(`SELECT COUNT(*) FROM messages WHERE created_at >= now() - interval '7 days' AND deleted_at IS NULL`).Scan(&curr); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM messages WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days' AND deleted_at IS NULL`).Scan(&prev); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	return curr, pctChange(curr, prev), nil
}

// GetTicketsCountAndVariance returns number of reports (tickets) created in last 7 days and variance
func GetTicketsCountAndVariance() (int, float64, error) {
	var curr int
	var prev int
	if err := database.QueryRow(`SELECT COUNT(*) FROM reports WHERE created_at >= now() - interval '7 days'`).Scan(&curr); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM reports WHERE created_at >= now() - interval '14 days' AND created_at < now() - interval '7 days'`).Scan(&prev); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	return curr, pctChange(curr, prev), nil
}
