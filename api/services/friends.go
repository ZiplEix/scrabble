package services

import (
	"context"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
)

// AddFriend adds a user to the viewer's friend list
func AddFriend(userID, friendID int64) error {
	_, err := database.DB.ExecContext(context.Background(), `
		INSERT INTO user_friends (user_id, friend_id, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id, friend_id) DO NOTHING
	`, userID, friendID)
	return err
}

// RemoveFriend removes a user from the viewer's friend list
func RemoveFriend(userID, friendID int64) error {
	_, err := database.DB.ExecContext(context.Background(), `
		DELETE FROM user_friends
		WHERE user_id = $1 AND friend_id = $2
	`, userID, friendID)
	return err
}

// GetFriends returns the list of friends for a user
func GetFriends(userID int64) ([]response.FriendResponse, error) {
	rows, err := database.DB.Query(`
		SELECT u.id, u.username, u.rating, u.role
		FROM user_friends uf
		JOIN users u ON uf.friend_id = u.id
		WHERE uf.user_id = $1
		ORDER BY u.username ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []response.FriendResponse
	for rows.Next() {
		var f response.FriendResponse
		if err := rows.Scan(&f.ID, &f.Username, &f.Rating, &f.Role); err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}

	if friends == nil {
		friends = []response.FriendResponse{}
	}
	return friends, nil
}

// GetRecentOpponents returns up to 10 players the user has played with recently
func GetRecentOpponents(userID int64) ([]response.FriendResponse, error) {
	rows, err := database.DB.Query(`
		SELECT u.id, u.username, u.rating, u.role
		FROM (
			SELECT gp2.player_id, MAX(g.created_at) as last_played
			FROM game_players gp1
			JOIN game_players gp2 ON gp1.game_id = gp2.game_id AND gp2.player_id != gp1.player_id
			JOIN games g ON gp1.game_id = g.id
			WHERE gp1.player_id = $1
			GROUP BY gp2.player_id
		) sub
		JOIN users u ON sub.player_id = u.id
		ORDER BY sub.last_played DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var opponents []response.FriendResponse
	for rows.Next() {
		var o response.FriendResponse
		if err := rows.Scan(&o.ID, &o.Username, &o.Rating, &o.Role); err != nil {
			return nil, err
		}
		opponents = append(opponents, o)
	}

	if opponents == nil {
		opponents = []response.FriendResponse{}
	}
	return opponents, nil
}
