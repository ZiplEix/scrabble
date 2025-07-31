package utils

import (
	"os"
	"time"

	dbModels "github.com/ZiplEix/scrabble/api/models/database"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
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
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		zap.L().Error("failed to sign token", zap.Error(err), zap.String("username", user.Username))
		return "", err
	}
	return tokenString, nil

}
