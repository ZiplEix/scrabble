package services

import (
	"errors"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	dbModels "github.com/ZiplEix/scrabble/api/models/database"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, password string) (*dbModels.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("failed to hash password", zap.Error(err), zap.String("username", username))
		return nil, err
	}

	query := `
		INSERT INTO users (username, password, created_at)
		VALUES ($1, $2, $3)
	`

	_, err = database.Exec(query, username, string(hashed), time.Now())
	if err != nil {
		zap.L().Error("failed to create user", zap.Error(err), zap.String("username", username))
		return nil, err
	}

	var user dbModels.User
	err = database.QueryRow("SELECT id, username, created_at FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			zap.L().Error("user not found after creation", zap.String("username", username))
			return nil, errors.New("user not found")
		}
		zap.L().Error("failed to query user after creation", zap.Error(err), zap.String("username", username))
		return nil, err
	}

	return &user, nil
}

func VerifyUser(username, password string) (*dbModels.User, error) {
	var user dbModels.User
	query := `SELECT id, username, password, created_at FROM users WHERE username = $1`
	err := database.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		zap.L().Error("user not found", zap.String("username", username))
		return nil, errors.New("user not found")
	} else if err != nil {
		zap.L().Error("failed to query user", zap.Error(err), zap.String("username", username))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		zap.L().Error("password mismatch", zap.String("username", username))
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func UpdateUserPassword(username, newPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("failed to hash password", zap.Error(err), zap.String("username", username))
		return err
	}

	query := `
		UPDATE users
		SET password = $1
		WHERE username = $2
	`

	_, err = database.Exec(query, string(hashed), username)
	if err != nil {
		zap.L().Error("failed to update user password", zap.Error(err), zap.String("username", username))
		return err
	}

	return nil
}
