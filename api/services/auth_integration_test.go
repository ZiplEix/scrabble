package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
)

// TestMain initializes the DB connection for auth service tests.
func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5433/scrabble_test?sslmode=disable"
	}
	// Try to init DB (test runner script should have started postgres and applied migrations)
	_ = database.Init(dsn)

	code := m.Run()
	os.Exit(code)
}

func resetUsers(t *testing.T) {
	t.Helper()
	// Delete in FK-safe order across services package
	// Reads -> messages/moves -> game_players -> games -> push_subscriptions -> reports -> users
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

func TestCreateUser_Success(t *testing.T) {
	resetUsers(t)

	u, err := CreateUser("alice", "s3cret")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "alice", u.Username)
	assert.NotZero(t, u.ID)

	// Verify row exists and password is hashed (not equal to plaintext)
	var storedHash string
	err = database.QueryRow("SELECT password FROM users WHERE username = $1", "alice").Scan(&storedHash)
	require.NoError(t, err)
	assert.NotEmpty(t, storedHash)
	assert.NotEqual(t, "s3cret", storedHash)
}

func TestCreateUser_Duplicate(t *testing.T) {
	resetUsers(t)

	_, err := CreateUser("bob", "pwd")
	require.NoError(t, err)

	_, err = CreateUser("bob", "pwd2")
	require.Error(t, err)
}

func TestVerifyUser_Success(t *testing.T) {
	resetUsers(t)
	_, err := CreateUser("charlie", "pass")
	require.NoError(t, err)

	u, err := VerifyUser("charlie", "pass")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "charlie", u.Username)
}

func TestVerifyUser_InvalidPassword(t *testing.T) {
	resetUsers(t)
	_, err := CreateUser("diana", "right")
	require.NoError(t, err)

	u, err := VerifyUser("diana", "wrong")
	require.Error(t, err)
	assert.Nil(t, u)
}

func TestVerifyUser_NotFound(t *testing.T) {
	resetUsers(t)
	u, err := VerifyUser("nobody", "whatever")
	require.Error(t, err)
	assert.Nil(t, u)
}

func TestUpdateUserPassword_Success(t *testing.T) {
	resetUsers(t)
	_, err := CreateUser("eve", "old")
	require.NoError(t, err)

	err = UpdateUserPassword("eve", "new")
	require.NoError(t, err)

	// Old password should fail
	uOld, errOld := VerifyUser("eve", "old")
	require.Error(t, errOld)
	assert.Nil(t, uOld)

	// New password should succeed
	uNew, errNew := VerifyUser("eve", "new")
	require.NoError(t, errNew)
	require.NotNil(t, uNew)
}

func TestGetUserByUsername(t *testing.T) {
	resetUsers(t)
	_, err := CreateUser("frank", "top")
	require.NoError(t, err)

	u, err := GetUserByUsername("frank")
	require.NotNil(t, u)
	assert.Equal(t, "frank", u.Username)
	assert.NoError(t, err)

	u2, err2 := GetUserByUsername("ghost")
	assert.Nil(t, u2)
	assert.Error(t, err2)
}
