package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/word"
	"go.uber.org/zap"
)

func drawLetters(available *[]rune, count int) []rune {
	if len(*available) < count {
		count = len(*available)
	}

	drawn := make([]rune, 0, count)
	for i := 0; i < count; i++ {
		index := time.Now().UnixNano() % int64(len(*available))
		drawn = append(drawn, (*available)[index])
		*available = append((*available)[:index], (*available)[index+1:]...)
	}

	return drawn
}

func rackContains(rack string, letters []request.PlacedLetter) bool {
	rackCount := map[rune]int{}
	for _, r := range rack {
		rackCount[r]++
	}
	for _, l := range letters {
		c := rune(l.Char[0])
		if rackCount[c] == 0 {
			return false
		}
		rackCount[c]--
	}
	return true
}

func updatePlayerRack(gameID string, userID int64, rack string, played []request.PlacedLetter) (string, error) {
	// 1. Retirer les lettres jouées
	for _, l := range played {
		c := rune(l.Char[0])
		i := strings.IndexRune(rack, c)
		if i == -1 {
			return "", fmt.Errorf("letter %c not in rack", c)
		}
		rack = rack[:i] + rack[i+1:]
	}

	// 2. Tirer des nouvelles lettres du sac
	var available string
	err := database.QueryRow(`SELECT available_letters FROM games WHERE id = $1`, gameID).Scan(&available)
	if err != nil {
		return "", err
	}

	drawCount := 7 - len(rack)
	availableRunes := []rune(available)
	newLetters := drawLetters(&availableRunes, drawCount)

	// 3. Mettre à jour le sac
	_, err = database.Exec(`UPDATE games SET available_letters = $1 WHERE id = $2`, string(availableRunes), gameID)
	if err != nil {
		return "", err
	}

	return rack + string(newLetters), nil
}

type Pos struct{ X, Y int }

func validatePlayerInGame(gameID string, userID int64) error {
	var dummy int
	err := database.QueryRow(`SELECT 1 FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&dummy)
	if err != nil {
		zap.L().Error("failed to validate player in game", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return errors.New("unauthorized or game not found")
	}
	return nil
}

func loadBoard(gameID string) ([15][15]string, error) {
	var boardRaw []byte
	err := database.QueryRow(`SELECT board FROM games WHERE id = $1`, gameID).Scan(&boardRaw)
	if err != nil {
		zap.L().Error("failed to load game board", zap.Error(err), zap.String("game_id", gameID))
		return [15][15]string{}, errors.New("failed to load board")
	}
	var board [15][15]string
	if err := json.Unmarshal(boardRaw, &board); err != nil {
		zap.L().Error("failed to unmarshal game board", zap.Error(err), zap.String("game_id", gameID))
		return board, errors.New("invalid board data")
	}
	return board, nil
}

func applyLetters(board *[15][15]string, letters []request.PlacedLetter) error {
	for _, l := range letters {
		if board[l.Y][l.X] != "" {
			return fmt.Errorf("cell at %d,%d already occupied", l.X, l.Y)
		}
		board[l.Y][l.X] = l.Char
	}
	return nil
}

func computeMoveScore(board [15][15]string, placed []request.PlacedLetter) int {
	isNew := make(map[Pos]bool)
	for _, l := range placed {
		isNew[Pos{l.X, l.Y}] = true
	}
	total := 0

	calcWord := func(startX, startY, dx, dy int) int {
		wordMultiplier := 1
		wordScore := 0
		x, y := startX, startY
		for x >= 0 && x < 15 && y >= 0 && y < 15 {
			letter := board[y][x]
			if letter == "" {
				break
			}
			letterScore := word.LetterValues[letter]
			if isNew[Pos{x, y}] {
				switch word.SpecialCells[[2]int{x, y}] {
				case "DL":
					letterScore *= 2
				case "TL":
					letterScore *= 3
				case "DW", "★":
					wordMultiplier *= 2
				case "TW":
					wordMultiplier *= 3
				}
			}
			wordScore += letterScore
			x += dx
			y += dy
		}
		return wordScore * wordMultiplier
	}

	seen := make(map[string]bool)
	for _, l := range placed {
		for _, dir := range [][2]int{{1, 0}, {0, 1}} {
			dx, dy := dir[0], dir[1]
			startX, startY := l.X, l.Y
			for {
				nx, ny := startX-dx, startY-dy
				if nx < 0 || ny < 0 || nx >= 15 || ny >= 15 || board[ny][nx] == "" {
					break
				}
				startX, startY = nx, ny
			}
			word := ""
			x, y := startX, startY
			for x >= 0 && x < 15 && y >= 0 && y < 15 && board[y][x] != "" {
				word += board[y][x]
				x += dx
				y += dy
			}
			if len(word) > 1 && !seen[word] {
				total += calcWord(startX, startY, dx, dy)
				seen[word] = true
			}
		}
	}
	if len(placed) == 7 {
		total += 50
	}
	return total
}

func rackPoints(rack string) int {
	pts := 0
	for _, c := range rack {
		pts += word.LetterValues[strings.ToUpper(string(c))]
	}
	return pts
}

func finishGame(tx *sql.Tx, gameID string, lastPlayerID int64) error {
	rows, err := tx.Query(
		`SELECT player_id, rack FROM game_players WHERE game_id = $1`, gameID,
	)
	if err != nil {
		zap.L().Error("failed to query game players", zap.Error(err), zap.String("game_id", gameID))
		return err
	}
	defer rows.Close()

	totalLeftover := 0
	for rows.Next() {
		var pid int64
		var rack string
		if err := rows.Scan(&pid, &rack); err != nil {
			zap.L().Error("failed to scan game player", zap.Error(err), zap.String("game_id", gameID), zap.Int64("player_id", pid))
			return err
		}
		lp := rackPoints(rack)
		// retire les points non joués
		if _, err := tx.Exec(
			`UPDATE game_players SET score = score - $1 WHERE game_id = $2 AND player_id = $3`,
			lp, gameID, pid,
		); err != nil {
			zap.L().Error("failed to update game player score", zap.Error(err), zap.String("game_id", gameID), zap.Int64("player_id", pid))
			return err
		}
		// cumule pour le bonus
		if pid != lastPlayerID {
			totalLeftover += lp
		}
	}
	// bonus pour le finisseur (si lastPlayerID != 0)
	if lastPlayerID != 0 && totalLeftover > 0 {
		if _, err := tx.Exec(
			`UPDATE game_players SET score = score + $1 WHERE game_id = $2 AND player_id = $3`,
			totalLeftover, gameID, lastPlayerID,
		); err != nil {
			zap.L().Error("failed to update last player score", zap.Error(err), zap.String("game_id", gameID), zap.Int64("last_player_id", lastPlayerID))
			return err
		}
	}

	// on marque la partie terminée, stocke le vainqueur et l'heure
	var winnerUsername sql.NullString
	if lastPlayerID != 0 {
		// on récupère le username du vainqueur
		if err := tx.QueryRow(
			`SELECT username FROM users WHERE id = $1`, lastPlayerID,
		).Scan(&winnerUsername); err != nil {
			zap.L().Error("failed to get winner username", zap.Error(err), zap.Int64("last_player_id", lastPlayerID))
			return err
		}
	}
	_, err = tx.Exec(
		`UPDATE games
           SET status           = 'ended',
               winner_username  = $1,
               ended_at         = now()
         WHERE id = $2`,
		winnerUsername.String, gameID,
	)
	if err != nil {
		zap.L().Error("failed to update game status", zap.Error(err), zap.String("game_id", gameID))
		return err
	}

	return nil
}
