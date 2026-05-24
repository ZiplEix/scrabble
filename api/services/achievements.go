package services

import (
	"database/sql"
	"strings"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	"go.uber.org/zap"
)

// UnlockAchievement débloque un succès pour un utilisateur
func UnlockAchievement(userID int64, achievementID string) error {
	_, err := database.Exec(`
		INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id, achievement_id) DO NOTHING
	`, userID, achievementID)
	if err != nil {
		zap.L().Error("Failed to unlock achievement", zap.Int64("user_id", userID), zap.String("achievement_id", achievementID), zap.Error(err))
		return err
	}
	return nil
}

// CheckAndUnlockPlayMoveAchievements vérifie et débloque les succès liés aux coups posés
func CheckAndUnlockPlayMoveAchievements(userID int64, letters []request.PlacedLetter, moveScore int, word string) {
	// 1. Bingo ! (poser les 7 lettres d'un coup)
	if len(letters) == 7 {
		_ = UnlockAchievement(userID, "bingo")
	}

	// 2. Coup de Génie (marquer >= 75 points en un seul coup)
	if moveScore >= 75 {
		_ = UnlockAchievement(userID, "high_scorer")
	}

	// 3. Amoureux des Mots (placer une lettre rare K, W, X, Y, Z rapportant au moins 30 points)
	containsRare := false
	for _, l := range letters {
		char := strings.ToUpper(l.Char)
		if char == "K" || char == "W" || char == "X" || char == "Y" || char == "Z" {
			containsRare = true
			break
		}
	}
	if containsRare && moveScore >= 30 {
		_ = UnlockAchievement(userID, "word_smith")
	}

	// 4. Premier Pas (placer son tout premier mot dans une partie)
	var moveCount int
	err := database.QueryRow(`SELECT COUNT(*) FROM game_moves WHERE player_id = $1`, userID).Scan(&moveCount)
	if err == nil && moveCount == 1 {
		_ = UnlockAchievement(userID, "first_step")
	}

	// 5. L'Illusionniste (placer un mot utilisant un joker)
	hasBlank := false
	for _, l := range letters {
		if l.Blank {
			hasBlank = true
			break
		}
	}
	if hasBlank {
		_ = UnlockAchievement(userID, "joker_master")
	}

	// 6. Linguiste Émérite (poser un mot de 8 lettres ou plus)
	if len(word) >= 8 {
		_ = UnlockAchievement(userID, "long_word")
	}

	// 7. Demi-Siècle (marquer au moins 50 points en un seul coup)
	if moveScore >= 50 {
		_ = UnlockAchievement(userID, "half_century")
	}
}

