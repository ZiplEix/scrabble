package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ZiplEix/scrabble/api/config"
	"github.com/ZiplEix/scrabble/api/database"
)

type PlacedLetter struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Char  string `json:"char"`
	Blank bool   `json:"blank,omitempty"`
}

type PlayMoveRequest struct {
	Word    string         `json:"word"`
	Letters []PlacedLetter `json:"letters"`
	Score   int            `json:"score"`
}

type User struct {
	ID       int64
	Username string
	Rating   int
}

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

	fmt.Println("🏆 Starting retroactive achievements recalculation for all users...")

	// 1. Fetch all users
	rows, err := database.Query("SELECT id, username, rating FROM users")
	if err != nil {
		fmt.Printf("Failed to query users: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Rating); err != nil {
			fmt.Printf("Failed to scan user: %v\n", err)
			continue
		}
		users = append(users, u)
	}

	fmt.Printf("📊 Recalculating achievements for %d users...\n", len(users))

	for _, u := range users {
		fmt.Printf("\n👤 Processing user: %s (ID: %d)\n", u.Username, u.ID)
		unlockedCount := 0

		// Helper to unlock achievement
		unlock := func(achID string) {
			_, err := database.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, $2, now())
				ON CONFLICT (user_id, achievement_id) DO NOTHING
			`, u.ID, achID)
			if err == nil {
				fmt.Printf("   🎉 Unlocked: %s\n", achID)
				unlockedCount++
			}
		}

		// 1. first_blood (at least 1 win in ended games)
		var wins int
		err := database.QueryRow(`
			SELECT COUNT(*) FROM games 
			WHERE status = 'ended' AND winner_username = $1
		`, u.Username).Scan(&wins)
		if err == nil && wins >= 1 {
			unlock("first_blood")
		}

		// 2. puzzle_solver, sharp_mind, puzzle_expert
		var solvedPuzzles int
		err = database.QueryRow(`
			SELECT COUNT(*) FROM puzzle_attempts 
			WHERE user_id = $1 AND success = true
		`, u.ID).Scan(&solvedPuzzles)
		if err == nil {
			if solvedPuzzles >= 1 {
				unlock("puzzle_solver")
			}
			if solvedPuzzles >= 5 {
				unlock("sharp_mind")
			}
			if solvedPuzzles >= 15 {
				unlock("puzzle_expert")
			}
		}

		// 3. marathoner, veteran
		var completeGames int
		err = database.QueryRow(`
			SELECT COUNT(*) FROM game_players gp
			JOIN games g ON gp.game_id = g.id
			WHERE gp.player_id = $1 AND g.status = 'ended'
		`, u.ID).Scan(&completeGames)
		if err == nil {
			if completeGames >= 10 {
				unlock("marathoner")
			}
			if completeGames >= 50 {
				unlock("veteran")
			}
		}

		// 4. friendly_rivalry
		var opponentsCount int
		err = database.QueryRow(`
			SELECT COUNT(DISTINCT gp2.player_id)
			FROM game_players gp1
			JOIN game_players gp2 ON gp1.game_id = gp2.game_id
			JOIN games g ON gp1.game_id = g.id
			WHERE gp1.player_id = $1 AND gp2.player_id != $1 AND g.status = 'ended'
		`, u.ID).Scan(&opponentsCount)
		if err == nil && opponentsCount >= 3 {
			unlock("friendly_rivalry")
		}

		// 5. scrabble_master, elite_player
		var maxGameScore int
		err = database.QueryRow(`
			SELECT COALESCE(MAX(score), 0) FROM game_players gp
			JOIN games g ON gp.game_id = g.id
			WHERE gp.player_id = $1 AND g.status = 'ended'
		`, u.ID).Scan(&maxGameScore)
		if err == nil {
			if maxGameScore > 400 {
				unlock("scrabble_master")
			}
			if maxGameScore > 500 {
				unlock("elite_player")
			}
		}

		// 6. elo_master
		if u.Rating >= 1800 {
			unlock("elo_master")
		}

		// 7. chatty
		var messagesCount int
		err = database.QueryRow(`SELECT COUNT(*) FROM messages WHERE user_id = $1`, u.ID).Scan(&messagesCount)
		if err == nil && messagesCount >= 1 {
			unlock("chatty")
		}

		// 8. first_step
		var movesCount int
		err = database.QueryRow(`SELECT COUNT(*) FROM game_moves WHERE player_id = $1`, u.ID).Scan(&movesCount)
		if err == nil && movesCount >= 1 {
			unlock("first_step")
		}

		// 9. Fetch moves to evaluate gameplay accomplishments:
		// bingo, high_scorer, half_century, word_smith, joker_master, long_word
		mRows, err := database.Query(`SELECT move FROM game_moves WHERE player_id = $1`, u.ID)
		if err == nil {
			hasBingo := false
			hasHighScorer := false
			hasHalfCentury := false
			hasWordSmith := false
			hasJoker := false
			hasLongWord := false

			for mRows.Next() {
				var moveRaw []byte
				if err := mRows.Scan(&moveRaw); err != nil {
					continue
				}

				var req PlayMoveRequest
				if err := json.Unmarshal(moveRaw, &req); err != nil {
					continue
				}

				if len(req.Letters) == 7 {
					hasBingo = true
				}
				if req.Score >= 75 {
					hasHighScorer = true
				}
				if req.Score >= 50 {
					hasHalfCentury = true
				}
				if len(req.Word) >= 8 {
					hasLongWord = true
				}

				containsRare := false
				hasBlank := false
				for _, l := range req.Letters {
					if l.Blank {
						hasBlank = true
					}
					char := strings.ToUpper(l.Char)
					if char == "K" || char == "W" || char == "X" || char == "Y" || char == "Z" {
						containsRare = true
					}
				}

				if containsRare && req.Score >= 30 {
					hasWordSmith = true
				}
				if hasBlank {
					hasJoker = true
				}
			}
			mRows.Close()

			if hasBingo {
				unlock("bingo")
			}
			if hasHighScorer {
				unlock("high_scorer")
			}
			if hasHalfCentury {
				unlock("half_century")
			}
			if hasWordSmith {
				unlock("word_smith")
			}
			if hasJoker {
				unlock("joker_master")
			}
			if hasLongWord {
				unlock("long_word")
			}
		}

		// 10. comeback_kid (won by < 10 points difference)
		// Fetch games won by this user
		gRows, err := database.Query(`
			SELECT id FROM games 
			WHERE status = 'ended' AND winner_username = $1
		`, u.Username)
		if err == nil {
			hasComeback := false
			for gRows.Next() {
				var gameID string
				if err := gRows.Scan(&gameID); err != nil {
					continue
				}

				// Get the top 2 scores
				sRows, err := database.Query(`
					SELECT score FROM game_players 
					WHERE game_id = $1 
					ORDER BY score DESC 
					LIMIT 2
				`, gameID)
				if err == nil {
					var scores []int
					for sRows.Next() {
						var s int
						if err := sRows.Scan(&s); err == nil {
							scores = append(scores, s)
						}
					}
					sRows.Close()

					if len(scores) == 2 && (scores[0]-scores[1]) < 10 {
						hasComeback = true
						break
					}
				}
			}
			gRows.Close()

			if hasComeback {
				unlock("comeback_kid")
			}
		}

		// 11. serial_winner (3 wins in a row)
		// Fetch the user's ended games status
		sRows, err := database.Query(`
			SELECT g.winner_username
			FROM games g
			JOIN game_players gp ON g.id = gp.game_id
			WHERE gp.player_id = $1 AND g.status = 'ended'
			ORDER BY g.ended_at DESC
		`, u.ID)
		if err == nil {
			var winners []string
			for sRows.Next() {
				var w string
				if err := sRows.Scan(&w); err == nil {
					winners = append(winners, w)
				}
			}
			sRows.Close()

			hasStreak := false
			for i := 0; i <= len(winners)-3; i++ {
				if winners[i] == u.Username && winners[i+1] == u.Username && winners[i+2] == u.Username {
					hasStreak = true
					break
				}
			}
			if hasStreak {
				unlock("serial_winner")
			}
		}

		// 12. night_owl (played ended game at late hour)
		var nightGames int
		err = database.QueryRow(`
			SELECT COUNT(*) FROM games g
			JOIN game_players gp ON g.id = gp.game_id
			WHERE gp.player_id = $1 AND g.status = 'ended' 
			AND (EXTRACT(HOUR FROM g.ended_at) >= 23 OR EXTRACT(HOUR FROM g.ended_at) < 5)
		`, u.ID).Scan(&nightGames)
		if err == nil && nightGames >= 1 {
			unlock("night_owl")
		}

		fmt.Printf("   ✨ Finished. Unlocked %d new achievements for %s.\n", unlockedCount, u.Username)
	}

	fmt.Println("\n🎉 Achievements retroactive recalculation complete!")
}
