package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
)

func resetPushSubs(t *testing.T) {
	t.Helper()
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

type subDoc struct {
	Endpoint string            `json:"endpoint"`
	Keys     map[string]string `json:"keys"`
}

func readSubscriptionJSON(t *testing.T, userID int64) subDoc {
	t.Helper()
	var subJSON string
	err := database.QueryRow("SELECT subscription FROM push_subscriptions WHERE user_id = $1", userID).Scan(&subJSON)
	require.NoError(t, err)
	var doc subDoc
	require.NoError(t, json.Unmarshal([]byte(subJSON), &doc))
	return doc
}

func TestPushSubscribe_Insert(t *testing.T) {
	resetPushSubs(t)
	u, err := CreateUser("notifier", "pwd")
	require.NoError(t, err)

	sub := []byte(`{"endpoint":"https://ep","keys":{"p256dh":"a","auth":"b"}}`)
	err = PushSubscribe(u.ID, sub)
	require.NoError(t, err)

	doc := readSubscriptionJSON(t, u.ID)
	assert.Equal(t, "https://ep", doc.Endpoint)
	assert.Equal(t, "a", doc.Keys["p256dh"])
	assert.Equal(t, "b", doc.Keys["auth"])
}

func TestPushSubscribe_Upsert(t *testing.T) {
	resetPushSubs(t)
	u, err := CreateUser("upserter", "pwd")
	require.NoError(t, err)

	err = PushSubscribe(u.ID, []byte(`{"endpoint":"https://old","keys":{"p256dh":"1","auth":"2"}}`))
	require.NoError(t, err)
	doc1 := readSubscriptionJSON(t, u.ID)
	assert.Equal(t, "https://old", doc1.Endpoint)

	// upsert with new endpoint
	err = PushSubscribe(u.ID, []byte(`{"endpoint":"https://new","keys":{"p256dh":"x","auth":"y"}}`))
	require.NoError(t, err)
	doc2 := readSubscriptionJSON(t, u.ID)
	assert.Equal(t, "https://new", doc2.Endpoint)
	assert.Equal(t, "x", doc2.Keys["p256dh"])
	assert.Equal(t, "y", doc2.Keys["auth"])
}

func TestPushUnsubscribe(t *testing.T) {
	resetPushSubs(t)
	u, err := CreateUser("unsub", "pwd")
	require.NoError(t, err)

	err = PushSubscribe(u.ID, []byte(`{"endpoint":"x","keys":{}}`))
	require.NoError(t, err)

	err = PushUnsubscribe(u.ID)
	require.NoError(t, err)

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM push_subscriptions WHERE user_id = $1", u.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestPushUnsubscribe_NoRow_NoError(t *testing.T) {
	resetPushSubs(t)
	u, err := CreateUser("no_row", "pwd")
	require.NoError(t, err)

	// no subscribe performed
	err = PushUnsubscribe(u.ID)
	require.NoError(t, err)
}
