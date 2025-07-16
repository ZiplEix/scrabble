package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

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

		// Injecte l'ID utilisateur dans le contexte Echo
		c.Set(UserIDKey, userID)

		return next(c)
	}
}
