package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

// GetCurrentPuzzle retourne le puzzle du jour
func GetCurrentPuzzle(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	ctx := c.Request().Context()

	puzzle, err := services.GetCurrentPuzzle(ctx)
	if err != nil {
		logctx.Add(c, "reason", "get_current_puzzle_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "failed to get puzzle",
			"message": "Impossible de récupérer le puzzle du jour",
		})
	}

	// Vérifier si le joueur a déjà tenté ce puzzle
	hasAttempted, err := services.HasPlayerAttemptedPuzzle(ctx, userID, puzzle.ID)
	if err != nil {
		logctx.Add(c, "reason", "check_attempt_failed")
		logctx.Add(c, "error", err.Error())
	}
	puzzle.HasPlayerAttempted = hasAttempted

	return c.JSON(http.StatusOK, puzzle)
}

// GetPuzzleByID retourne un puzzle spécifique
func GetPuzzleByID(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	puzzleID := c.Param("id")
	if puzzleID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing puzzle id",
			"message": "ID du puzzle requis",
		})
	}

	ctx := c.Request().Context()

	puzzle, err := services.GetPuzzleByID(ctx, puzzleID)
	if err != nil {
		logctx.Add(c, "reason", "get_puzzle_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "failed to get puzzle",
			"message": "Impossible de récupérer le puzzle",
		})
	}

	// Vérifier si le joueur a déjà tenté ce puzzle
	hasAttempted, err := services.HasPlayerAttemptedPuzzle(ctx, userID, puzzle.ID)
	if err != nil {
		logctx.Add(c, "reason", "check_attempt_failed")
		logctx.Add(c, "error", err.Error())
	}
	puzzle.HasPlayerAttempted = hasAttempted

	return c.JSON(http.StatusOK, puzzle)
}

// GetPuzzleHistory retourne l'historique des puzzles du joueur
func GetPuzzleHistory(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := c.Request().Context()

	history, err := services.GetPuzzleHistory(ctx, userID, limit, offset)
	if err != nil {
		logctx.Add(c, "reason", "get_history_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "failed to get history",
			"message": "Impossible de récupérer l'historique",
		})
	}

	if history == nil {
		history = []*response.PuzzleHistory{}
	}

	return c.JSON(http.StatusOK, history)
}

// StartPuzzle enregistre le timestamp de début d'un joueur pour un puzzle
// Doit être appelé avant SubmitPuzzleAttempt
func StartPuzzle(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	puzzleID := c.Param("id")
	if puzzleID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing puzzle id",
			"message": "ID du puzzle requis",
		})
	}

	ctx := c.Request().Context()

	started, err := services.StartPuzzle(ctx, userID, puzzleID)
	if err != nil {
		logctx.Add(c, "reason", "start_puzzle_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   err.Error(),
			"message": "Impossible de démarrer le puzzle",
		})
	}

	return c.JSON(http.StatusOK, started)
}

// SimulatePuzzleScore simule le score d'un coup sur un puzzle sans le soumettre
func SimulatePuzzleScore(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	puzzleID := c.Param("id")
	if puzzleID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing puzzle id",
			"message": "ID du puzzle requis",
		})
	}

	var body struct {
		Letters []request.PlacedLetter `json:"letters"`
	}
	if err := c.Bind(&body); err != nil {
		logctx.Add(c, "reason", "bind_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "invalid request body",
			"message": "Requête invalide",
		})
	}

	score, err := services.SimulatePuzzleScore(c.Request().Context(), userID, puzzleID, body.Letters)
	if err != nil {
		logctx.Add(c, "reason", "simulate_puzzle_score_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   err.Error(),
			"message": "Impossible de simuler le score",
		})
	}

	return c.JSON(http.StatusOK, map[string]int{"score": score})
}

// SubmitPuzzleAttempt soumet une tentative de puzzle
func SubmitPuzzleAttempt(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	var req request.SubmitPuzzleAttemptRequest
	if err := c.Bind(&req); err != nil {
		logctx.Add(c, "reason", "bind_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide",
		})
	}

	if req.PuzzleID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "puzzle_id required",
			"message": "ID du puzzle requis",
		})
	}

	ctx := c.Request().Context()

	attempt, err := services.SubmitPuzzleAttempt(ctx, userID, &req)
	if err != nil {
		logctx.Add(c, "reason", "submit_attempt_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   err.Error(),
			"message": "Erreur lors de la soumission de la tentative",
		})
	}

	return c.JSON(http.StatusOK, attempt)
}

// GetPuzzleLeaderboard retourne le classement d'un puzzle
func GetPuzzleLeaderboard(c echo.Context) error {
	puzzleID := c.Param("id")
	if puzzleID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing puzzle id",
			"message": "ID du puzzle requis",
		})
	}

	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := c.Request().Context()

	leaderboard, err := services.GetPuzzleLeaderboard(ctx, puzzleID, limit, offset)
	if err != nil {
		logctx.Add(c, "reason", "get_leaderboard_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "failed to get leaderboard",
			"message": "Impossible de récupérer le classement",
		})
	}

	if leaderboard == nil {
		leaderboard = []*response.PuzzleDailyLeaderboard{}
	}

	return c.JSON(http.StatusOK, leaderboard)
}

// GetPuzzleStats retourne les stats du joueur
func GetPuzzleStats(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized",
			"message": "Vous devez être connecté",
		})
	}

	ctx := c.Request().Context()

	stats, err := services.GetPlayerPuzzleStats(ctx, userID)
	if err != nil {
		logctx.Add(c, "reason", "get_stats_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "failed to get stats",
			"message": "Impossible de récupérer les statistiques",
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// GeneratePuzzleAdmin génère un nouveau puzzle (admin only)
func GeneratePuzzleAdmin(c echo.Context) error {
	level := 1 // Default: easy
	if l := c.QueryParam("level"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed >= 0 && parsed <= 3 {
			level = parsed
		}
	}

	ctx := c.Request().Context()

	puzzle, err := services.GenerateDailyPuzzle(ctx, level)
	if err != nil {
		logctx.Add(c, "reason", "generate_failed")
		logctx.Add(c, "error", err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "failed to generate puzzle",
			"message": "Erreur lors de la génération du puzzle",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Puzzle généré avec succès",
		"id":      puzzle.ID,
		"date":    puzzle.PuzzleDate,
		"level":   puzzle.Level,
	})
}
