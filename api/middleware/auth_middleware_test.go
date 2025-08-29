package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/services"
)

func TestMain(m *testing.M) {
	// Init DB for middleware tests
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5433/scrabble_test?sslmode=disable"
	}
	_ = database.Init(dsn)

	// Default secret for JWT middleware tests
	_ = os.Setenv("JWT_SECRET", "testsecret")

	code := m.Run()
	os.Exit(code)
}

// reset FK-related tables used by services we touch (users lookups etc.)
func resetMiddlewareDeps(t *testing.T) {
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

func makeEchoCtx(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// ---------------- RequireAuth ----------------

func TestRequireAuth_NoBearer(t *testing.T) {
	c, rec := makeEchoCtx(http.MethodGet, "/x", "")
	nextCalled := false
	h := RequireAuth(func(c echo.Context) error { nextCalled = true; return nil })

	err := h(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, 200, rec.Code) // handler not written response directly
	assert.False(t, nextCalled)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	c, _ := makeEchoCtx(http.MethodGet, "/x", "")
	// Set an invalid token
	c.Request().Header.Set("Authorization", "Bearer not_a_jwt")
	nextCalled := false
	h := RequireAuth(func(c echo.Context) error { nextCalled = true; return nil })

	err := h(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.False(t, nextCalled)
}

func TestRequireAuth_ValidToken_SetsUserID_AndCallsNext(t *testing.T) {
	c, _ := makeEchoCtx(http.MethodGet, "/x", "")

	secret := []byte(os.Getenv("JWT_SECRET"))
	userID := int64(42)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
	})
	tokenStr, err := token.SignedString(secret)
	require.NoError(t, err)

	c.Request().Header.Set("Authorization", "Bearer "+tokenStr)

	var gotUserID any
	nextCalled := false
	h := RequireAuth(func(c echo.Context) error {
		nextCalled = true
		gotUserID = c.Get(UserIDKey)
		return nil
	})

	err = h(c)
	require.NoError(t, err)
	assert.True(t, nextCalled)
	assert.EqualValues(t, userID, gotUserID)
}

// ---------------- RequireAdmin ----------------

func TestRequireAdmin_NoUserInContext(t *testing.T) {
	c, _ := makeEchoCtx(http.MethodGet, "/admin", "")
	nextCalled := false
	h := RequireAdmin(func(c echo.Context) error { nextCalled = true; return nil })

	err := h(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.False(t, nextCalled)
}

func TestRequireAdmin_UserNotAdmin(t *testing.T) {
	resetMiddlewareDeps(t)
	u, err := services.CreateUser("normal_user", "pwd")
	require.NoError(t, err)

	c, _ := makeEchoCtx(http.MethodGet, "/admin", "")
	c.Set(UserIDKey, u.ID)

	nextCalled := false
	h := RequireAdmin(func(c echo.Context) error { nextCalled = true; return nil })

	err = h(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
	assert.False(t, nextCalled)
}

func TestRequireAdmin_AdminOK(t *testing.T) {
	resetMiddlewareDeps(t)
	u, err := services.CreateUser("admin_user", "pwd")
	require.NoError(t, err)
	// Promote to admin
	_, err = database.Exec("UPDATE users SET role = 'admin' WHERE id = $1", u.ID)
	require.NoError(t, err)

	c, _ := makeEchoCtx(http.MethodGet, "/admin", "")
	c.Set(UserIDKey, u.ID)

	nextCalled := false
	h := RequireAdmin(func(c echo.Context) error { nextCalled = true; return nil })

	err = h(c)
	require.NoError(t, err)
	assert.True(t, nextCalled)
}
