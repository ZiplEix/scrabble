package main

import (
	"fmt"
	"os"

	"github.com/ZiplEix/scrabble/api/config"
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/services"
)

func main() {
	config.InitEnv()
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		fmt.Println("POSTGRES_URL environment variable is not set")
		os.Exit(1)
	}

	if err := database.Init(dsn); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🚀 Starting IPS recalculation for all users...")

	// 1. Fetch all users
	rows, err := database.Query("SELECT id, username FROM users")
	if err != nil {
		fmt.Printf("Failed to query users: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	type user struct {
		id       int64
		username string
	}
	var users []user
	for rows.Next() {
		var u user
		if err := rows.Scan(&u.id, &u.username); err != nil {
			fmt.Printf("Failed to scan user: %v\n", err)
			continue
		}
		users = append(users, u)
	}

	fmt.Printf("📊 Processing %d users...\n", len(users))

	// 2. Recalculate IPS for each user
	for _, u := range users {
		// Fetch last 10 finished games for this user
		mRows, err := database.Query(`
			SELECT 
				gp.score, 
				(g.winner_username = u.username) as is_winner
			FROM game_players gp
			JOIN games g ON gp.game_id = g.id
			JOIN users u ON gp.player_id = u.id
			WHERE gp.player_id = $1 AND g.status = 'ended'
			ORDER BY g.ended_at DESC
			LIMIT 10
		`, u.id)
		if err != nil {
			fmt.Printf("⚠️  Error fetching matches for user %s: %v\n", u.username, err)
			continue
		}

		var matches []services.MatchInfo
		for mRows.Next() {
			var m services.MatchInfo
			if err := mRows.Scan(&m.Score, &m.IsWinner); err != nil {
				fmt.Printf("Failed to scan match info: %v\n", err)
				continue
			}
			matches = append(matches, m)
		}
		mRows.Close()

		ips := services.CalculateIPS(matches)

		// 3. Update DB
		_, err = database.Exec("UPDATE users SET rating = $1 WHERE id = $2", ips, u.id)
		if err != nil {
			fmt.Printf("❌ Failed to update IPS for user %s: %v\n", u.username, err)
		} else {
			fmt.Printf("✅ %s: %d IPS (%d matches processed)\n", u.username, ips, len(matches))
		}
	}

	fmt.Println("🎉 IPS recalculation complete!")
}
