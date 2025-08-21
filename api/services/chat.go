package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/utils"
	"go.uber.org/zap"
)

// IsUserInGame returns true if the user is part of the game
func IsUserInGame(userID int64, gameID string) (bool, error) {
	var dummy int
	err := database.QueryRow(`SELECT 1 FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		zap.L().Error("failed to validate user in game", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return false, err
	}
	return true, nil
}

// CreateMessage persists a message and returns a serializable representation
func CreateMessage(userID int64, gameID, content string, meta map[string]any) (map[string]any, error) {
	// validate membership
	inGame, err := IsUserInGame(userID, gameID)
	if err != nil {
		return nil, err
	}
	if !inGame {
		return nil, fmt.Errorf("user not in game")
	}

	metaJSON := []byte(nil)
	if meta != nil {
		if b, err := json.Marshal(meta); err == nil {
			metaJSON = b
		} else {
			// ignore meta on marshal failure
			zap.L().Warn("failed to marshal message meta", zap.Error(err))
		}
	}

	var id int64
	var createdAt time.Time
	err = database.QueryRow(`INSERT INTO messages (game_id, user_id, content, meta, created_at) VALUES ($1, $2, $3, $4, now()) RETURNING id, created_at`, gameID, userID, content, metaJSON).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	go func(gameID string) {
		// get sender username
		var senderUsername string
		err := database.QueryRow(`SELECT username FROM users WHERE id = $1`, userID).Scan(&senderUsername)
		if err != nil {
			zap.L().Warn("failed to fetch sender username", zap.Error(err), zap.Int64("user_id", userID))
			return
		}

		// get the game name
		var gameName string
		err = database.QueryRow(`SELECT name FROM games WHERE id = $1`, gameID).Scan(&gameName)
		if err != nil {
			zap.L().Warn("failed to fetch game name", zap.Error(err), zap.String("game_id", gameID))
			return
		}

		var users []int
		rows, err := database.Query(`SELECT user_id FROM game_players WHERE game_id = $1`, gameID)
		if err != nil {
			zap.L().Warn("failed to fetch game players", zap.Error(err), zap.String("game_id", gameID))
			return
		}
		for rows.Next() {
			var uid int
			if err := rows.Scan(&uid); err != nil {
				zap.L().Warn("failed to scan game player", zap.Error(err), zap.String("game_id", gameID))
				continue // skip this user if there's an error
			}
			users = append(users, uid)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			zap.L().Warn("error reading game players", zap.Error(err), zap.String("game_id", gameID))
		}

		// helper: check if a user allows message notifications
		userAllowsMessages := func(uid int64) bool {
			// try a dedicated boolean column first (backwards compatible)
			var allowed sql.NullBool
			if err := database.QueryRow(`SELECT messages_notifications_enabled FROM users WHERE id = $1`, uid).Scan(&allowed); err == nil {
				if allowed.Valid {
					return allowed.Bool
				}
			}
			// fallback: try JSONB column "notification_prefs"
			var prefsRaw []byte
			if err := database.QueryRow(`SELECT notification_prefs FROM users WHERE id = $1`, uid).Scan(&prefsRaw); err == nil && len(prefsRaw) > 0 {
				var prefs map[string]any
				if err := json.Unmarshal(prefsRaw, &prefs); err == nil {
					if v, ok := prefs["messages"]; ok {
						if b, ok := v.(bool); ok {
							return b
						}
					}
				}
			}
			// default: allow notifications
			return true
		}

		for _, uid := range users {
			// skip sender
			if int64(uid) == userID {
				continue
			}
			if !userAllowsMessages(int64(uid)) {
				continue
			}
			payload := utils.NotificationPayload{
				Title: "Nouveau message de " + senderUsername + " dans la partie " + gameName,
				Body:  content,
				Url:   fmt.Sprintf("/games/%s/chat", gameID),
			}
			_ = utils.SendNotificationToUserByID(int64(uid), payload)
		}
	}(gameID)

	// build response
	msg := map[string]any{
		"id":         id,
		"game_id":    gameID,
		"user_id":    userID,
		"content":    content,
		"meta":       meta,
		"created_at": createdAt,
	}
	return msg, nil
}

// GetMessages returns messages for a game. It validates that the user is part of the game.
func GetMessages(userID int64, gameID string) ([]map[string]any, error) {
	// validate membership
	inGame, err := IsUserInGame(userID, gameID)
	if err != nil {
		return nil, err
	}
	if !inGame {
		return nil, fmt.Errorf("user not in game")
	}

	rows, err := database.Query(`SELECT id, user_id, content, meta, created_at FROM messages WHERE game_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []map[string]any
	for rows.Next() {
		var id int64
		var uid int64
		var content string
		var metaRaw []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &uid, &content, &metaRaw, &createdAt); err != nil {
			return nil, err
		}
		var meta any = nil
		if len(metaRaw) > 0 {
			var m any
			if err := json.Unmarshal(metaRaw, &m); err == nil {
				meta = m
			}
		}
		res = append(res, map[string]any{
			"id":         id,
			"user_id":    uid,
			"content":    content,
			"meta":       meta,
			"created_at": createdAt,
		})
	}
	return res, nil
}

