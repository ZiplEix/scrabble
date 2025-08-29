package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestUsers_FilterAndExclude(t *testing.T) {
	resetUsers(t)

	// current user (should be excluded)
	me, err := CreateUser("meuser", "x")
	require.NoError(t, err)

	// candidates
	_, _ = CreateUser("alpha", "x")
	_, _ = CreateUser("Alain", "x")
	_, _ = CreateUser("ALBERT", "x")
	_, _ = CreateUser("bob", "x")

	// also create a username starting with query but it's self (to check exclusion by id)
	// already created meuser which starts with "me"

	got, err := SuggestUsers(me.ID, "al")
	require.NoError(t, err)
	require.NotNil(t, got)

	// Expect only alpha, Alain, ALBERT (case-insensitive prefix match), not bob and not meuser
	for _, u := range got {
		// none of them should be me
		assert.NotEqual(t, me.ID, u.ID)
		assert.True(t, strings.HasPrefix(strings.ToLower(u.Username), "al"))
	}
	// Minimally, we should have 3 suggestions
	assert.GreaterOrEqual(t, len(got), 3)
}

func TestSuggestUsers_Limit10(t *testing.T) {
	resetUsers(t)
	me, err := CreateUser("current", "x")
	require.NoError(t, err)

	// create >10 matching users
	for i := 0; i < 15; i++ {
		_, err := CreateUser("testuser"+string(rune('a'+i)), "x")
		require.NoError(t, err)
	}
	// create some non-matching
	_, _ = CreateUser("other1", "x")
	_, _ = CreateUser("xyz", "x")

	got, err := SuggestUsers(me.ID, "testuser")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.LessOrEqual(t, len(got), 10)
}

func TestSuggestUsers_NoMatches(t *testing.T) {
	resetUsers(t)
	me, err := CreateUser("john", "x")
	require.NoError(t, err)

	_, _ = CreateUser("alice", "x")
	_, _ = CreateUser("bob", "x")

	got, err := SuggestUsers(me.ID, "zzz")
	require.NoError(t, err)
	require.Len(t, got, 0)
}
