package stats

import (
	"database/sql"
	"math"

	"github.com/ZiplEix/scrabble/api/database"
)

// getGamesCountAndTop returns (games_count, top_percent, error)
func GetGamesCountAndTop(userID int64) (int, int, error) {
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
func GetBestScoreAndTop(userID int64) (int, int, error) {
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
func GetVictoriesAndTop(userID int64) (int, int, error) {
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
func GetAvgScoreAndTop(userID int64) (int, int, error) {
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
func GetAvgPointsPerMoveAndTop(userID int64) (int, int, error) {
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
func GetBestMoveScoreAndTop(userID int64) (int, int, error) {
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
