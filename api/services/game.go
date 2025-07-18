package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/ZiplEix/scrabble/api/word"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Plateau vide 15x15
func initEmptyBoard() [15][15]string {
	return [15][15]string{}
}

// Lettres classiques du scrabble français
const initialLetters = "AAAAAAAAAEEEEEEEEEEEEIIIIIIIIONNNNNNRRRRRRTTTTTTLLLLSSSSUDDDGGGMMMBBCCPPFFHHVVJQKWXYZ"

func CreateGame(userID int64, name string, usernames []string) (*uuid.UUID, error) {
	board := initEmptyBoard()

	available := []rune(initialLetters)
	utils.ShuffleRunes(available)

	rack := utils.DrawLetters(&available, 7)

	boardJSON, err := json.Marshal(board)
	if err != nil {
		return nil, err
	}

	availableStr := string(available)
	gameID := uuid.New()

	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("Failed to rollback transaction: %v\n", err)
		}
	}()

	// Création du jeu
	_, err = tx.Exec(`
		INSERT INTO games (id, name, created_by, current_turn, board, available_letters, created_at)
		VALUES ($1, $2, $3, $3, $4, $5, $6)
	`, gameID, name, userID, boardJSON, availableStr, time.Now())
	if err != nil {
		return nil, err
	}

	// Ajouter le créateur en premier
	rackStr := string(rack)
	_, err = tx.Exec(`
		INSERT INTO game_players (game_id, player_id, rack, position, score)
		VALUES ($1, $2, $3, 0, 0)
	`, gameID, userID, rackStr)
	if err != nil {
		return nil, err
	}

	// Récupérer les IDs des autres joueurs
	if len(usernames) > 0 {
		query := `SELECT id, username FROM users WHERE username = ANY($1)`
		rows, err := tx.Query(query, pq.Array(usernames))
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		type player struct {
			id   int64
			name string
		}
		var otherPlayers []player

		for rows.Next() {
			var otherID int64
			var uname string
			if err := rows.Scan(&otherID, &uname); err != nil {
				return nil, err
			}
			otherPlayers = append(otherPlayers, player{otherID, uname})
		}

		// 💡 Ici seulement on fait les autres tx.Exec
		position := 1
		for _, p := range otherPlayers {
			rack := drawLetters(&available, 7)
			rackStr := string(rack)

			_, err := tx.Exec(`
				INSERT INTO game_players (game_id, player_id, rack, position, score)
				VALUES ($1, $2, $3, $4, 0)
			`, gameID, p.id, rackStr, position)
			if err != nil {
				return nil, err
			}
			position++
		}
	}

	// Mise à jour du sac de lettres
	_, err = tx.Exec(`
		UPDATE games SET available_letters = $1 WHERE id = $2
	`, string(available), gameID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &gameID, nil
}

func DeleteGame(userID int64, gameID string) error {
	var createdBy int64
	err := database.QueryRow(`SELECT created_by FROM games WHERE id = $1`, gameID).Scan(&createdBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("game not found")
		}
		return err
	}

	if createdBy != userID {
		return fmt.Errorf("unauthorized: you are not the creator of the game")
	}

	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("Failed to rollback transaction: %v\n", err)
		}
	}()

	if _, err := tx.Exec(`DELETE FROM game_moves WHERE game_id = $1`, gameID); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM game_players WHERE game_id = $1`, gameID); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM games WHERE id = $1`, gameID); err != nil {
		return err
	}

	return tx.Commit()
}

func RenameGame(userID int64, gameID string, newName string) error {
	var createdBy int64
	err := database.QueryRow(`SELECT created_by FROM games WHERE id = $1`, gameID).Scan(&createdBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("game not found")
		}
		return err
	}

	if createdBy != userID {
		return fmt.Errorf("unauthorized: you are not the creator of the game")
	}

	_, err = database.Query(`UPDATE games SET name = $1 WHERE id = $2`, newName, gameID)
	if err != nil {
		return fmt.Errorf("failed to rename game: %w", err)
	}

	return nil
}

