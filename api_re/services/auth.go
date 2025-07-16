package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	dbModels "github.com/ZiplEix/scrabble/api/models/database"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, password string) (*dbModels.User, error) {
	fmt.Println("Creating user:", username)

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (username, password, created_at)
		VALUES ($1, $2, $3)
	`

	rows, err := database.Query(query, username, string(hashed), time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user dbModels.User
	err = database.QueryRow("SELECT id, username, created_at FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func VerifyUser(username, password string) (*dbModels.User, error) {
	var user dbModels.User
	query := `SELECT id, username, password, created_at FROM users WHERE username = $1`
	err := database.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}