// CheckAndUnlockGameFinishedAchievements vérifie et débloque les succès liés à la fin de partie
func CheckAndUnlockGameFinishedAchievements(tx *sql.Tx, gameID string, winnerID int64, playerIDs []int64) {
	// 1. Premier Sang (première victoire)
	if winnerID != 0 {
		var winCount int
		err := tx.QueryRow(`
			SELECT COUNT(*) FROM games 
			WHERE status = 'ended' AND winner_username = (SELECT username FROM users WHERE id = $1)
		`, winnerID).Scan(&winCount)
		if err == nil && winCount == 1 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'first_blood', now())
				ON CONFLICT DO NOTHING
			`, winnerID)
		}
	}

	// 11. Tueur de Géants (battre le bot Scrabby en 1vs1)
	if BotUserID != -1 && len(playerIDs) == 2 && winnerID != 0 && winnerID != BotUserID {
		hasBot := false
		for _, pid := range playerIDs {
			if pid == BotUserID {
				hasBot = true
				break
			}
		}
		if hasBot {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'bot_slayer', now())
				ON CONFLICT DO NOTHING
			`, winnerID)
		}
	}

	for _, pid := range playerIDs {
		// 2. Grand Maître (> 400 points dans la partie)
		var score int
		err := tx.QueryRow(`SELECT score FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, pid).Scan(&score)
		if err == nil && score > 400 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'scrabble_master', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 3. Marathonien (jouer 10 parties complètes)
		var completeGames int
		err = tx.QueryRow(`
			SELECT COUNT(*) FROM game_players gp
			JOIN games g ON gp.game_id = g.id
			WHERE gp.player_id = $1 AND g.status = 'ended'
		`, pid).Scan(&completeGames)
		if err == nil && completeGames >= 10 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'marathoner', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 4. Vétéran (jouer 50 parties complètes)
		if err == nil && completeGames >= 50 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'veteran', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 5. Rivalité Amicale (jouer contre au moins 3 adversaires différents)
		var opponentsCount int
		err = tx.QueryRow(`
			SELECT COUNT(DISTINCT gp2.player_id)
			FROM game_players gp1
			JOIN game_players gp2 ON gp1.game_id = gp2.game_id
			JOIN games g ON gp1.game_id = g.id
			WHERE gp1.player_id = $1 AND gp2.player_id != $1 AND g.status = 'ended'
		`, pid).Scan(&opponentsCount)
		if err == nil && opponentsCount >= 3 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'friendly_rivalry', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 6. Légende du Club (marquer plus de 500 points dans une seule partie)
		if err == nil && score > 500 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'elite_player', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 7. Série Victorieuse (remporter 3 victoires consécutives)
		var winsInARow int
		err = tx.QueryRow(`
			WITH last_games AS (
				SELECT g.id, g.winner_username, gp.player_id, u.username
				FROM games g
				JOIN game_players gp ON g.id = gp.game_id
				JOIN users u ON gp.player_id = u.id
				WHERE gp.player_id = $1 AND g.status = 'ended'
				ORDER BY g.created_at DESC
				LIMIT 3
			)
			SELECT COUNT(*) FROM last_games
			WHERE winner_username = username
		`, pid).Scan(&winsInARow)
		if err == nil && winsInARow == 3 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'serial_winner', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 8. Oiseau de Nuit (terminer une partie entre 23h et 5h du matin)
		hour := time.Now().Hour()
		if hour >= 23 || hour < 5 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'night_owl', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}

		// 9. Stratège (atteindre un classement de 500 IPS ou plus)
		var rating int
		err = tx.QueryRow(`SELECT rating FROM users WHERE id = $1`, pid).Scan(&rating)
		if err == nil && rating >= 500 {
			_, _ = tx.Exec(`
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, 'ips_master', now())
				ON CONFLICT DO NOTHING
			`, pid)
		}
	}

	// 10. Le Survivant (gagner avec moins de 10 points d'avance)
	if winnerID != 0 {
		rows, err := tx.Query(`
			SELECT score FROM game_players 
			WHERE game_id = $1 
			ORDER BY score DESC 
			LIMIT 2
		`, gameID)
		if err == nil {
			var scores []int
			for rows.Next() {
				var s int
				if err := rows.Scan(&s); err == nil {
					scores = append(scores, s)
				}
			}
			rows.Close()

			if len(scores) == 2 && (scores[0]-scores[1]) < 10 {
				_, _ = tx.Exec(`
					INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
					VALUES ($1, 'comeback_kid', now())
					ON CONFLICT DO NOTHING
				`, winnerID)
			}
		}
	}
}

// CheckAndUnlockPuzzleAchievement débloque le succès pour avoir résolu un puzzle quotidien
func CheckAndUnlockPuzzleAchievement(userID int64) {
	_ = UnlockAchievement(userID, "puzzle_solver")

	var solveCount int
	err := database.QueryRow(`SELECT COUNT(*) FROM puzzle_attempts WHERE user_id = $1 AND success = true`, userID).Scan(&solveCount)
	if err == nil {
		if solveCount >= 5 {
			_ = UnlockAchievement(userID, "sharp_mind")
		}
		if solveCount >= 15 {
			_ = UnlockAchievement(userID, "puzzle_expert")
		}
	}
}
