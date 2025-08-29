package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
)

func resetAllMeDeps(t *testing.T) {
	t.Helper()
	// Order matters due to FKs
	_, err := database.Exec("DELETE FROM game_message_reads")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM messages")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM game_moves")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM game_players")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM games")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM push_subscriptions")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM reports")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM users")
	require.NoError(t, err)
}

func TestGetMeInfo_EmptyDefaults(t *testing.T) {
	resetAllMeDeps(t)

	u, err := CreateUser("martin", "x")
	require.NoError(t, err)

	me, err := GetMeInfo(u.ID)
	require.NoError(t, err)

	assert.Equal(t, u.ID, me.ID)
	assert.Equal(t, "martin", me.Username)
	assert.Equal(t, "user", me.Role)
	assert.False(t, me.CreatedAt.IsZero())
	assert.Equal(t, 0, me.GamesCount)
	assert.Equal(t, 0, me.BestScore)
	assert.Equal(t, 0, me.Victories)
	assert.InDelta(t, 0.0, me.AvgScore, 0.0001)
	// By default, prefs JSON default enables both
	assert.True(t, me.TurnNotificationsEnabled)
	assert.True(t, me.MessagesNotificationsEnabled)
	// No push subscription yet
	assert.False(t, me.NotificationsEnabled)
}

func TestGetMeInfo_WithSubsAndStats(t *testing.T) {
	resetAllMeDeps(t)

	u, err := CreateUser("winner", "x")
	require.NoError(t, err)

	// Add a push subscription to enable notifications
	_, err = database.Exec(`INSERT INTO push_subscriptions (user_id, subscription) VALUES ($1, $2::jsonb)`, u.ID, "{}")
	require.NoError(t, err)

	// Create one ongoing game and one ended game
	boardJSON := "[]"
	// Ongoing game
	g1 := uuid.New().String()
	_, err = database.Exec(`
        INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, created_at)
        VALUES ($1, $2, $3, $4, NULL, $5::jsonb, $6, now())
    `, g1, "game ongoing", u.ID, "ongoing", boardJSON, "")
	require.NoError(t, err)

	// Ended game with winner_username = user
	g2 := uuid.New().String()
	endedAt := time.Now()
	_, err = database.Exec(`
        INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, created_at, winner_username, ended_at)
        VALUES ($1, $2, $3, $4, NULL, $5::jsonb, $6, now(), $7, $8)
    `, g2, "game ended", u.ID, "ended", boardJSON, "", u.Username, endedAt)
	require.NoError(t, err)

	// Link user to both games with different scores
	_, err = database.Exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ($1, $2, $3, $4, $5)`, g1, u.ID, "", 1, 10)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ($1, $2, $3, $4, $5)`, g2, u.ID, "", 1, 20)
	require.NoError(t, err)

	me, err := GetMeInfo(u.ID)
	require.NoError(t, err)

	assert.Equal(t, 2, me.GamesCount)
	assert.Equal(t, 20, me.BestScore)
	assert.Equal(t, 1, me.Victories)
	assert.InDelta(t, 20.0, me.AvgScore, 0.0001)
	assert.True(t, me.NotificationsEnabled)
}

func TestUpdateUserNotificationPrefs_Toggle(t *testing.T) {
	resetAllMeDeps(t)

	u, err := CreateUser("prefs", "x")
	require.NoError(t, err)

	// Toggle to false/true
	err = UpdateUserNotificationPrefs(u.ID, map[string]bool{"turn": false, "messages": true})
	require.NoError(t, err)

	me, err := GetMeInfo(u.ID)
	require.NoError(t, err)
	assert.False(t, me.TurnNotificationsEnabled)
	assert.True(t, me.MessagesNotificationsEnabled)
}
