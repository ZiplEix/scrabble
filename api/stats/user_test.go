package stats

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
)

// TestMain: initialise la connexion à la DB de test (comme dans services/*_integration_test.go)
func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5433/scrabble_test?sslmode=disable"
	}
	_ = database.Init(dsn)
	code := m.Run()
	os.Exit(code)
}

func resetAll(t *testing.T) {
	t.Helper()
	// purge dans l'ordre des FK
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

type fixture struct {
	alice int64
	bob   int64
	carol int64
}

func seedStatsFixture(t *testing.T) fixture {
	t.Helper()
	resetAll(t)

	// users
	var aliceID, bobID, carolID int64
	require.NoError(t, database.QueryRow(`INSERT INTO users (username, password) VALUES ('alice','x') RETURNING id`).Scan(&aliceID))
	require.NoError(t, database.QueryRow(`INSERT INTO users (username, password) VALUES ('bob','x') RETURNING id`).Scan(&bobID))
	require.NoError(t, database.QueryRow(`INSERT INTO users (username, password) VALUES ('carol','x') RETURNING id`).Scan(&carolID))

	// minimal fields for games
	// board JSONB and available_letters are NOT NULL
	// g1,g2,g4,g5,g6: ended; g3: ongoing
	// winners: g1,g2,g5,g6 -> alice ; g4 -> bob
	exec := func(q string, args ...any) { _, err := database.Exec(q, args...); require.NoError(t, err) }

	exec(`INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, winner_username, ended_at)
        VALUES ('00000000-0000-0000-0000-000000000001','G1', $1::int, 'ended', $1::int, '{}'::jsonb, 'AAAA', 'alice', now())`, aliceID)
	exec(`INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, winner_username, ended_at)
        VALUES ('00000000-0000-0000-0000-000000000002','G2', $1::int, 'ended', $1::int, '{}'::jsonb, 'AAAA', 'alice', now())`, aliceID)
	exec(`INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters)
        VALUES ('00000000-0000-0000-0000-000000000003','G3', $1::int, 'ongoing', $1::int, '{}'::jsonb, 'AAAA')`, bobID)
	exec(`INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, winner_username, ended_at)
        VALUES ('00000000-0000-0000-0000-000000000004','G4', $1::int, 'ended', $1::int, '{}'::jsonb, 'AAAA', 'bob', now())`, bobID)
	exec(`INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, winner_username, ended_at)
        VALUES ('00000000-0000-0000-0000-000000000005','G5', $1::int, 'ended', $1::int, '{}'::jsonb, 'AAAA', 'alice', now())`, aliceID)
	exec(`INSERT INTO games (id, name, created_by, status, current_turn, board, available_letters, winner_username, ended_at)
        VALUES ('00000000-0000-0000-0000-000000000006','G6', $1::int, 'ended', $1::int, '{}'::jsonb, 'AAAA', 'alice', now())`, aliceID)

	// game_players scores
	// G1: alice 60, bob 30 (ended)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000001', $1::int, '', 1, 60)`, aliceID)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000001', $1::int, '', 2, 30)`, bobID)
	// G2: alice 40, carol 10 (ended)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000002', $1::int, '', 1, 40)`, aliceID)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000002', $1::int, '', 2, 10)`, carolID)
	// G3: bob 15, carol 20 (ongoing)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000003', $1::int, '', 1, 15)`, bobID)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000003', $1::int, '', 2, 20)`, carolID)
	// G4: bob 50, carol 5 (ended)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000004', $1::int, '', 1, 50)`, bobID)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000004', $1::int, '', 2, 5)`, carolID)
	// G5: alice 55 (ended)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000005', $1::int, '', 1, 55)`, aliceID)
	// G6: alice 10, bob 10 (ended)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000006', $1::int, '', 1, 10)`, aliceID)
	exec(`INSERT INTO game_players (game_id, player_id, rack, position, score) VALUES ('00000000-0000-0000-0000-000000000006', $1::int, '', 2, 10)`, bobID)

	// game_moves (JSONB with score)
	// alice: 12, 8  => avg 10, best 12
	exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ('00000000-0000-0000-0000-000000000001', $1::int, '{"score":12}')`, aliceID)
	exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ('00000000-0000-0000-0000-000000000002', $1::int, '{"score":8}')`, aliceID)
	// bob: 20, 25 => avg 22.5 (~23), best 25
	exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ('00000000-0000-0000-0000-000000000004', $1::int, '{"score":20}')`, bobID)
	exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ('00000000-0000-0000-0000-000000000006', $1::int, '{"score":25}')`, bobID)
	// carol: 7, 7 => avg 7, best 7
	exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ('00000000-0000-0000-0000-000000000002', $1::int, '{"score":7}')`, carolID)
	exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ('00000000-0000-0000-0000-000000000003', $1::int, '{"score":7}')`, carolID)

	return fixture{alice: aliceID, bob: bobID, carol: carolID}
}

