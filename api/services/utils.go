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
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/ZiplEix/scrabble/api/word"
	"go.uber.org/zap"
)

// buildBoardBlanks parcourt l'historique des coups pour connaître les positions
// où des jokers ont été posés. Utile pour le calcul de score des mots croisés.
func buildBoardBlanks(gameID string) map[Pos]bool {
	res := map[Pos]bool{}
	rows, err := database.Query(`SELECT move FROM game_moves WHERE game_id = $1 ORDER BY created_at ASC`, gameID)
	if err != nil {
		return res
	}
	defer rows.Close()
	for rows.Next() {
		var moveRaw []byte
		if err := rows.Scan(&moveRaw); err != nil {
			continue
		}
		var mv struct {
			Letters []request.PlacedLetter `json:"letters"`
		}
		if err := json.Unmarshal(moveRaw, &mv); err != nil {
			continue
		}
		for _, pl := range mv.Letters {
			if pl.Blank {
				res[Pos{pl.X, pl.Y}] = true
			}
		}
	}
	return res
}

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
		// Si c'est une lettre blanche, on consomme un joker '?'
		if l.Blank {
			if rackCount['?'] == 0 {
				return false
			}
			rackCount['?']--
			continue
		}
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
		var toRemove rune
		if l.Blank {
			toRemove = '?'
		} else {
			toRemove = rune(l.Char[0])
		}
		i := strings.IndexRune(rack, toRemove)
		if i == -1 {
			return "", fmt.Errorf("letter %c not in rack", toRemove)
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

// resolveBlanks tente d'attribuer automatiquement des jokers ('?') aux lettres manquantes
// lorsque le client n'a pas renseigné le champ Blank. Respecte aussi les Blank déjà posés.
func resolveBlanks(rack string, letters []request.PlacedLetter) ([]request.PlacedLetter, error) {
	// Compter les lettres dans le rack
	rackCount := map[rune]int{}
	for _, r := range rack {
		rackCount[r]++
	}

	// Consommer d'abord les blanks déjà marqués
	used := make([]request.PlacedLetter, len(letters))
	copy(used, letters)
	for _, l := range used {
		if l.Blank {
			if rackCount['?'] == 0 {
				return nil, fmt.Errorf("invalid move: no blank in rack")
			}
			rackCount['?']--
		}
	}

	// Pass 1: essayer de consommer les vraies lettres
	for i := range used {
		if used[i].Blank {
			continue
		}
		c := rune(used[i].Char[0])
		if rackCount[c] > 0 {
			rackCount[c]--
			continue
		}
		// sinon, poser un joker si dispo
		if rackCount['?'] > 0 {
			used[i].Blank = true
			rackCount['?']--
		} else {
			return nil, fmt.Errorf("invalid move: you don't have the required letters")
		}
	}

	return used, nil
}

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

// computeMoveScore calcule le score du coup en tenant compte des jokers.
// boardBlank indique quelles positions du plateau sont des jokers (0 point), y compris celles posées lors de coups précédents.
func computeMoveScore(board [15][15]string, placed []request.PlacedLetter, boardBlank map[Pos]bool) int {
	isNew := make(map[Pos]bool)
	isBlank := make(map[Pos]bool)
	for _, l := range placed {
		pos := Pos{l.X, l.Y}
		isNew[pos] = true
		if l.Blank {
			isBlank[pos] = true
		}
	}
	// Merge historique
	for p, b := range boardBlank {
		if b {
			isBlank[p] = true
		}
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
			// score de la lettre (0 si joker à cette position)
			letterScore := 0
			if !isBlank[Pos{x, y}] {
				letterScore = word.LetterValues[letter]
			}
			if isNew[Pos{x, y}] {
				switch word.SpecialCells[[2]int{x, y}] {
				case "DL":
					// DL ne s'applique pas aux jokers (0 reste 0)
					letterScore *= 2
				case "TL":
					// TL ne s'applique pas aux jokers
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
		if c == '?' {
			continue // joker vaut 0
		}
		pts += word.LetterValues[strings.ToUpper(string(c))]
	}
	return pts
}

// tx is a transaction that must be committed by the caller
func finishGame(tx *sql.Tx, gameID string, lastPlayerID int64) error {
	type leftover struct {
		pid int64
		lp  int
	}
	var lefts []leftover

	rows, err := tx.Query(
		`SELECT player_id, rack FROM game_players WHERE game_id = $1`,
		gameID,
	)
	if err != nil {
		return err
	}
	for rows.Next() {
		var pid int64
		var rack string
		if err := rows.Scan(&pid, &rack); err != nil {
			rows.Close()
			return err
		}
		lefts = append(lefts, leftover{pid, rackPoints(rack)})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	totalLeftover := 0
	for _, l := range lefts {
		if _, err := tx.Exec(
			`UPDATE game_players SET score = score - $1
			WHERE game_id = $2 AND player_id = $3`,
			l.lp, gameID, l.pid,
		); err != nil {
			return fmt.Errorf("failed to update game player %d: %w", l.pid, err)
		}
		if l.pid != lastPlayerID {
			totalLeftover += l.lp
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

	// determiener le vainqueur
	var winnerID int64
	var winnerScore int
	err = tx.QueryRow(
		`SELECT player_id, score
         FROM game_players
         WHERE game_id = $1
         ORDER BY score DESC
         LIMIT 1`,
		gameID,
	).Scan(&winnerID, &winnerScore)
	if err != nil {
		return fmt.Errorf("failed to determine winner: %w", err)
	}

	// Récupération du username du winner
	var winnerUsername sql.NullString
	if err := tx.QueryRow(
		`SELECT username FROM users WHERE id = $1`, winnerID,
	).Scan(&winnerUsername); err != nil {
		zap.L().Error("failed to get winner username", zap.Error(err), zap.Int64("winner_id", winnerID))
		return err
	}

	// marquer la partie comme terminée avec le nom du gagnant
	_, err = tx.Exec(
		`UPDATE games
            SET status          = 'ended',
                winner_username = $1,
                ended_at        = now()
          WHERE id = $2`,
		winnerUsername.String, gameID,
	)
	if err != nil {
		zap.L().Error("failed to update game status", zap.Error(err), zap.String("game_id", gameID))
		return err
	}

	var gameName string
	if err := tx.QueryRow(
		`SELECT name FROM games WHERE id = $1`, gameID,
	).Scan(&gameName); err != nil {
		zap.L().Error("failed to get game name", zap.Error(err), zap.String("game_id", gameID))
		return err
	}

	sendNotif := func(uid int64, payload utils.NotificationPayload) {
		go func(userID int64, pl utils.NotificationPayload) {
			if err := utils.SendNotificationToUserByID(userID, pl); err != nil {
				// si pas d'abonnement push → info, sinon warn
				if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set") {
					zap.L().Info("no push subscription for user", zap.Int64("user_id", userID))
					return
				}
				zap.L().Warn("failed to send push notification", zap.Error(err), zap.Int64("user_id", userID))
			}
		}(uid, payload)
	}

	// gagnant
	if winnerUsername.Valid {
		sendNotif(winnerID, utils.NotificationPayload{
			Title: "Vous avez gagné la partie \"" + gameName + "\"!",
			Body:  fmt.Sprintf("Félicitations %s, vous avez remporté la partie avec %d points!", winnerUsername.String, winnerScore),
			Url:   fmt.Sprintf("https://scrabble.baptiste.zip/games/%s", gameID),
		})
	}

	// autres joueurs
	for _, l := range lefts {
		if l.pid == winnerID {
			continue // skip winner
		}
		var username sql.NullString
		if err := tx.QueryRow(`SELECT username FROM users WHERE id = $1`, l.pid).Scan(&username); err != nil {
			zap.L().Error("failed to get player username", zap.Error(err), zap.Int64("player_id", l.pid))
			return err
		}
		var userPts int64
		if err := tx.QueryRow(
			`SELECT score FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, l.pid,
		).Scan(&userPts); err != nil {
			zap.L().Error("failed to get player score", zap.Error(err), zap.Int64("player_id", l.pid))
			return err
		}
		if username.Valid {
			sendNotif(l.pid, utils.NotificationPayload{
				Title: "Partie terminée: \"" + gameName + "\"",
				Body:  fmt.Sprintf("%s a gagné avec %d points!\nVous avez terminé avec %d points.", winnerUsername.String, winnerScore, userPts),
				Url:   fmt.Sprintf("https://scrabble.baptiste.zip/games/%s", gameID),
			})
		}
	}

	return nil
}
