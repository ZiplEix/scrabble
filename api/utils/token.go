package utils

import (
	"context"
	"os"
	"time"

	dbModels "github.com/ZiplEix/scrabble/api/models/database"
	"github.com/ZiplEix/scrabble/api/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	return []byte(secret)
}

func GenerateToken(user dbModels.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
	})

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		logger.Error(context.Background(), "failed to sign token", "error", err, "username", user.Username)
		return "", err
	}
	return tokenString, nil

}
