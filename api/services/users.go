package services

import (
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
)

func SuggestUsers(userID int64, query string) ([]response.SuggestUsersResponse, error) {
	rows, err := database.DB.Query(`
		SELECT id, username FROM users
		WHERE id != $1 AND LOWER(username) LIKE LOWER($2)
		LIMIT 10
	`, userID, query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []response.SuggestUsersResponse
	for rows.Next() {
		var u response.SuggestUsersResponse
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			continue
		}
		suggestions = append(suggestions, u)
	}

	return suggestions, nil
}
