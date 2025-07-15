package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/google/uuid"
)

// Plateau vide 15x15
func initEmptyBoard() [15][15]string {
	return [15][15]string{}
}

// Lettres classiques du scrabble français
const initialLetters = "AAAAAAAAAEEEEEEEEEEEEIIIIIIIIONNNNNNRRRRRRTTTTTTLLLLSSSSUDDDGGGMMMBBCCPPFFHHVVJQKWXYZ"

func CreateGame(userID int64, name string) (*uuid.UUID, error) {
	board := initEmptyBoard()

	available := []rune(initialLetters)

	rack := drawLetters(&available, 7)

	boardJSON, err := json.Marshal(board)
	if err != nil {
		return nil, err
	}

	availableStr := string(available)
	gameID := uuid.New()

	query := `
		INSERT INTO games (id, name, created_by, current_turn, board, available_letters, created_at)
		VALUES ($1, $2, $3, $3, $4, $5, $6)
	`

	_, err = database.Query(query, gameID, name, userID, boardJSON, availableStr, time.Now())
	if err != nil {
		return nil, err
	}

	rackStr := string(rack)
	query = `
		INSERT INTO game_players (game_id, player_id, rack, position, score)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = database.Query(query, gameID, userID, rackStr, 0, 0)
	if err != nil {
		return nil, err
	}

	return &gameID, nil
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
	// 1. Vérifier que l'utilisateur est bien dans la partie et que c'est à son tour
	var currentTurn int64
	err := database.QueryRow(`SELECT current_turn FROM games WHERE id = $1`, gameID).Scan(&currentTurn)
	if err != nil {
		return fmt.Errorf("game not found")
	}
	if currentTurn != userID {
		return fmt.Errorf("not your turn")
	}

	// 2. Récupérer le rack du joueur
	var rack string
	err = database.QueryRow(`
		SELECT rack FROM game_players WHERE game_id = $1 AND player_id = $2
	`, gameID, userID).Scan(&rack)
	if err != nil {
		return fmt.Errorf("player not in game")
	}

	// 3. Vérifier que les lettres posées sont bien disponibles dans le rack
	if !rackContains(rack, req.Letters) {
		return fmt.Errorf("invalid move: you don't have the required letters")
	}

	// 3.2 Vérifier que les lettres sont alignées (même ligne ou même colonne)
	if len(req.Letters) == 0 {
		return fmt.Errorf("no letters provided")
	} else if len(req.Letters) > 7 {
		return fmt.Errorf("cannot place more than 7 letters in one move")
	}

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

	// 4. Charger le plateau
	var boardRaw []byte
	err = database.QueryRow(`SELECT board FROM games WHERE id = $1`, gameID).Scan(&boardRaw)
	if err != nil {
		return fmt.Errorf("failed to load board")
	}

	var board [15][15]string
	if err := json.Unmarshal(boardRaw, &board); err != nil {
		return fmt.Errorf("failed to decode board")
	}

	// 4.2 Vérifier placement valide (centre ou contact avec mot existant)
	isFirstMove := true
	for y := 0; y < 15 && isFirstMove; y++ {
		for x := 0; x < 15 && isFirstMove; x++ {
			if board[y][x] != "" {
				isFirstMove = false
			}
		}
	}

	if isFirstMove {
		// Doit contenir la case centrale
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
		// Doit toucher une lettre déjà posée
		touchesExisting := false
		for _, l := range req.Letters {
			adjacent := [][2]int{
				{l.X - 1, l.Y},
				{l.X + 1, l.Y},
				{l.X, l.Y - 1},
				{l.X, l.Y + 1},
			}
			for _, letter := range adjacent {
				x := letter[0]
				y := letter[1]
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

	// 5. Appliquer les lettres sur le plateau
	for _, l := range req.Letters {
		if board[l.Y][l.X] != "" {
			return fmt.Errorf("cell already occupied")
		}
		board[l.Y][l.X] = l.Char
	}

	// 6. Recalculer le rack (retirer les lettres posées, tirer de nouvelles lettres)
	newRack, err := updatePlayerRack(gameID, userID, rack, req.Letters)
	if err != nil {
		return fmt.Errorf("failed to update rack: %v", err)
	}

	// 7. Encoder le nouveau plateau
	newBoardJSON, _ := json.Marshal(board)

	// 8. Calcul du score → placeholder (1 point par lettre)
	moveScore := len(req.Letters)

	// 9. Enregistrer le coup
	moveJSON, _ := json.Marshal(req)
	_, err = database.Pool.Exec(context.Background(), `
		INSERT INTO game_moves (game_id, player_id, move)
		VALUES ($1, $2, $3)
	`, gameID, userID, moveJSON)
	if err != nil {
		return fmt.Errorf("failed to insert move")
	}

	// 10. Update du plateau, du sac, du rack, du score, du tour
	tx, err := database.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin tx")
	}
	defer tx.Rollback(context.Background())

	// Plateau
	_, err = tx.Exec(context.Background(), `UPDATE games SET board = $1 WHERE id = $2`, newBoardJSON, gameID)
	if err != nil {
		return err
	}

	// Rack + score
	_, err = tx.Exec(context.Background(), `
		UPDATE game_players SET rack = $1, score = score + $2
		WHERE game_id = $3 AND player_id = $4
	`, newRack, moveScore, gameID, userID)
	if err != nil {
		return err
	}

	// Tour suivant
	var nextPlayerID int64
	err = database.QueryRow(`
		SELECT player_id FROM game_players
		WHERE game_id = $1
		ORDER BY position
	`, gameID).Scan(&nextPlayerID) // Pour l'instant on force retour à joueur 0
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), `
		UPDATE games SET current_turn = $1 WHERE id = $2
	`, nextPlayerID, gameID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func GetGamesByUserID(userID int64) ([]response.GameSummary, error) {
	query := `
		SELECT g.id, g.name, g.current_turn, u.username
		FROM games g
		JOIN users u ON u.id = g.current_turn
		JOIN game_players gp ON gp.game_id = g.id
		WHERE gp.player_id = $1
		AND g.status = 'ongoing'
		ORDER BY g.created_at DESC
	`

	rows, err := database.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []response.GameSummary

	for rows.Next() {
		var g response.GameSummary
		err := rows.Scan(&g.ID, &g.Name, &g.CurrentTurnUserID, &g.CurrentTurnUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game row: %w", err)
		}
		games = append(games, g)
	}

	return games, nil
}
