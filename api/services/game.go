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
	"go.uber.org/zap"
)

// Plateau vide 15x15
func initEmptyBoard() [15][15]string {
	return [15][15]string{}
}

// Lettres classiques du scrabble français
const initialLetters = "AAAAAAAAAEEEEEEEEEEEEIIIIIIIIONNNNNNRRRRRRTTTTTTLLLLSSSSUDDDGGGMMMBBCCPPFFHHVVJQKWXYZ"

func CreateGame(userID int64, name string, usernames []string, revangeFrom *string) (*uuid.UUID, error) {
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

	// Si une partie d'origine est fournie pour une revanche, vérifier que
	// l'utilisateur courant est bien le créateur de cette partie.
	if revangeFrom != nil {
		var srcCreatedBy int64
		err := database.QueryRow(`SELECT created_by FROM games WHERE id = $1`, *revangeFrom).Scan(&srcCreatedBy)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("source game not found")
			}
			return nil, err
		}
		if srcCreatedBy != userID {
			return nil, fmt.Errorf("only the creator of the original game can create a rematch")
		}
	}

	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			zap.L().Error("Failed to rollback transaction", zap.Error(err), zap.String("game_id", gameID.String()))
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
			zap.L().Error("Failed to rollback transaction", zap.Error(err), zap.String("game_id", gameID))
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
       SELECT id, name, board, available_letters,
			 current_turn, status, created_by,
			 winner_username, ended_at
       FROM games
       WHERE id = $1
    `
	var (
		boardJSON      []byte
		avail          string
		game           response.GameInfo
		createdBy      int64
		winnerUsername sql.NullString
		endedAt        sql.NullTime
	)
	err = database.QueryRow(gameQuery, gameID).Scan(
		&game.ID, &game.Name, &boardJSON, &avail,
		&game.CurrentTurn, &game.Status, &createdBy,
		&winnerUsername, &endedAt,
	)
	if err != nil {
		return nil, err
	}
	game.RemainingLetters = len(avail)
	_ = json.Unmarshal(boardJSON, &game.Board)

	// transfert dans le DTO
	game.IsYourGame = (createdBy == userID)
	if winnerUsername.Valid {
		game.WinnerUsername = winnerUsername.String
	}
	if endedAt.Valid {
		game.EndedAt = &endedAt.Time
	}

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
		zap.L().Error("failed to get current turn", zap.Error(err), zap.String("game_id", gameID))
		return fmt.Errorf("game not found")
	}
	if currentTurn != userID {
		zap.L().Error("not your turn", zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return fmt.Errorf("not your turn")
	}

	// 2. Récupération du rack
	var rack string
	err = database.QueryRow(`SELECT rack FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&rack)
	if err != nil {
		zap.L().Error("failed to get player rack", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return fmt.Errorf("player not in game")
	}

	if !rackContains(rack, req.Letters) {
		zap.L().Error("invalid move: player does not have required letters", zap.String("rack", rack), zap.Any("letters", req.Letters), zap.String("game_id", gameID))
		return fmt.Errorf("invalid move: you don't have the required letters")
	}
	if len(req.Letters) == 0 {
		zap.L().Error("no letters provided in move request", zap.String("game_id", gameID))
		return fmt.Errorf("no letters provided")
	} else if len(req.Letters) > 7 {
		zap.L().Error("cannot place more than 7 letters in one move", zap.Int("letters_count", len(req.Letters)), zap.String("game_id", gameID))
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
		zap.L().Error("letters must be aligned in the same row or column", zap.Any("letters", req.Letters), zap.String("game_id", gameID))
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
			zap.L().Error("first move must cover the center cell", zap.Any("letters", req.Letters), zap.String("game_id", gameID))
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
			zap.L().Error("word must connect to existing letters", zap.Any("letters", req.Letters), zap.String("game_id", gameID))
			return fmt.Errorf("word must connect to existing letters")
		}
	}

	// 6. Appliquer les lettres
	if err := applyLetters(&board, req.Letters); err != nil {
		zap.L().Error("failed to apply letters to board", zap.Error(err), zap.Any("letters", req.Letters), zap.String("game_id", gameID))
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
			zap.L().Error("invalid word played", zap.String("word", w), zap.String("game_id", gameID))
			return fmt.Errorf("invalid word played: %s", w)
		}
	}

	// 8. Recalcul rack et score
	newRack, err := updatePlayerRack(gameID, userID, rack, req.Letters)
	if err != nil {
		zap.L().Error("failed to update player rack", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return fmt.Errorf("failed to update rack: %v", err)
	}
	moveScore := computeMoveScore(board, req.Letters)

	// 9. Enregistrement du coup
	moveJSON, _ := json.Marshal(req)
	_, err = database.Exec(`INSERT INTO game_moves (game_id, player_id, move) VALUES ($1, $2, $3)`, gameID, userID, moveJSON)
	if err != nil {
		zap.L().Error("failed to insert move", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return fmt.Errorf("failed to insert move: %v", err)
	}

	// 10. Mise à jour transactionnelle et met le pass_count à 0
	tx, err := database.DB.BeginTx(context.Background(), nil)
	if err != nil {
		zap.L().Error("failed to begin transaction", zap.Error(err), zap.String("game_id", gameID))
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			zap.L().Error("failed to rollback transaction", zap.Error(err), zap.String("game_id", gameID))
		}
	}()

	newBoardJSON, _ := json.Marshal(board)
	_, err = tx.Exec(`UPDATE games SET board = $1, pass_count = 0 WHERE id = $2`, newBoardJSON, gameID)
	if err != nil {
		zap.L().Error("failed to update game board", zap.Error(err), zap.String("game_id", gameID))
		return err
	}
	_, err = tx.Exec(`UPDATE game_players SET rack = $1, score = score + $2 WHERE game_id = $3 AND player_id = $4`, newRack, moveScore, gameID, userID)
	if err != nil {
		zap.L().Error("failed to update game player", zap.Error(err), zap.String("game_id", gameID), zap.Int64("user_id", userID))
		return err
	}

	var currentPosition int
	err = tx.QueryRow(`SELECT position FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, userID).Scan(&currentPosition)
	if err != nil {
		zap.L().Error("failed to get player position", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return err
	}
	var nextPlayerID int64
	err = tx.QueryRow(`SELECT player_id FROM game_players WHERE game_id = $1 AND position = (($2 + 1) % (SELECT COUNT(*) FROM game_players WHERE game_id = $1))`, gameID, currentPosition).Scan(&nextPlayerID)
	if err != nil {
		zap.L().Error("failed to get next player ID", zap.Error(err), zap.Int64("user_id", userID), zap.String("game_id", gameID))
		return err
	}

	// Si le rack du joueur est vide ET que le sac est vide, on termine la partie
	var bag string
	if err := tx.QueryRow(
		`SELECT available_letters FROM games WHERE id = $1`, gameID,
	).Scan(&bag); err != nil {
		zap.L().Error("failed to get available letters", zap.Error(err), zap.String("game_id", gameID))
		return err
	}
	if len(newRack) == 0 && len(bag) == 0 {
		if err := finishGame(tx, gameID, userID); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			zap.L().Error("failed to commit transaction after finishGame", zap.Error(err), zap.String("game_id", gameID))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Sinon, passe au joueur suivant
	_, err = tx.Exec(`UPDATE games SET current_turn = $1 WHERE id = $2`, nextPlayerID, gameID)
	if err != nil {
		zap.L().Error("failed to update current turn", zap.Error(err), zap.Int64("next_player_id", nextPlayerID), zap.String("game_id", gameID))
		return err
	}

	if err := tx.Commit(); err != nil {
		zap.L().Error("failed to commit transaction", zap.Error(err), zap.String("game_id", gameID))
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
		Url:   fmt.Sprintf("https://scrabble.baptiste.zip/games/%s", gameID),
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
			zap.L().Error("Failed to rollback transaction", zap.Error(err), zap.String("game_id", gameID))
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
			g.status,
			g.current_turn,
			COALESCE(g.winner_username, '') AS winner_username,
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
		err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Status,
			&g.CurrentTurnUserID,
			&g.WinnerUsername,
			&g.CurrentTurnUsername,
			&g.LastPlayTime,
			&g.IsYourGame,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game row: %w", err)
		}
		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over game rows: %w", err)
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
	ctx := context.Background()

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			zap.L().Error("rollback failed", zap.Error(rbErr), zap.String("game_id", gameID))
		}
	}()

	// Vérifie le tour (et verrouille la ligne game)
	var currentTurn int64
	if err := tx.QueryRowContext(ctx,
		`SELECT current_turn FROM games WHERE id = $1 FOR UPDATE`, gameID,
	).Scan(&currentTurn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("game not found")
		}
		return err
	}
	if currentTurn != userID {
		return errors.New("not your turn")
	}

	// Enregistre le "pass" DANS la transaction
	passMove := map[string]any{"type": "pass"}
	moveJSON, _ := json.Marshal(passMove)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO game_moves (game_id, player_id, move) VALUES ($1, $2, $3)`,
		gameID, userID, moveJSON,
	); err != nil {
		return errors.New("failed to record pass")
	}

	// Récupère position du joueur
	var position int
	if err := tx.QueryRowContext(ctx,
		`SELECT position FROM game_players WHERE game_id = $1 AND player_id = $2`,
		gameID, userID,
	).Scan(&position); err != nil {
		return err
	}

	// Détermine prochain joueur
	var nextPlayer int64
	if err := tx.QueryRowContext(ctx,
		`SELECT player_id FROM game_players
		  WHERE game_id = $1
		    AND position = (($2 + 1) % (SELECT COUNT(*) FROM game_players WHERE game_id = $1))`,
		gameID, position,
	).Scan(&nextPlayer); err != nil {
		return err
	}

	// Incrémente pass_count ET met à jour le tour courant, en récupérant la valeur
	var passCount int
	if err := tx.QueryRowContext(ctx,
		`UPDATE games
		    SET pass_count = pass_count + 1,
		        current_turn = $2
		  WHERE id = $1
		  RETURNING pass_count`,
		gameID, nextPlayer,
	).Scan(&passCount); err != nil {
		return err
	}

	// Nombre de joueurs
	var playerCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1`, gameID,
	).Scan(&playerCount); err != nil {
		return err
	}

	// Fin de partie ? (2 passes * nb joueurs)
	if passCount >= playerCount*2 {
		if err := finishGame(tx, gameID, 0); err != nil {
			return err
		}
		// IMPORTANT: commit après finishGame
		return tx.Commit()
	}

	return tx.Commit()
}
