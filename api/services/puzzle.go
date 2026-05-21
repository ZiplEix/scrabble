package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/midgame"
	dbmodels "github.com/ZiplEix/scrabble/api/models/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	resp "github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/word"
	"github.com/google/uuid"
)

const (
	PuzzleLevelInfinite = 0
	PuzzleLevelEasy     = 1
	PuzzleLevelMedium   = 2
	PuzzleLevelHard     = 3

	PuzzleTimeoutEasy   = 180 // 3 minutes
	PuzzleTimeoutMedium = 300 // 5 minutes
	PuzzleTimeoutHard   = 420 // 7 minutes
)

// GetTimeoutForLevel retourne le timeout en secondes pour un niveau
func GetTimeoutForLevel(level int) int {
	switch level {
	case PuzzleLevelInfinite:
		return 0
	case PuzzleLevelMedium:
		return PuzzleTimeoutMedium
	case PuzzleLevelHard:
		return PuzzleTimeoutHard
	default:
		return PuzzleTimeoutEasy
	}
}

// GetTargetWordsForLevel retourne la densité cible du plateau
// pour simuler une partie déjà bien avancée.
func GetTargetWordsForLevel(level int) int {
	switch level {
	case PuzzleLevelInfinite:
		return 16
	case PuzzleLevelMedium:
		return 16
	case PuzzleLevelHard:
		return 20
	default:
		return 12
	}
}

// GetDailyPuzzleLevel choisit une difficulté automatiquement selon la date.
// Cycle déterministe: facile -> moyen -> difficile.
func GetDailyPuzzleLevel(day time.Time) int {
	levels := []int{PuzzleLevelEasy, PuzzleLevelMedium, PuzzleLevelHard}
	idx := day.UTC().YearDay() % len(levels)
	return levels[idx]
}