func GetGameDetails(userID int64, gameID string) (*response.GameInfo, error) {
	// 1. Vérifie que le joueur participe
	checkQuery := `
		SELECT 1 FROM game_players WHERE game_id = $1 AND player_id = $2
	`
	var dummy int
	err := database.QueryRow(checkQuery, gameID, userID).Scan(&dummy)
	if err != nil {
		return nil, errors.New("unauthorized or game not found")
	}

	// 2. Récupère info partie
	gameQuery := `
		SELECT id, name, board, available_letters, current_turn, status
		FROM games
		WHERE id = $1
	`
	var (
		boardJSON []byte
		avail     string
		game      response.GameInfo
	)
	err = database.QueryRow(gameQuery, gameID).Scan(
		&game.ID, &game.Name, &boardJSON, &avail, &game.CurrentTurn, &game.Status,
	)
	if err != nil {
		return nil, err
	}
	game.RemainingLetters = len(avail)
	_ = json.Unmarshal(boardJSON, &game.Board)

	// 3. Récupère ton rack
	err = database.QueryRow(`SELECT rack FROM game_players WHERE game_id=$1 AND player_id=$2`,
		gameID, userID).Scan(&game.YourRack)
	if err != nil {
		return nil, err
	}

	// 4. Récupère les joueurs
	playerRows, err := database.Query(`
		SELECT gp.player_id, u.username, gp.score, gp.position
		FROM game_players gp
		JOIN users u ON gp.player_id = u.id
		WHERE gp.game_id = $1
		ORDER BY gp.position
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer playerRows.Close()

	for playerRows.Next() {
		var p response.PlayerInfo
		err := playerRows.Scan(&p.ID, &p.Username, &p.Score, &p.Position)
		if err != nil {
			return nil, err
		}
		if p.ID == game.CurrentTurn {
			game.CurrentTurnName = p.Username
		}
		game.Players = append(game.Players, p)
	}

	// 5. Récupère l’historique des coups
	moveRows, err := database.Query(`
		SELECT player_id, move, created_at
		FROM game_moves
		WHERE game_id = $1
		ORDER BY created_at ASC
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer moveRows.Close()

	for moveRows.Next() {
		var mv response.MoveInfo
		var moveJSON []byte
		if err := moveRows.Scan(&mv.PlayerID, &moveJSON, &mv.PlayedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(moveJSON, &mv.Move)
		game.Moves = append(game.Moves, mv)
	}

	return &game, nil
}

func PlayMove(gameID string, userID int64, req request.PlayMoveRequest) error {
	// 1. Vérification de l'appartenance au jeu et du tour
	if err := validatePlayerInGame(gameID, userID); err != nil {
		return err
	}
	var currentTurn int64
	err := database.QueryRow(`SELECT current_turn FROM games WHERE id = $1`, gameID).Scan(&currentTurn)
	if err != nil {
		return fmt.Errorf("game not found")
	}
	if currentTurn != userID {
		return fmt.Errorf("not your turn")
	}

	// 2. Récupération du rack
	var rack string
	err = database.QueryRow(`SELECT rack FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&rack)
	if err != nil {
		return fmt.Errorf("player not in game")
	}

	if !rackContains(rack, req.Letters) {
		return fmt.Errorf("invalid move: you don't have the required letters")
	}
	if len(req.Letters) == 0 {
		return fmt.Errorf("no letters provided")
	} else if len(req.Letters) > 7 {
		return fmt.Errorf("cannot place more than 7 letters in one move")
	}

	// 3. Vérification alignement
	sameRow := true
	sameCol := true
	firstX := req.Letters[0].X
	firstY := req.Letters[0].Y
	for _, l := range req.Letters {
		if l.X != firstX {
			sameCol = false
		}
		if l.Y != firstY {
			sameRow = false
		}
	}
	if !sameRow && !sameCol {
		return fmt.Errorf("letters must be aligned in the same row or column")
	}

	// 4. Chargement du plateau
	board, err := loadBoard(gameID)
	if err != nil {
		return err
	}

	// 5. Vérification placement (centre ou contact)
	isFirstMove := true
	for y := 0; y < 15 && isFirstMove; y++ {
		for x := 0; x < 15 && isFirstMove; x++ {
			if board[y][x] != "" {
				isFirstMove = false
			}
		}
	}

	if isFirstMove {
		found := false
		for _, l := range req.Letters {
			if l.X == 7 && l.Y == 7 {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("first move must cover the center cell")
		}
	} else {
		touchesExisting := false
		for _, l := range req.Letters {
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				x, y := l.X+d[0], l.Y+d[1]
				if x >= 0 && x < 15 && y >= 0 && y < 15 && board[y][x] != "" {
					touchesExisting = true
					break
				}
			}
			if touchesExisting {
				break
			}
		}
		if !touchesExisting {
			return fmt.Errorf("word must connect to existing letters")
		}
	}

	// 6. Appliquer les lettres
	if err := applyLetters(&board, req.Letters); err != nil {
		return err
	}

	// 7. Validation des mots
	letterMap := make(map[Pos]string)
	for _, l := range req.Letters {
		letterMap[Pos{l.X, l.Y}] = l.Char
	}
	visited := make(map[Pos]bool)
	words := []string{}
	dirs := []struct{ dx, dy int }{{1, 0}, {0, 1}}
	for _, l := range req.Letters {
		for _, dir := range dirs {
			startX, startY := l.X, l.Y
			for {
				nx, ny := startX-dir.dx, startY-dir.dy
				if nx < 0 || nx >= 15 || ny < 0 || ny >= 15 || board[ny][nx] == "" {
					break
				}
				startX, startY = nx, ny
			}
			word := ""
			hasAtLeastTwo := false
			x, y := startX, startY
			for x >= 0 && x < 15 && y >= 0 && y < 15 {
				letter := board[y][x]
				if letter == "" {
					break
				}
				word += letter
				if _, ok := letterMap[Pos{x, y}]; ok {
					hasAtLeastTwo = true
				}
				x += dir.dx
				y += dir.dy
			}
			if len(word) > 1 && hasAtLeastTwo {
				pos := Pos{startX, startY}
				if !visited[pos] {
					words = append(words, word)
					visited[pos] = true
				}
			}
		}
	}
	for _, w := range words {
		if !word.WordExists(w) {
			return fmt.Errorf("invalid word played: %s", w)
		}
	}

	// 8. Recalcul rack et score
	newRack, err := updatePlayerRack(gameID, userID, rack, req.Letters)
	if err != nil {
		return fmt.Errorf("failed to update rack: %v", err)
	}
	moveScore := computeMoveScore(board, req.Letters)

	// 9. Enregistrement du coup
	moveJSON, _ := json.Marshal(req)
	_, err = database.Exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ($1, $2, $3)`, gameID, userID, moveJSON)
	if err != nil {
		return fmt.Errorf("failed to insert move")
	}

	// 10. Mise à jour transactionnelle et met le pass_count à 0
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("Failed to rollback transaction: %v\n", err)
		}
	}()

	newBoardJSON, _ := json.Marshal(board)
	_, err = tx.Exec(`UPDATE games SET board = $1, pass_count = 0 WHERE id = $2`, newBoardJSON, gameID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE game_players SET rack = $1, score = score + $2 WHERE game_id = $3 AND player_id = $4`, newRack, moveScore, gameID, userID)
	if err != nil {
		return err
	}

	var currentPosition int
	err = tx.QueryRow(`SELECT position FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&currentPosition)
	if err != nil {
		return err
	}
	var nextPlayerID int64
	err = tx.QueryRow(`SELECT player_id FROM game_players WHERE game_id = $1 AND position = (($2 + 1) % (SELECT COUNT(*) FROM game_players WHERE game_id = $1))`, gameID, currentPosition).Scan(&nextPlayerID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE games SET current_turn = $1 WHERE id = $2`, nextPlayerID, gameID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	var username, gameName string
	err = database.QueryRow(`SELECT username FROM users WHERE id = $1`, userID).Scan(&username)
	if err != nil {
		username = "Un joueur"
	}
	err = database.QueryRow(`SELECT name FROM games WHERE id = $1`, gameID).Scan(&gameName)
	if err != nil {
		gameName = "une partie"
	}

	pluralSuffix := ""
	if moveScore != 1 {
		pluralSuffix = "s"
	}

	notificationPayload := utils.NotificationPayload{
		Title: "C'est à toi de jouer !",
		Body:  fmt.Sprintf("%s a joué un coup à %d point%s dans %s", username, moveScore, pluralSuffix, gameName),
		Url:   fmt.Sprintf("https://scrabble.baptiste.zip/game/%s", gameID),
	}
	_ = utils.SendNotificationToUserByID(nextPlayerID, notificationPayload)

	return nil
}

func GetNewRack(userID int64, gameID string) ([]string, error) {
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("Failed to rollback transaction: %v\n", err)
		}
	}()

	// Vérifier que c'est au tour du joueur
	var currentTurn int64
	err = tx.QueryRow(`SELECT current_turn FROM games WHERE id = $1`, gameID).Scan(&currentTurn)
	if err != nil {
		return nil, fmt.Errorf("game not found")
	}
	if currentTurn != userID {
		return nil, fmt.Errorf("not your turn")
	}

	// Récupérer l'ancien rack du joueur
	var oldRack string
	err = tx.QueryRow(`SELECT rack FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&oldRack)
	if err != nil {
		return nil, fmt.Errorf("player not in game")
	}

	// Récupérer le sac actuel
	var bag string
	err = tx.QueryRow(`SELECT available_letters FROM games WHERE id = $1`, gameID).Scan(&bag)
	if err != nil {
		return nil, fmt.Errorf("failed to load bag")
	}

	newRack, updatedBag := utils.DrawLettersFromString(bag, 7)
	if len(newRack) == 0 {
		return nil, fmt.Errorf("no letters left in the bag")
	}

	newBag := updatedBag + oldRack

	// Update le rack du joueur
	_, err = tx.Exec(`
		UPDATE game_players SET rack = $1
		WHERE game_id = $2 AND player_id = $3
	`, strings.Join(newRack, ""), gameID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update rack")
	}

	// Update le sac
	_, err = tx.Exec(`UPDATE games SET available_letters = $1 WHERE id = $2`, newBag, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to update bag")
	}

	// Calculer le joueur suivant
	var currentPosition int
	err = tx.QueryRow(`
		SELECT position FROM game_players
		WHERE game_id = $1 AND player_id = $2
	`, gameID, userID).Scan(&currentPosition)
	if err != nil {
		return nil, err
	}

	var nextPlayerID int64
	err = tx.QueryRow(`
		SELECT player_id FROM game_players
		WHERE game_id = $1 AND position = (
			($2 + 1) % (SELECT COUNT(*) FROM game_players WHERE game_id = $1)
		)
	`, gameID, currentPosition).Scan(&nextPlayerID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`UPDATE games SET current_turn = $1 WHERE id = $2`, nextPlayerID, gameID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newRack, nil
}

func GetGamesByUserID(userID int64) ([]response.GameSummary, error) {
	query := `
		SELECT
			g.id,
			g.name,
			g.current_turn,
			u.username,
			COALESCE((
				SELECT MAX(created_at)
				FROM game_moves
				WHERE game_id = g.id
			), g.created_at) AS last_play_time,
			(g.created_by = $1) AS is_your_game
		FROM games g
		JOIN users u ON u.id = g.current_turn
		JOIN game_players gp ON gp.game_id = g.id
		WHERE gp.player_id = $1
		AND g.status = 'ongoing'
		ORDER BY last_play_time DESC
	`

	rows, err := database.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []response.GameSummary

	for rows.Next() {
		var g response.GameSummary
		err := rows.Scan(&g.ID, &g.Name, &g.CurrentTurnUserID, &g.CurrentTurnUsername, &g.LastPlayTime, &g.IsYourGame)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game row: %w", err)
		}
		games = append(games, g)
	}

	return games, nil
}

