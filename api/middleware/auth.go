package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/pkg/logger"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type userIDCtxKey struct{}
type usernameCtxKey struct{}

var (
	userIDKey   = userIDCtxKey{}
	usernameKey = usernameCtxKey{}
)

func init() {
	logger.RegisterContextKey(userIDKey, "user_id")
	logger.RegisterContextKey(usernameKey, "username")
}

const UserIDKey = "user_id"

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	secret := []byte(os.Getenv("JWT_SECRET"))

	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no auth bearer token")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("unauthorized, non valid token: %v", err))
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token payload")
		}
		userID := int64(userIDFloat)

		username, _ := claims["username"].(string)

		// Injecte l'ID utilisateur dans le contexte Echo
		c.Set(UserIDKey, userID)
		if username != "" {
			c.Set("username", username)
		}

		// Injecte dans le context.Context standard de la requête HTTP
		req := c.Request()
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		if username != "" {
			ctx = context.WithValue(ctx, usernameKey, username)
		}
		c.SetRequest(req.WithContext(ctx))

		return next(c)
	}
}

func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := utils.GetUserID(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no user_id in context")
		}

		var role string
		err := database.DB.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "failed to check user role")
		}

		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin access required")
		}

		return next(c)
	}
}