func TestGetGamesCountAndTop(t *testing.T) {
	fx := seedStatsFixture(t)

	count, top, err := GetGamesCountAndTop(fx.alice)
	require.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Equal(t, 33, top)

	count, top, err = GetGamesCountAndTop(fx.bob)
	require.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Equal(t, 33, top) // tie rank=1 -> 33%

	count, top, err = GetGamesCountAndTop(fx.carol)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 100, top)
}

func TestGetBestScoreAndTop(t *testing.T) {
	fx := seedStatsFixture(t)

	best, top, err := GetBestScoreAndTop(fx.alice)
	require.NoError(t, err)
	assert.Equal(t, 60, best)
	assert.Equal(t, 33, top)

	best, top, err = GetBestScoreAndTop(fx.bob)
	require.NoError(t, err)
	assert.Equal(t, 50, best)
	assert.Equal(t, 67, top)

	best, top, err = GetBestScoreAndTop(fx.carol)
	require.NoError(t, err)
	assert.Equal(t, 20, best)
	assert.Equal(t, 100, top)
}

func TestGetVictoriesAndTop(t *testing.T) {
	fx := seedStatsFixture(t)

	wins, top, err := GetVictoriesAndTop(fx.alice)
	require.NoError(t, err)
	assert.Equal(t, 4, wins)
	assert.Equal(t, 33, top)

	wins, top, err = GetVictoriesAndTop(fx.bob)
	require.NoError(t, err)
	assert.Equal(t, 1, wins)
	assert.Equal(t, 67, top)

	// carol has 0 victories, rank row absent -> top=0
	wins, top, err = GetVictoriesAndTop(fx.carol)
	require.NoError(t, err)
	assert.Equal(t, 0, wins)
	assert.Equal(t, 100, top)
}

func TestGetAvgScoreAndTop(t *testing.T) {
	fx := seedStatsFixture(t)

	// Only ended games are considered
	avg, top, err := GetAvgScoreAndTop(fx.alice)
	require.NoError(t, err)
	assert.Equal(t, 41, avg) // round(41.25)
	assert.Equal(t, 33, top)

	avg, top, err = GetAvgScoreAndTop(fx.bob)
	require.NoError(t, err)
	assert.Equal(t, 30, avg)
	assert.Equal(t, 67, top)

	avg, top, err = GetAvgScoreAndTop(fx.carol)
	require.NoError(t, err)
	assert.Equal(t, 8, avg) // round(7.5)
	assert.Equal(t, 100, top)
}

func TestGetAvgPointsPerMoveAndTop(t *testing.T) {
	fx := seedStatsFixture(t)

	avg, top, err := GetAvgPointsPerMoveAndTop(fx.bob)
	require.NoError(t, err)
	assert.Equal(t, 23, avg) // round(22.5)
	assert.Equal(t, 33, top)

	avg, top, err = GetAvgPointsPerMoveAndTop(fx.alice)
	require.NoError(t, err)
	assert.Equal(t, 10, avg)
	assert.Equal(t, 67, top)

	avg, top, err = GetAvgPointsPerMoveAndTop(fx.carol)
	require.NoError(t, err)
	assert.Equal(t, 7, avg)
	assert.Equal(t, 100, top)
}

func TestGetBestMoveScoreAndTop(t *testing.T) {
	fx := seedStatsFixture(t)

	best, top, err := GetBestMoveScoreAndTop(fx.bob)
	require.NoError(t, err)
	assert.Equal(t, 25, best)
	assert.Equal(t, 33, top)

	best, top, err = GetBestMoveScoreAndTop(fx.alice)
	require.NoError(t, err)
	assert.Equal(t, 12, best)
	assert.Equal(t, 67, top)

	best, top, err = GetBestMoveScoreAndTop(fx.carol)
	require.NoError(t, err)
	assert.Equal(t, 7, best)
	assert.Equal(t, 100, top)
}
