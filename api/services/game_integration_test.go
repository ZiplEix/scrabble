package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
)

// resetAllGamesDeps supprime les données dans un ordre compatible avec les FKs
func resetAllGamesDeps(t *testing.T) {
	t.Helper()
	// Order: reads -> messages -> moves -> game_players -> games -> push_subscriptions -> reports -> users
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

func mustCreateUser(t *testing.T, username string) int64 {
	t.Helper()
	u, err := CreateUser(username, "pwd")
	require.NoError(t, err)
	return u.ID
}

func setPlayerRack(t *testing.T, gameID string, userID int64, rack string) {
	t.Helper()
	_, err := database.Exec(`UPDATE game_players SET rack = $1 WHERE game_id = $2 AND player_id = $3`, rack, gameID, userID)
	require.NoError(t, err)
}

func setGameTurnAndBag(t *testing.T, gameID string, turnUserID int64, bag string) {
	t.Helper()
	_, err := database.Exec(`UPDATE games SET current_turn = $1, available_letters = $2 WHERE id = $3`, turnUserID, bag, gameID)
	require.NoError(t, err)
}

func getGameFieldString(t *testing.T, gameID, field string) string {
	t.Helper()
	var s sql.NullString
	err := database.QueryRow(`SELECT `+field+` FROM games WHERE id = $1`, gameID).Scan(&s)
	require.NoError(t, err)
	if s.Valid {
		return s.String
	}
	return ""
}

func TestCreateGame_Success(t *testing.T) {
	resetAllGamesDeps(t)

	creator := mustCreateUser(t, "creator")
	_ = mustCreateUser(t, "p2")
	_ = mustCreateUser(t, "p3")

	gid, err := CreateGame(creator, "ma partie", []string{"p2", "p3"}, nil)
	require.NoError(t, err)
	require.NotNil(t, gid)

	// game exists and players rows created (3 players)
	var cnt int
	err = database.QueryRow("SELECT COUNT(*) FROM game_players WHERE game_id = $1", *gid).Scan(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 3, cnt)

	// current_turn should be the creator
	var ct int64
	err = database.QueryRow("SELECT current_turn FROM games WHERE id = $1", *gid).Scan(&ct)
	require.NoError(t, err)
	assert.Equal(t, creator, ct)
}

func TestCreateGame_RematchChecks(t *testing.T) {
	resetAllGamesDeps(t)

	u1 := mustCreateUser(t, "alice")
	_ = mustCreateUser(t, "bob")

	// original game
	gid, err := CreateGame(u1, "g1", []string{"bob"}, nil)
	require.NoError(t, err)
	require.NotNil(t, gid)

	orig := gid.String()

	// same creator can rematch
	gid2, err := CreateGame(u1, "rev", []string{"bob"}, &orig)
	require.NoError(t, err)
	require.NotNil(t, gid2)

	// other user cannot rematch from this original
	u2 := GetUserByUsername("bob").ID
	gid3, err := CreateGame(u2, "rev2", []string{"alice"}, &orig)
	require.Error(t, err)
	assert.Nil(t, gid3)

	// non-existent source
	fake := uuid.New().String()
	gid4, err := CreateGame(u1, "rev3", []string{"bob"}, &fake)
	require.Error(t, err)
	assert.Nil(t, gid4)
}

func TestDeleteGame_OnlyCreator(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "owner")
	_ = mustCreateUser(t, "guest")

	gid, err := CreateGame(u1, "to-delete", []string{"guest"}, nil)
	require.NoError(t, err)

	// non creator cannot delete
	err = DeleteGame(GetUserByUsername("guest").ID, gid.String())
	require.Error(t, err)

	// creator can delete
	err = DeleteGame(u1, gid.String())
	require.NoError(t, err)

	var count int
	_ = database.QueryRow("SELECT COUNT(*) FROM games WHERE id = $1", gid.String()).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestRenameGame_OnlyCreator(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "boss")
	_ = mustCreateUser(t, "other")
	gid, err := CreateGame(u1, "old", []string{"other"}, nil)
	require.NoError(t, err)

	// ok as creator
	err = RenameGame(u1, gid.String(), "new-name")
	require.NoError(t, err)
	name := getGameFieldString(t, gid.String(), "name")
	assert.Equal(t, "new-name", name)

	// forbidden as non-creator
	err = RenameGame(GetUserByUsername("other").ID, gid.String(), "x")
	require.Error(t, err)
}