// GenerateDailyPuzzle génère ou retourne le puzzle du jour
// Crée un nouveau puzzle si aucun n'existe pour la date actuelle
// Le seed est utilisé pour garantir la reproductibilité
func GenerateDailyPuzzle(ctx context.Context, level int) (*dbmodels.DailyPuzzle, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Vérifier si un puzzle existe déjà pour aujourd'hui
	existing := &dbmodels.DailyPuzzle{}
	err := database.QueryRow(`
		SELECT id, puzzle_date, level, board, available_letters, seed, created_at, updated_at
		FROM daily_puzzles
		WHERE puzzle_date = $1
	`, today).Scan(
		&existing.ID,
		&existing.PuzzleDate,
		&existing.Level,
		&existing.Board,
		&existing.AvailableLetters,
		&existing.Seed,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)

	if err == nil {
		// Cas test: permettre de passer le puzzle du jour en mode infini même s'il existe déjà.
		if level == PuzzleLevelInfinite && existing.Level != PuzzleLevelInfinite {
			err = database.QueryRow(`
				UPDATE daily_puzzles
				SET level = $1, updated_at = $2
				WHERE id = $3
				RETURNING id, puzzle_date, level, board, available_letters, seed, created_at, updated_at
			`, level, time.Now().UTC(), existing.ID).Scan(
				&existing.ID,
				&existing.PuzzleDate,
				&existing.Level,
				&existing.Board,
				&existing.AvailableLetters,
				&existing.Seed,
				&existing.CreatedAt,
				&existing.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
		}

		// Puzzle existe déjà
		return existing, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	if level != PuzzleLevelInfinite && (level < PuzzleLevelEasy || level > PuzzleLevelHard) {
		level = GetDailyPuzzleLevel(today)
	}

	// Créer un nouveau puzzle midgame déterministe pour la journée.
	seedInt := today.Unix() + int64(level*1000)
	seed := fmt.Sprintf("puzzle_%d_%d", seedInt, level)

	gen := midgame.NewGenerator(GetTargetWordsForLevel(level), seedInt)
	genResult, err := gen.Generate()
	if err != nil {
		return nil, err
	}

	boardJSON, err := json.Marshal(genResult.Board)
	if err != nil {
		return nil, err
	}

	puzzle := &dbmodels.DailyPuzzle{
		ID:               uuid.New().String(),
		PuzzleDate:       today,
		Level:            level,
		Board:            boardJSON,
		AvailableLetters: genResult.PlayerRack,
		Seed:             seed,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// Sauvegarder en DB
	err = database.QueryRow(`
		INSERT INTO daily_puzzles (id, puzzle_date, level, board, available_letters, seed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, puzzle_date, level, board, available_letters, seed, created_at, updated_at
	`,
		puzzle.ID,
		puzzle.PuzzleDate,
		puzzle.Level,
		puzzle.Board,
		puzzle.AvailableLetters,
		puzzle.Seed,
		puzzle.CreatedAt,
		puzzle.UpdatedAt,
	).Scan(
		&puzzle.ID,
		&puzzle.PuzzleDate,
		&puzzle.Level,
		&puzzle.Board,
		&puzzle.AvailableLetters,
		&puzzle.Seed,
		&puzzle.CreatedAt,
		&puzzle.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return puzzle, nil
}

// GetCurrentPuzzle retourne le puzzle du jour
// Retourne un puzzle par défaut de niveau EASY si aucun n'existe
func GetCurrentPuzzle(ctx context.Context) (*resp.PuzzleInfo, error) {
	puzzle, err := GenerateDailyPuzzle(ctx, GetDailyPuzzleLevel(time.Now().UTC()))
	if err != nil {
		return nil, err
	}

	var board interface{}
	err = json.Unmarshal(puzzle.Board, &board)
	if err != nil {
		return nil, err
	}

	return &resp.PuzzleInfo{
		ID:               puzzle.ID,
		PuzzleDate:       puzzle.PuzzleDate.Format("2006-01-02"),
		Level:            puzzle.Level,
		Board:            board,
		AvailableLetters: puzzle.AvailableLetters,
		TimeoutSeconds:   GetTimeoutForLevel(puzzle.Level),
		CreatedAt:        puzzle.CreatedAt,
	}, nil
}

// GetPuzzleByID retourne un puzzle spécifique
func GetPuzzleByID(ctx context.Context, puzzleID string) (*resp.PuzzleInfo, error) {
	puzzle := &dbmodels.DailyPuzzle{}
	err := database.QueryRow(`
		SELECT id, puzzle_date, level, board, available_letters, seed, created_at, updated_at
		FROM daily_puzzles
		WHERE id = $1
	`, puzzleID).Scan(
		&puzzle.ID,
		&puzzle.PuzzleDate,
		&puzzle.Level,
		&puzzle.Board,
		&puzzle.AvailableLetters,
		&puzzle.Seed,
		&puzzle.CreatedAt,
		&puzzle.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	var board interface{}
	err = json.Unmarshal(puzzle.Board, &board)
	if err != nil {
		return nil, err
	}

	return &resp.PuzzleInfo{
		ID:               puzzle.ID,
		PuzzleDate:       puzzle.PuzzleDate.Format("2006-01-02"),
		Level:            puzzle.Level,
		Board:            board,
		AvailableLetters: puzzle.AvailableLetters,
		TimeoutSeconds:   GetTimeoutForLevel(puzzle.Level),
		CreatedAt:        puzzle.CreatedAt,
	}, nil
}

// GetPuzzleHistory retourne l'historique des puzzles avec les tentatives du joueur
func GetPuzzleHistory(ctx context.Context, playerID int64, limit int, offset int) ([]*resp.PuzzleHistory, error) {
	rows, err := database.Query(`
		SELECT 
			dp.id, 
			dp.puzzle_date, 
			dp.level,
			COALESCE(pa.submitted_at IS NOT NULL AND pa.score IS NOT NULL, false) as has_attempted,
			pa.id,
			pa.score,
			pa.time_used,
			pa.words_played,
			pa.submitted_at,
			pa.started_at
		FROM daily_puzzles dp
		LEFT JOIN puzzle_attempts pa ON dp.id = pa.puzzle_id AND pa.player_id = $1
		ORDER BY dp.puzzle_date DESC
		LIMIT $2 OFFSET $3
	`, playerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*resp.PuzzleHistory

	for rows.Next() {
		var (
			puzzleID        string
			puzzleDate      time.Time
			level           int
			hasAttempted    bool
			attemptID       sql.NullString
			score           sql.NullInt64
			timeUsed        sql.NullInt64
			wordsPlayedJSON sql.NullString
			submittedAt     sql.NullTime
			startedAt       sql.NullTime
		)

		err := rows.Scan(
			&puzzleID,
			&puzzleDate,
			&level,
			&hasAttempted,
			&attemptID,
			&score,
			&timeUsed,
			&wordsPlayedJSON,
			&submittedAt,
			&startedAt,
		)
		if err != nil {
			return nil, err
		}

		h := &resp.PuzzleHistory{
			ID:           puzzleID,
			PuzzleDate:   puzzleDate.Format("2006-01-02"),
			Level:        level,
			HasAttempted: hasAttempted,
		}

		if hasAttempted && attemptID.Valid {
			submittedAtPtr := (*time.Time)(nil)
			if submittedAt.Valid {
				submittedAtPtr = &submittedAt.Time
			}
			createdAt := startedAt.Time
			h.PlayerAttempt = &resp.PuzzleAttempt{
				ID:           attemptID.String,
				PuzzleID:     puzzleID,
				PlayerID:     playerID,
				StartedAt:    startedAt.Time,
				Score:        int(score.Int64),
				TimeUsedSecs: int(timeUsed.Int64),
				SubmittedAt:  submittedAtPtr,
				CreatedAt:    createdAt,
			}

			if wordsPlayedJSON.Valid {
				var words []resp.PuzzleWordRecord
				err := json.Unmarshal([]byte(wordsPlayedJSON.String), &words)
				if err == nil {
					h.PlayerAttempt.WordsPlayed = words
				}
			}

			// Calculer le rang du jour
			rank, err := getPuzzleAttemptRank(ctx, puzzleID, int(score.Int64))
			if err == nil {
				h.PlayerAttempt.RankToday = rank
			}
		}

		history = append(history, h)
	}

	return history, rows.Err()
}

// SubmitPuzzleAttempt soumet une tentative de puzzle.
// Retrouve la session existante (started_at), calcule le time_used côté serveur,
// valide que le timeout n'est pas dépassé, puis enregistre le résultat.
func SubmitPuzzleAttempt(ctx context.Context, playerID int64, req *request.SubmitPuzzleAttemptRequest) (*resp.PuzzleAttempt, error) {
	// Récupérer la session de démarrage (StartPuzzle doit avoir été appelé avant)
	var attemptID string
	var startedAt time.Time
	var alreadySubmitted bool
	err := database.QueryRow(`
		SELECT id, started_at, submitted_at IS NOT NULL AND score IS NOT NULL
		FROM puzzle_attempts
		WHERE puzzle_id = $1 AND player_id = $2
	`, req.PuzzleID, playerID).Scan(&attemptID, &startedAt, &alreadySubmitted)

	if err == sql.ErrNoRows {
		return nil, errors.New("vous devez d'abord ouvrir le puzzle avant de soumettre")
	}
	if err != nil {
		return nil, err
	}
	if alreadySubmitted {
		return nil, errors.New("vous avez déjà soumis une réponse pour ce puzzle")
	}

	// Récupérer le puzzle pour calculer le timeout et valider/scorer le coup.
	var (
		boardRaw         []byte
		availableLetters string
		level            int
	)
	err = database.QueryRow(`
		SELECT board, available_letters, level
		FROM daily_puzzles
		WHERE id = $1
	`, req.PuzzleID).Scan(&boardRaw, &availableLetters, &level)
	if err != nil {
		return nil, err
	}

	// Calculer le temps écoulé côté serveur — on ne fait jamais confiance au client
	now := time.Now().UTC()
	timeUsed := int(now.Sub(startedAt).Seconds())
	expectedTimeout := GetTimeoutForLevel(level)

	if expectedTimeout > 0 && timeUsed > expectedTimeout+10 { // +10s de buffer réseau
		return nil, errors.New("le temps imparti a été dépassé")
	}

	// Valider et scorer les mouvements.
	var (
		score       int
		wordsPlayed []resp.PuzzleWordRecord
	)

	if len(req.Letters) > 0 {
		var board [15][15]string
		if err := json.Unmarshal(boardRaw, &board); err != nil {
			return nil, fmt.Errorf("failed to unmarshal puzzle board: %w", err)
		}

		if rl, err := resolveBlanks(availableLetters, req.Letters); err == nil {
			req.Letters = rl
		}

		if err := applyLetters(&board, req.Letters); err != nil {
			return nil, err
		}

		formedWords := extractFormedWords(board, req.Letters)
		for _, fw := range formedWords {
			if !word.WordExists(fw.Word) {
				return nil, fmt.Errorf("mot invalide: %s", fw.Word)
			}
		}

		isNew := make(map[Pos]bool, len(req.Letters))
		isBlank := make(map[Pos]bool, len(req.Letters))
		boardBlank := map[Pos]bool{}
		for _, pl := range req.Letters {
			pos := Pos{pl.X, pl.Y}
			isNew[pos] = true
			if pl.Blank {
				isBlank[pos] = true
				boardBlank[pos] = true
			}
		}

		score = computeMoveScore(board, req.Letters, boardBlank)
		wordsPlayed = make([]resp.PuzzleWordRecord, 0, len(formedWords))
		for _, fw := range formedWords {
			direction := "horizontal"
			if fw.DY != 0 {
				direction = "vertical"
			}
			wordsPlayed = append(wordsPlayed, resp.PuzzleWordRecord{
				Word:      fw.Word,
				Position:  fmt.Sprintf("%d,%d", fw.StartX, fw.StartY),
				Direction: direction,
				Score:     computeWordScore(board, fw, isNew, isBlank),
			})
		}
	} else {
		// Fallback de compatibilité avec les anciens clients qui envoient uniquement words_played.
		score, wordsPlayed, err = validateAndScorePuzzleAttempt(req.WordsPlayed)
		if err != nil {
			return nil, err
		}
	}

	wordsJSON, err := json.Marshal(wordsPlayed)
	if err != nil {
		return nil, err
	}

	// Mettre à jour la ligne existante avec le résultat
	var submittedAt time.Time
	err = database.QueryRow(`
		UPDATE puzzle_attempts
		SET score = $1, words_played = $2, time_used = $3, submitted_at = $4
		WHERE id = $5
		RETURNING submitted_at
	`, score, wordsJSON, timeUsed, now, attemptID).Scan(&submittedAt)
	if err != nil {
		return nil, err
	}

	// Calculer le rang
	rank, err := getPuzzleAttemptRank(ctx, req.PuzzleID, score)
	if err != nil {
		rank = 1
	}

	result := &resp.PuzzleAttempt{
		ID:           attemptID,
		PuzzleID:     req.PuzzleID,
		PlayerID:     playerID,
		StartedAt:    startedAt,
		Score:        score,
		TimeUsedSecs: timeUsed,
		RankToday:    rank,
		SubmittedAt:  &submittedAt,
		CreatedAt:    startedAt,
	}
	result.WordsPlayed = wordsPlayed

	return result, nil
}

// StartPuzzle enregistre le timestamp de début d'un joueur pour un puzzle.
// Si la session existe déjà (joueur reprend la session), retourne la session existante.
// Si déjà soumis, retourne une erreur.
func StartPuzzle(ctx context.Context, playerID int64, puzzleID string) (*resp.PuzzleStarted, error) {
	// Récupérer le niveau du puzzle
	var level int
	err := database.QueryRow(`SELECT level FROM daily_puzzles WHERE id = $1`, puzzleID).Scan(&level)
	if err != nil {
		return nil, err
	}

	// Mode test (niveau infini): on réinitialise systématiquement la tentative
	// pour permettre de rejouer le puzzle autant de fois que nécessaire.
	if level == PuzzleLevelInfinite {
		now := time.Now().UTC()
		attemptID := uuid.New().String()
		var startedAt time.Time

		err = database.QueryRow(`
			INSERT INTO puzzle_attempts (id, puzzle_id, player_id, started_at, score, words_played, time_used, submitted_at, created_at)
			VALUES ($1, $2, $3, $4, NULL, NULL, 0, NULL, $5)
			ON CONFLICT (puzzle_id, player_id)
			DO UPDATE SET
				id = EXCLUDED.id,
				started_at = EXCLUDED.started_at,
				created_at = EXCLUDED.created_at,
				score = NULL,
				words_played = NULL,
				time_used = 0,
				submitted_at = NULL
			RETURNING id, started_at
		`, attemptID, puzzleID, playerID, now, now).Scan(&attemptID, &startedAt)
		if err != nil {
			return nil, err
		}

		return &resp.PuzzleStarted{
			AttemptID:      attemptID,
			StartedAt:      startedAt,
			TimeoutSeconds: GetTimeoutForLevel(level),
			AlreadyStarted: false,
		}, nil
	}

	// Vérifier si une session existe déjà
	var attemptID string
	var startedAt time.Time
	var alreadySubmitted bool
	err = database.QueryRow(`
		SELECT id, started_at, submitted_at IS NOT NULL AND score IS NOT NULL
		FROM puzzle_attempts
		WHERE puzzle_id = $1 AND player_id = $2
	`, puzzleID, playerID).Scan(&attemptID, &startedAt, &alreadySubmitted)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == nil {
		if alreadySubmitted {
			return nil, errors.New("vous avez déjà soumis une réponse pour ce puzzle")
		}

		// Session déjà existante — retourner la session existante
		return &resp.PuzzleStarted{
			AttemptID:      attemptID,
			StartedAt:      startedAt,
			TimeoutSeconds: GetTimeoutForLevel(level),
			AlreadyStarted: true,
		}, nil
	}

	// Créer une nouvelle session
	now := time.Now().UTC()
	newAttemptID := uuid.New().String()

	err = database.QueryRow(`
		INSERT INTO puzzle_attempts (id, puzzle_id, player_id, started_at, score, words_played, time_used, submitted_at, created_at)
		VALUES ($1, $2, $3, $4, NULL, NULL, 0, NULL, $5)
		RETURNING id, started_at
	`, newAttemptID, puzzleID, playerID, now, now).Scan(&newAttemptID, &startedAt)
	if err != nil {
		return nil, err
	}

	return &resp.PuzzleStarted{
		AttemptID:      newAttemptID,
		StartedAt:      startedAt,
		TimeoutSeconds: GetTimeoutForLevel(level),
		AlreadyStarted: false,
	}, nil
}

// SimulatePuzzleScore simule le score d'un coup sans soumettre la tentative.
func SimulatePuzzleScore(ctx context.Context, playerID int64, puzzleID string, letters []request.PlacedLetter) (int, error) {
	if len(letters) == 0 {
		return 0, nil
	}

	var (
		boardRaw         []byte
		availableLetters string
		level            int
	)
	err := database.QueryRow(`
		SELECT board, available_letters, level
		FROM daily_puzzles
		WHERE id = $1
	`, puzzleID).Scan(&boardRaw, &availableLetters, &level)
	if err != nil {
		return 0, err
	}

	var (
		startedAt   time.Time
		score       sql.NullInt64
		submittedAt sql.NullTime
	)
	err = database.QueryRow(`
		SELECT started_at, score, submitted_at
		FROM puzzle_attempts
		WHERE puzzle_id = $1 AND player_id = $2
	`, puzzleID, playerID).Scan(&startedAt, &score, &submittedAt)
	if err == sql.ErrNoRows {
		return 0, errors.New("vous devez d'abord démarrer le puzzle")
	}
	if err != nil {
		return 0, err
	}
	if submittedAt.Valid && score.Valid {
		return 0, errors.New("vous avez déjà soumis ce puzzle")
	}

	expectedTimeout := GetTimeoutForLevel(level)
	if expectedTimeout > 0 && int(time.Since(startedAt).Seconds()) > expectedTimeout+10 {
		return 0, errors.New("le temps imparti a été dépassé")
	}

	var board [15][15]string
	if err := json.Unmarshal(boardRaw, &board); err != nil {
		return 0, fmt.Errorf("failed to unmarshal puzzle board: %w", err)
	}

	if rl, err := resolveBlanks(availableLetters, letters); err == nil {
		letters = rl
	}

	if err := applyLetters(&board, letters); err != nil {
		return 0, err
	}

	boardBlank := map[Pos]bool{}
	for _, pl := range letters {
		if pl.Blank {
			boardBlank[Pos{pl.X, pl.Y}] = true
		}
	}

	return computeMoveScore(board, letters, boardBlank), nil
}

// GetPuzzleLeaderboard retourne le classement du jour pour un puzzle
func GetPuzzleLeaderboard(ctx context.Context, puzzleID string, limit int, offset int) ([]*resp.PuzzleDailyLeaderboard, error) {
	rows, err := database.Query(`
		SELECT 
			ROW_NUMBER() OVER (ORDER BY pa.score DESC) as rank,
			pa.player_id,
			u.username,
			pa.score,
			pa.time_used,
			pa.submitted_at
		FROM puzzle_attempts pa
		JOIN users u ON pa.player_id = u.id
		WHERE pa.puzzle_id = $1 AND pa.submitted_at IS NOT NULL AND pa.score IS NOT NULL
		ORDER BY pa.score DESC
		LIMIT $2 OFFSET $3
	`, puzzleID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []*resp.PuzzleDailyLeaderboard

	for rows.Next() {
		var rank int
		var score sql.NullInt64
		var timeUsed sql.NullInt64
		var submittedAt sql.NullTime
		entry := &resp.PuzzleDailyLeaderboard{}

		err := rows.Scan(
			&rank,
			&entry.PlayerID,
			&entry.Username,
			&score,
			&timeUsed,
			&submittedAt,
		)
		if err != nil {
			return nil, err
		}

		entry.Rank = rank
		entry.Score = int(score.Int64)
		entry.TimeUsed = int(timeUsed.Int64)
		entry.SubmittedAt = submittedAt.Time
		entry.Attempts = 1 // Une tentative par joueur max
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, rows.Err()
}

// HasPlayerAttemptedPuzzle vérifie si un joueur a déjà soumis une réponse pour un puzzle
func HasPlayerAttemptedPuzzle(ctx context.Context, playerID int64, puzzleID string) (bool, error) {
	var level int
	err := database.QueryRow(`SELECT level FROM daily_puzzles WHERE id = $1`, puzzleID).Scan(&level)
	if err != nil {
		return false, err
	}

	// En mode test infini, on ne bloque jamais l'interface avec un état "déjà tenté".
	if level == PuzzleLevelInfinite {
		return false, nil
	}

	var count int
	err = database.QueryRow(`
		SELECT COUNT(*) FROM puzzle_attempts 
		WHERE puzzle_id = $1 AND player_id = $2 AND submitted_at IS NOT NULL AND score IS NOT NULL
	`, puzzleID, playerID).Scan(&count)

	return count > 0, err
}

// GetPlayerPuzzleStats retourne les stats du joueur pour tous les puzzles
func GetPlayerPuzzleStats(ctx context.Context, playerID int64) (*resp.PuzzleStats, error) {
	stats := &resp.PuzzleStats{}

	err := database.QueryRow(`
		SELECT 
			COUNT(*) as total_attempts,
			COALESCE(MAX(score), 0) as best_score,
			COALESCE(AVG(score)::int, 0) as average_score
		FROM puzzle_attempts
		WHERE player_id = $1 AND submitted_at IS NOT NULL AND score IS NOT NULL
	`, playerID).Scan(
		&stats.TotalAttempts,
		&stats.BestScore,
		&stats.AverageScore,
	)

	if err != nil {
		return nil, err
	}

	// Nombre de puzzles complétés
	err = database.QueryRow(`
		SELECT COUNT(DISTINCT puzzle_id) FROM puzzle_attempts
		WHERE player_id = $1 AND submitted_at IS NOT NULL AND score IS NOT NULL
	`, playerID).Scan(&stats.CompletedPuzzles)

	return stats, err
}

// ============= Private helpers =============

// validateAndScorePuzzleAttempt valide les mots et calcule le score total
func validateAndScorePuzzleAttempt(wordsForSubmit []request.PuzzleWordForSubmit) (int, []resp.PuzzleWordRecord, error) {
	totalScore := 0
	wordsPlayed := []resp.PuzzleWordRecord{}

	for _, w := range wordsForSubmit {
		// Valider que le mot existe dans le dictionnaire
		if !word.WordExists(w.Word) {
			return 0, nil, fmt.Errorf("mot invalide: %s", w.Word)
		}

		// Calculer le score du mot (pour l'instant, utiliser longueur * 1)
		// TODO: implémenter scoring correct avec valeurs des lettres
		score := len(w.Word)

		totalScore += score
		wordsPlayed = append(wordsPlayed, resp.PuzzleWordRecord{
			Word:      w.Word,
			Position:  w.Position,
			Direction: w.Direction,
			Score:     score,
		})
	}

	return totalScore, wordsPlayed, nil
}

// getPuzzleAttemptRank retourne le rang du joueur pour ce puzzle (parmi les soumis)
func getPuzzleAttemptRank(ctx context.Context, puzzleID string, score int) (int, error) {
	var rank int
	err := database.QueryRow(`
		SELECT COUNT(*) + 1 FROM puzzle_attempts
		WHERE puzzle_id = $1 AND score > $2 AND submitted_at IS NOT NULL
	`, puzzleID, score).Scan(&rank)

	return rank, err
}