func SimulateScore(gameID string, userID int64, letters []request.PlacedLetter) (int, error) {
	if len(letters) == 0 {
		return 0, nil
	}
	if err := validatePlayerInGame(gameID, userID); err != nil {
		return 0, err
	}
	board, err := loadBoard(gameID)
	if err != nil {
		return 0, err
	}
	if err := applyLetters(&board, letters); err != nil {
		return 0, err
	}
	return computeMoveScore(board, letters), nil
}

func PassTurn(userID int64, gameID string) error {
	// Vérifie que c’est bien à ce joueur de jouer
	var currentTurn int64
	err := database.QueryRow(`SELECT current_turn FROM games WHERE id = $1`, gameID).Scan(&currentTurn)
	if err != nil {
		return errors.New("game not found")
	}
	if currentTurn != userID {
		return errors.New("not your turn")
	}

	// Enregistre un "coup" vide
	passMove := map[string]any{"type": "pass"}
	moveJSON, _ := json.Marshal(passMove)
	_, err = database.Exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ($1, $2, $3)`, gameID, userID, moveJSON)
	if err != nil {
		return errors.New("failed to record pass")
	}

	// Commence la transaction pour changer de tour et mettre à jour le compteur
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Incrémente pass_count
	_, err = tx.Exec(`UPDATE games SET pass_count = pass_count + 1 WHERE id = $1`, gameID)
	if err != nil {
		return err
	}

	// Récupère position du joueur
	var position int
	err = tx.QueryRow(`SELECT position FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&position)
	if err != nil {
		return err
	}

	// Détermine prochain joueur
	var nextPlayer int64
	err = tx.QueryRow(`
		SELECT player_id FROM game_players
		WHERE game_id = $1 AND position = (($2 + 1) % (SELECT COUNT(*) FROM game_players WHERE game_id = $1))
	`, gameID, position).Scan(&nextPlayer)
	if err != nil {
		return err
	}

	// Met à jour le joueur courant
	_, err = tx.Exec(`UPDATE games SET current_turn = $1 WHERE id = $2`, nextPlayer, gameID)
	if err != nil {
		return err
	}

	// Vérifie si la partie est terminée (ex : 2 passes consécutives pour 2 joueurs, ou plus si tu veux)
	var passCount int
	err = tx.QueryRow(`SELECT pass_count FROM games WHERE id = $1`, gameID).Scan(&passCount)
	if err != nil {
		return err
	}
	if passCount >= 2*2 { // 2 passes chacun dans une partie à 2
		_, err = tx.Exec(`UPDATE games SET status = 'ended' WHERE id = $1`, gameID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
