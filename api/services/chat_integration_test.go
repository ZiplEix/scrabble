package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"strconv"

	"github.com/ZiplEix/scrabble/api/database"
)

func resetChatDeps(t *testing.T) {
	t.Helper()
	// Respect FK constraints
	_, err := database.Exec("DELETE FROM game_message_reads")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM messages")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM game_players")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM games")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM users")
	require.NoError(t, err)
}

func createGameWithPlayers(t *testing.T, uids ...int64) string {
	t.Helper()
	gid := uuid.New().String()
	// minimal valid game
	_, err := database.Exec(`
        INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, created_at)
        VALUES ($1, $2, $3, 'ongoing', NULL, $4::jsonb, $5, now())
    `, gid, "chat-game", uids[0], "[]", "")
	require.NoError(t, err)
	pos := 1
	for _, uid := range uids {
		_, err := database.Exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ($1, $2, '', $3, 0)`, gid, uid, pos)
		require.NoError(t, err)
		pos++
	}
	return gid
}

func TestIsUserInGame(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1", "x")
	u2, _ := CreateUser("p2", "x")
	gid := createGameWithPlayers(t, u1.ID)

	ok, err := IsUserInGame(u1.ID, gid)
	require.NoError(t, err)
	assert.True(t, ok)

	ok2, err := IsUserInGame(u2.ID, gid)
	require.NoError(t, err)
	assert.False(t, ok2)
}

func TestCreateMessage_AndGetMessages(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("sender", "x")
	u2, _ := CreateUser("reader", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	meta := map[string]any{"foo": "bar"}
	msg, err := CreateMessage(u1.ID, gid, "hello world", meta)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "hello world", msg["content"])
	assert.Equal(t, meta, msg["meta"])
	assert.NotZero(t, msg["id"])

	// both players should be able to read
	list1, err := GetMessages(u1.ID, gid)
	require.NoError(t, err)
	require.Len(t, list1, 1)
	assert.Equal(t, "hello world", list1[0]["content"])

	list2, err := GetMessages(u2.ID, gid)
	require.NoError(t, err)
	require.Len(t, list2, 1)

	// stranger cannot read
	stranger, _ := CreateUser("stranger", "x")
	_, err = GetMessages(stranger.ID, gid)
	require.Error(t, err)
}

func TestDeleteMessage_OwnerOnly(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("owner", "x")
	u2, _ := CreateUser("other", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	msg, err := CreateMessage(u1.ID, gid, "to delete", map[string]any{"k": "v"})
	require.NoError(t, err)
	id := msg["id"].(int64)

	// other user cannot delete
	err = DeleteMessage(u2.ID, gid, int64ToString(id))
	require.Error(t, err)

	// owner can delete
	err = DeleteMessage(u1.ID, gid, int64ToString(id))
	require.NoError(t, err)

	// Verify it no longer appears
	list, err := GetMessages(u1.ID, gid)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestMarkReadAndUnreadCounts(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("author", "x")
	u2, _ := CreateUser("reader2", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	// two messages from u1
	_, err := CreateMessage(u1.ID, gid, "m1", map[string]any{})
	require.NoError(t, err)
	msg2, err := CreateMessage(u1.ID, gid, "m2", map[string]any{})
	require.NoError(t, err)

	// u2 marks as read up to m2 explicitly
	id2 := msg2["id"].(int64)
	require.NoError(t, MarkMessagesRead(u2.ID, gid, id2))

	// Small delay to ensure next message is strictly after last_read_at (db timestamps)
	time.Sleep(200 * time.Millisecond)

	// new unread message
	_, err = CreateMessage(u1.ID, gid, "m3", map[string]any{})
	require.NoError(t, err)

	// per game
	cnt, err := GetUnreadCountForUserInGame(u2.ID, gid)
	require.NoError(t, err)
	assert.Equal(t, int64(1), cnt)

	// total across games
	total, err := GetTotalUnreadMessagesForUser(u2.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))

	// unread messages list with limit
	list, err := GetUnreadMessagesForUser(u2.ID, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)
}

func int64ToString(v int64) string {
	return fmtInt(v)
}

// split out to avoid importing strconv everywhere
func fmtInt(v int64) string {
	// simple strconv.FormatInt replacement without new import for conciseness
	// but we can import strconv directly; prefer clarity:
	return strconv.FormatInt(v, 10)
}

// -------- Additional edge cases --------

func TestCreateMessage_NonParticipant_Error(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_chat_np", "x")
	u2, _ := CreateUser("p2_chat_np", "x")
	gid := createGameWithPlayers(t, u1.ID) // only u1 in game

	_, err := CreateMessage(u2.ID, gid, "hi", map[string]any{"a": 1})
	require.Error(t, err)
}

func TestCreateMessage_InvalidMeta_Error(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_meta", "x")
	u2, _ := CreateUser("p2_meta", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	bad := map[string]any{"bad": func() {}}
	_, err := CreateMessage(u1.ID, gid, "hello", bad)
	require.Error(t, err)
}

func TestDeleteMessage_NotFound(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("owner_nf", "x")
	gid := createGameWithPlayers(t, u1.ID)

	// deleting unknown id should yield not found
	err := DeleteMessage(u1.ID, gid, "123456789")
	require.Error(t, err)
}

func TestDeleteMessage_WrongGame(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("owner_wg", "x")
	u2, _ := CreateUser("mate_wg", "x")
	g1 := createGameWithPlayers(t, u1.ID, u2.ID)
	g2 := createGameWithPlayers(t, u1.ID, u2.ID)

	m, err := CreateMessage(u1.ID, g1, "in g1", map[string]any{"k": "v"})
	require.NoError(t, err)
	id := m["id"].(int64)

	// try to delete using the other game id → forbidden
	err = DeleteMessage(u1.ID, g2, int64ToString(id))
	require.Error(t, err)
}

func TestMarkMessagesRead_NonParticipant_Error(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_mr_np", "x")
	u2, _ := CreateUser("p2_mr_np", "x")
	gid := createGameWithPlayers(t, u1.ID)

	err := MarkMessagesRead(u2.ID, gid, 0)
	require.Error(t, err)
}

func TestMarkMessagesRead_ZeroID_NoMessages_NoOp(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_mr0", "x")
	u2, _ := CreateUser("p2_mr0", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	// no messages yet → should be no-op without insert
	err := MarkMessagesRead(u2.ID, gid, 0)
	require.NoError(t, err)

	var cnt int
	err = database.QueryRow(`SELECT COUNT(*) FROM game_message_reads WHERE user_id = $1 AND game_id = $2`, u2.ID, gid).Scan(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 0, cnt)
}

func TestMarkMessagesRead_GreatestDoesNotRegress(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_mr_g", "x")
	u2, _ := CreateUser("p2_mr_g", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	m1, _ := CreateMessage(u1.ID, gid, "m1", map[string]any{})
	m2, _ := CreateMessage(u1.ID, gid, "m2", map[string]any{})
	m3, _ := CreateMessage(u1.ID, gid, "m3", map[string]any{})

	id1 := m1["id"].(int64)
	id2 := m2["id"].(int64)
	id3 := m3["id"].(int64)

	require.NoError(t, MarkMessagesRead(u2.ID, gid, id2))
	// regression attempt
	require.NoError(t, MarkMessagesRead(u2.ID, gid, id1))

	var lastID int64
	err := database.QueryRow(`SELECT last_read_message_id FROM game_message_reads WHERE user_id = $1 AND game_id = $2`, u2.ID, gid).Scan(&lastID)
	require.NoError(t, err)
	assert.Equal(t, id2, lastID)

	// advance
	require.NoError(t, MarkMessagesRead(u2.ID, gid, id3))
	err = database.QueryRow(`SELECT last_read_message_id FROM game_message_reads WHERE user_id = $1 AND game_id = $2`, u2.ID, gid).Scan(&lastID)
	require.NoError(t, err)
	assert.Equal(t, id3, lastID)
}

func TestGetUnreadMessagesForUser_LimitBounds(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_un", "x")
	u2, _ := CreateUser("p2_un", "x")
	gid := createGameWithPlayers(t, u1.ID, u2.ID)

	// Create 5 messages from u1
	for i := 0; i < 5; i++ {
		_, err := CreateMessage(u1.ID, gid, "msg", map[string]any{})
		require.NoError(t, err)
	}

	list1, err := GetUnreadMessagesForUser(u2.ID, 1)
	require.NoError(t, err)
	assert.Len(t, list1, 1)

	list0, err := GetUnreadMessagesForUser(u2.ID, 0) // defaults to 200
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list0), 5)
}

func TestGetUnreadCountForUserInGame_NotInGame_Error(t *testing.T) {
	resetChatDeps(t)
	u1, _ := CreateUser("p1_uc", "x")
	u2, _ := CreateUser("p2_uc", "x")
	gid := createGameWithPlayers(t, u1.ID) // only u1

	_, err := GetUnreadCountForUserInGame(u2.ID, gid)
	require.Error(t, err)
}