func TestGetGameDetails_AfterMove(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "p1")
	_ = mustCreateUser(t, "p2")
	gid, err := CreateGame(u1, "details", []string{"p2"}, nil)
	require.NoError(t, err)
	g := gid.String()

	// Prepare deterministic rack and bag for u1, ensure it's u1's turn
	// Play word "CHAT" on center horizontally: (5,7) to (8,7)
	setPlayerRack(t, g, u1, "CHATXYZ")
	setGameTurnAndBag(t, g, u1, "") // empty bag to simplify

	move := request.PlayMoveRequest{
		Letters: []request.PlacedLetter{
			{X: 5, Y: 7, Char: "C"},
			{X: 6, Y: 7, Char: "H"},
			{X: 7, Y: 7, Char: "A"},
			{X: 8, Y: 7, Char: "T"},
		},
	}
	err = PlayMove(g, u1, move)
	require.NoError(t, err)

	// Details for p1
	info, err := GetGameDetails(u1, g)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "details", info.Name)
	assert.Len(t, info.Players, 2)
	assert.NotEmpty(t, info.YourRack)
	// Le sac a été vidé avant le coup pour simplifier le test
	assert.Equal(t, 0, info.RemainingLetters)
	// After the move turn should be player 2
	assert.NotEqual(t, u1, info.CurrentTurn)

	// Moves should contain our move
	assert.GreaterOrEqual(t, len(info.Moves), 1)
}

func TestGetNewRack_ChangesRackAndAdvancesTurn(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "rackman")
	_ = mustCreateUser(t, "mate")
	gid, err := CreateGame(u1, "getrack", []string{"mate"}, nil)
	require.NoError(t, err)
	g := gid.String()

	// Force small rack and bag, ensure it's u1's turn
	setPlayerRack(t, g, u1, "ABC")
	setGameTurnAndBag(t, g, u1, "DEFGHIJKLMNOP")

	newRack, err := GetNewRack(u1, g)
	require.NoError(t, err)
	assert.Len(t, newRack, 7)

	// current turn should have advanced to next player
	var ct int64
	err = database.QueryRow("SELECT current_turn FROM games WHERE id = $1", g).Scan(&ct)
	require.NoError(t, err)
	assert.NotEqual(t, u1, ct)
}

func TestGetGamesByUserID_Basic(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "gamer1")
	_ = mustCreateUser(t, "gamer2")
	gid, err := CreateGame(u1, "liste", []string{"gamer2"}, nil)
	require.NoError(t, err)
	g := gid.String()

	// make a quick move to set last_play_time via game_moves
	setPlayerRack(t, g, u1, "AAABCDE") // enough A's
	setGameTurnAndBag(t, g, u1, "")
	err = PlayMove(g, u1, request.PlayMoveRequest{
		Letters: []request.PlacedLetter{{X: 7, Y: 7, Char: "A"}, {X: 8, Y: 7, Char: "A"}},
	})
	require.NoError(t, err)

	list, err := GetGamesByUserID(u1)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	found := false
	for _, it := range list {
		if it.ID == g {
			found = true
			assert.Equal(t, "liste", it.Name)
			assert.True(t, !it.LastPlayTime.IsZero())
			assert.True(t, it.IsYourGame)
		}
	}
	assert.True(t, found)
}

func TestSimulateScore_Positive(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "simu1")
	_ = mustCreateUser(t, "simu2")
	gid, err := CreateGame(u1, "simu", []string{"simu2"}, nil)
	require.NoError(t, err)
	g := gid.String()

	score, err := SimulateScore(g, u1, []request.PlacedLetter{{X: 7, Y: 7, Char: "A"}, {X: 8, Y: 7, Char: "B"}})
	require.NoError(t, err)
	assert.Greater(t, score, 0)
}

func TestPassTurn_EndsGameAfterDoubleRound(t *testing.T) {
	resetAllGamesDeps(t)
	u1 := mustCreateUser(t, "pass1")
	u2 := mustCreateUser(t, "pass2")
	gid, err := CreateGame(u1, "passes", []string{"pass2"}, nil)
	require.NoError(t, err)
	g := gid.String()

	// ensure turn is u1 initially (by CreateGame) and racks/bag are default
	// 4 passes (2 * nb joueurs)
	require.NoError(t, PassTurn(u1, g)) // turn -> u2
	require.NoError(t, PassTurn(u2, g)) // -> u1
	// slight sleep to avoid same timestamp edge cases on some DBs
	time.Sleep(50 * time.Millisecond)
	require.NoError(t, PassTurn(u1, g)) // -> u2
	require.NoError(t, PassTurn(u2, g)) // should end game now

	var status string
	var endedAt sql.NullTime
	err = database.QueryRow(`SELECT status, ended_at FROM games WHERE id = $1`, g).Scan(&status, &endedAt)
	require.NoError(t, err)
	assert.Equal(t, "ended", status)
	assert.True(t, endedAt.Valid)
}
