package services

import (
	"database/sql"
	"encoding/json"
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

// GetUserAchievements charge la liste des succès et leur statut pour un utilisateur
func GetUserAchievements(userID int64) ([]response.AchievementResponse, error) {
	rows, err := database.Query(`
		SELECT a.id, a.title, a.description, a.badge_icon, a.category,
			   (ua.user_id IS NOT NULL) as unlocked, ua.unlocked_at
		FROM achievements a
		LEFT JOIN user_achievements ua ON a.id = ua.achievement_id AND ua.user_id = $1
		ORDER BY a.title ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	achievements := make([]response.AchievementResponse, 0)
	for rows.Next() {
		var ach response.AchievementResponse
		var unlockedAt sql.NullTime
		err := rows.Scan(&ach.ID, &ach.Title, &ach.Description, &ach.BadgeIcon, &ach.Category, &ach.Unlocked, &unlockedAt)
		if err != nil {
			return nil, err
		}
		if unlockedAt.Valid {
			ach.UnlockedAt = &unlockedAt.Time
		}
		achievements = append(achievements, ach)
	}
	return achievements, rows.Err()
}

// GetUserPublicByID retourne les informations publiques d'un utilisateur
func GetUserPublicByID(userID int64, viewerID int64) (*response.UserPublicResponse, error) {
	var u response.UserPublicResponse
	var createdAt time.Time
	err := database.QueryRow("SELECT id, username, rating, role, created_at FROM users WHERE id = $1", userID).Scan(&u.ID, &u.Username, &u.Rating, &u.Role, &createdAt)
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

	// 1. Face-à-Face si viewerID != userID
	if viewerID != userID && viewerID > 0 {
		var hresponse response.HeadToHeadInfo
		hresponse.RecentGames = make([]response.CommonGameSummary, 0)
		rows, err := database.Query(`
			SELECT g.id, g.name, g.status, COALESCE(g.winner_username, '') as winner, g.created_at,
				   gp_viewer.score as viewer_score, gp_profile.score as profile_score,
				   u_viewer.username as viewer_name, u_profile.username as profile_name
			FROM games g
			JOIN game_players gp_viewer ON g.id = gp_viewer.game_id AND gp_viewer.player_id = $1
			JOIN game_players gp_profile ON g.id = gp_profile.game_id AND gp_profile.player_id = $2
			JOIN users u_viewer ON gp_viewer.player_id = u_viewer.id
			JOIN users u_profile ON gp_profile.player_id = u_profile.id
			ORDER BY g.created_at DESC
		`, viewerID, userID)
		if err == nil {
			defer rows.Close()
			var totalViewerScore, totalProfileScore int
			var finishedGamesCount int
			for rows.Next() {
				var (
					gid          string
					gname        string
					gstatus      string
					winner       string
					createdAt    time.Time
					viewerScore  int
					profileScore int
					viewerName   string
					profileName  string
				)
				err := rows.Scan(&gid, &gname, &gstatus, &winner, &createdAt, &viewerScore, &profileScore, &viewerName, &profileName)
				if err == nil {
					hresponse.GamesPlayed++
					if gstatus == "ended" {
						finishedGamesCount++
						totalViewerScore += viewerScore
						totalProfileScore += profileScore
						if winner == viewerName {
							hresponse.UserWins++
						} else if winner == profileName {
							hresponse.OpponentWins++
						}
					}
					// Garder les 5 plus récents
					if len(hresponse.RecentGames) < 5 {
						hresponse.RecentGames = append(hresponse.RecentGames, response.CommonGameSummary{
							ID:        gid,
							Name:      gname,
							Status:    gstatus,
							Winner:    winner,
							UserScore: viewerScore,
							OppScore:  profileScore,
							CreatedAt: createdAt,
						})
					}
				}
			}
			if finishedGamesCount > 0 {
				hresponse.UserAvgScore = float64(totalViewerScore) / float64(finishedGamesCount)
				hresponse.OppAvgScore = float64(totalProfileScore) / float64(finishedGamesCount)
			}
			u.HeadToHead = &hresponse
		}
	}

	// 2. Succès du joueur
	achievements, err := GetUserAchievements(userID)
	if err == nil {
		u.Achievements = achievements
	}

	return &u, nil
}

// GetAllUsers returns all users with aggregated metrics and games list
func GetAllUsers() ([]response.AdminUserInfo, error) {
	// 1) Fetch base users
	rows, err := database.Query(`
		SELECT id, username, COALESCE(role,'user') as role, created_at, COALESCE(notification_prefs, '{}'::jsonb)
		FROM users
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]response.AdminUserInfo, 0)
	idIndex := make(map[int64]int)
	for rows.Next() {
		var u response.AdminUserInfo
		var prefsJSON []byte
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &prefsJSON); err != nil {
			return nil, err
		}
		if len(prefsJSON) > 0 {
			var m map[string]any
			if err := json.Unmarshal(prefsJSON, &m); err == nil {
				u.NotificationPrefs = m
			}
		}
		users = append(users, u)
		idIndex[u.ID] = len(users) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return users, nil
	}

	// 2) Messages count per user
	msgRows, err := database.Query(`
		SELECT user_id, COUNT(*)
		FROM messages
		GROUP BY user_id
	`)
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var uid int64
			var cnt int
			if err := msgRows.Scan(&uid, &cnt); err == nil {
				if idx, ok := idIndex[uid]; ok {
					users[idx].MessagesCount = cnt
				}
			}
		}
	}

	// 3) Games count and statuses per user via game_players
	gpRows, err := database.Query(`
		SELECT gp.player_id, COUNT(*) as games,
			   SUM(CASE WHEN g.status = 'ongoing' THEN 1 ELSE 0 END) as ongoing,
			   SUM(CASE WHEN g.status != 'ongoing' THEN 1 ELSE 0 END) as finished
		FROM game_players gp
		JOIN games g ON g.id = gp.game_id
		GROUP BY gp.player_id
	`)
	if err == nil {
		defer gpRows.Close()
		for gpRows.Next() {
			var uid int64
			var games, ongoing, finished int
			if err := gpRows.Scan(&uid, &games, &ongoing, &finished); err == nil {
				if idx, ok := idIndex[uid]; ok {
					users[idx].GamesCount = games
					users[idx].OngoingGames = ongoing
					users[idx].FinishedGames = finished
				}
			}
		}
	}

	// 4) Last activity per user: prioritize game moves, then fallback to messages
	// 4.a) From game_moves
	lastMoveRows, err := database.Query(`
		SELECT player_id, MAX(created_at) FROM game_moves GROUP BY player_id
	`)
	if err == nil {
		defer lastMoveRows.Close()
		for lastMoveRows.Next() {
			var uid int64
			var ts sql.NullTime
			if err := lastMoveRows.Scan(&uid, &ts); err == nil {
				if ts.Valid {
					if idx, ok := idIndex[uid]; ok {
						t := ts.Time
						users[idx].LastActivityAt = &t
					}
				}
			}
		}
	}

	// 4.b) Fallback from messages only if LastActivityAt not already set
	lastMsgRows, err := database.Query(`
		SELECT user_id, MAX(created_at) FROM messages GROUP BY user_id
	`)
	if err == nil {
		defer lastMsgRows.Close()
		for lastMsgRows.Next() {
			var uid int64
			var ts sql.NullTime
			if err := lastMsgRows.Scan(&uid, &ts); err == nil {
				if ts.Valid {
					if idx, ok := idIndex[uid]; ok {
						if users[idx].LastActivityAt == nil || users[idx].LastActivityAt.Before(ts.Time) == false {
							// Only set if not set by moves; keep move time as priority
							// Note: if both exist we keep move time, so skip when already set
						} else {
							t := ts.Time
							users[idx].LastActivityAt = &t
						}
						if users[idx].LastActivityAt == nil { // if still nil, set from messages
							t := ts.Time
							users[idx].LastActivityAt = &t
						}
					}
				}
			}
		}
	}

	// 5) Games list per user
	gamesRows, err := database.Query(`
		SELECT gp.player_id, g.id::text, g.name, g.status, g.created_at
		FROM game_players gp
		JOIN games g ON g.id = gp.game_id
		ORDER BY g.created_at DESC
	`)
	if err == nil {
		defer gamesRows.Close()
		for gamesRows.Next() {
			var uid int64
			var game response.AdminUserGame
			if err := gamesRows.Scan(&uid, &game.ID, &game.Name, &game.Status, &game.CreatedAt); err == nil {
				if idx, ok := idIndex[uid]; ok {
					users[idx].Games = append(users[idx].Games, game)
				}
			}
		}
	}

	return users, nil
}