// DeleteMessage soft-deletes a message if the user is its owner and belongs to the game
func DeleteMessage(userID int64, gameID, msgID string) error {
	// membership check
	inGame, err := IsUserInGame(userID, gameID)
	if err != nil {
		return err
	}
	if !inGame {
		return fmt.Errorf("forbidden")
	}

	// check ownership and existence
	var ownerID int64
	var dbGameID string
	err = database.QueryRow(`SELECT user_id, game_id FROM messages WHERE id = $1 AND deleted_at IS NULL`, msgID).Scan(&ownerID, &dbGameID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}

	if dbGameID != gameID {
		return fmt.Errorf("forbidden")
	}
	if ownerID != userID {
		return fmt.Errorf("forbidden")
	}

	_, err = database.Exec(`UPDATE messages SET deleted_at = now() WHERE id = $1`, msgID)
	if err != nil {
		return err
	}
	return nil
}

// MarkMessagesRead sets the last_read_message_id for a user+game. If lastMessageID == 0
// the function will try to find the latest message id for the game and use it.
func MarkMessagesRead(userID int64, gameID string, lastMessageID int64) error {
	// validate membership
	inGame, err := IsUserInGame(userID, gameID)
	if err != nil {
		return err
	}
	if !inGame {
		return fmt.Errorf("user not in game")
	}

	if lastMessageID == 0 {
		// fetch latest message id
		var latestID int64
		err := database.QueryRow(`SELECT id FROM messages WHERE game_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`, gameID).Scan(&latestID)
		if err != nil {
			if err == sql.ErrNoRows {
				// nothing to mark
				return nil
			}
			return err
		}
		lastMessageID = latestID
	}

	// upsert into game_message_reads table
	_, err = database.Exec(`
		INSERT INTO game_message_reads (user_id, game_id, last_read_message_id, last_read_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, game_id) DO UPDATE
		SET last_read_message_id = GREATEST(coalesce(game_message_reads.last_read_message_id, 0), EXCLUDED.last_read_message_id),
			last_read_at = now()
	`, userID, gameID, lastMessageID)
	if err != nil {
		return err
	}
	return nil
}

// GetTotalUnreadMessagesForUser returns total count of messages newer than user's last_read per game
func GetTotalUnreadMessagesForUser(userID int64) (int64, error) {
	var total int64
	err := database.QueryRow(`
		SELECT COALESCE(SUM(cnt),0) FROM (
			SELECT g.id, COUNT(m.id) AS cnt
			FROM games g
			JOIN game_players gp ON gp.game_id = g.id
			JOIN messages m ON m.game_id = g.id AND m.deleted_at IS NULL
			LEFT JOIN game_message_reads r ON r.user_id = gp.player_id AND r.game_id = g.id::text
			WHERE gp.player_id = $1 AND (r.last_read_at IS NULL OR m.created_at > r.last_read_at)
			GROUP BY g.id
		) s
	`, userID).Scan(&total)
	return total, err
}

// GetUnreadMessagesForUser returns up to a limited list of unread messages for a user
// useful for debugging: returns id, game_id, user_id, content (trimmed), created_at
func GetUnreadMessagesForUser(userID int64, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := database.Query(`
		SELECT m.id, m.game_id, m.user_id, m.content, m.created_at
		FROM messages m
		JOIN game_players gp ON gp.game_id = m.game_id
		LEFT JOIN game_message_reads r ON r.user_id = gp.player_id AND r.game_id = m.game_id::text
		WHERE gp.player_id = $1 AND m.deleted_at IS NULL AND (r.last_read_at IS NULL OR m.created_at > r.last_read_at)
		ORDER BY m.created_at ASC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var id int64
		var gameID string
		var uid int64
		var content string
		var createdAt time.Time
		if err := rows.Scan(&id, &gameID, &uid, &content, &createdAt); err != nil {
			return nil, err
		}
		snippet := content
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		out = append(out, map[string]any{
			"id":         id,
			"game_id":    gameID,
			"user_id":    uid,
			"content":    snippet,
			"created_at": createdAt,
		})
	}
	return out, nil
}

// GetUnreadCountForUserInGame returns the number of unread messages for a user in a specific game
func GetUnreadCountForUserInGame(userID int64, gameID string) (int64, error) {
	// validate membership
	inGame, err := IsUserInGame(userID, gameID)
	if err != nil {
		return 0, err
	}
	if !inGame {
		return 0, fmt.Errorf("user not in game")
	}

	var total int64
	err = database.QueryRow(`
		SELECT COUNT(m.id)
		FROM messages m
		LEFT JOIN game_message_reads r ON r.user_id = $1 AND r.game_id = m.game_id::text
		WHERE m.game_id = $2 AND m.deleted_at IS NULL AND (r.last_read_at IS NULL OR m.created_at > r.last_read_at)
	`, userID, gameID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