// GetAdminUserByID returns a single AdminUserInfo for a given user id
func GetAdminUserByID(userID int64) (*response.AdminUserInfo, error) {
	var u response.AdminUserInfo

	// 1) Base user
	var prefsJSON []byte
	err := database.QueryRow(`
		SELECT id, username, COALESCE(role,'user') as role, created_at, COALESCE(notification_prefs, '{}'::jsonb)
		FROM users
		WHERE id = $1
	`, userID).Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &prefsJSON)
	if err != nil {
		return nil, err
	}
	if len(prefsJSON) > 0 {
		var m map[string]any
		if json.Unmarshal(prefsJSON, &m) == nil {
			u.NotificationPrefs = m
		}
	}

	// 2) Messages count
	if err := database.QueryRow(`SELECT COUNT(*) FROM messages WHERE user_id = $1`, userID).Scan(&u.MessagesCount); err != nil {
		// ignore error and keep zero if table missing or other issues
		_ = err
	}

	// 3) Games count and statuses
	var games, ongoing, finished sql.NullInt64
	if err := database.QueryRow(`
		SELECT COUNT(*) as games,
			   SUM(CASE WHEN g.status = 'ongoing' THEN 1 ELSE 0 END) as ongoing,
			   SUM(CASE WHEN g.status != 'ongoing' THEN 1 ELSE 0 END) as finished
		FROM game_players gp
		JOIN games g ON g.id = gp.game_id
		WHERE gp.player_id = $1
	`, userID).Scan(&games, &ongoing, &finished); err == nil {
		if games.Valid {
			u.GamesCount = int(games.Int64)
		}
		if ongoing.Valid {
			u.OngoingGames = int(ongoing.Int64)
		}
		if finished.Valid {
			u.FinishedGames = int(finished.Int64)
		}
	}

	// 4) Last activity: prefer game_moves over messages
	var lastMove sql.NullTime
	if err := database.QueryRow(`SELECT MAX(created_at) FROM game_moves WHERE player_id = $1`, userID).Scan(&lastMove); err == nil && lastMove.Valid {
		t := lastMove.Time
		u.LastActivityAt = &t
	} else {
		var lastMsg sql.NullTime
		if err := database.QueryRow(`SELECT MAX(created_at) FROM messages WHERE user_id = $1`, userID).Scan(&lastMsg); err == nil && lastMsg.Valid {
			t := lastMsg.Time
			u.LastActivityAt = &t
		}
	}

	// 5) Games list
	rows, err := database.Query(`
		SELECT g.id::text, g.name, g.status, g.created_at
		FROM game_players gp
		JOIN games g ON g.id = gp.game_id
		WHERE gp.player_id = $1
		ORDER BY g.created_at DESC
	`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var g response.AdminUserGame
			if err := rows.Scan(&g.ID, &g.Name, &g.Status, &g.CreatedAt); err == nil {
				u.Games = append(u.Games, g)
			}
		}
		_ = rows.Err()
	}

	return &u, nil
}
