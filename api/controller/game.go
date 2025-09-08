package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func CreateGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour créer une partie",
		})
	}

	var req request.CreateGameRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	if req.Name == "" {
		logctx.Add(c, "reason", "game_name_required")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "game name is required",
			"message": "Le nom de la partie est requis",
		})
	}

	var usernames []string
	for _, player := range req.Players {
		usernames = append(usernames, strings.ToLower(strings.TrimSpace(player)))
	}

	if len(usernames) > 3 {
		logctx.Add(c, "reason", "too_many_players")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "too many players",
			"message": "Vous ne pouvez pas inviter plus de 3 joueurs",
		})
	}

	gameID, err := services.CreateGame(userID, req.Name, usernames, req.RevangeFrom)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_create_game",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create game: %v", err),
			"message": "Erreur lors de la création de la partie, veuillez réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Game created",
		"game_id": gameID,
	})
}

func DeleteGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour supprimer une partie",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour la suppression",
		})
	}
	logctx.Add(c, "game_id", gameID)

	if err := services.DeleteGame(userID, gameID); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_delete_game",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to delete game: %v", err),
			"message": "Erreur lors de la suppression de la partie, veuillez réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusNoContent, map[string]string{
		"message": "Game deleted successfully",
	})
}

func RenameGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour renommer une partie",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour le renommage",
		})
	}
	logctx.Add(c, "game_id", gameID)

	var req request.RenameGameRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	if err := services.RenameGame(userID, gameID, req.NewName); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_rename_game",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to rename game: %v", err),
			"message": "Erreur lors du renommage de la partie, veuillez réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Game renamed successfully",
	})
}

func GetGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour accéder aux détails de la partie",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour accéder aux détails",
		})
	}
	logctx.Add(c, "game_id", gameID)

	game, err := services.GetGameDetails(userID, gameID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logctx.Add(c, "reason", "game_not_found")
			return c.JSON(http.StatusNotFound, echo.Map{
				"error":   fmt.Sprintf("failed to load game: %v", err),
				"message": "Erreur lors du chargement de la partie, la partie n'existe pas. Si vous pensez qu'il s'agit d'une erreur, veuillez recharger la page. Si le problème persiste, contactez le support.",
			})
		}
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_load_game",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusForbidden, echo.Map{
			"error":   fmt.Sprintf("failed to load game: %v", err),
			"message": "Erreur lors du chargement de la partie. Veuillez vérifier vos permissions ou recharger la page. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusOK, game)
}

func PlayMove(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour jouer un coup",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour jouer un coup",
		})
	}
	logctx.Add(c, "game_id", gameID)

	var req request.PlayMoveRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	if err := services.PlayMove(gameID, userID, req); err != nil {
		if strings.Contains(err.Error(), "not your turn") {
			logctx.Add(c, "reason", "not_your_turn")
			return c.JSON(http.StatusForbidden, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Ce n'est pas votre tour de jouer. Veuillez attendre votre tour.",
			})
		} else if strings.Contains(err.Error(), "invalid move") {
			logctx.Add(c, "reason", "invalid_move")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Coup invalide. Veuillez vérifier votre coup et réessayer.",
			})
		} else if strings.Contains(err.Error(), "more than 7 letters") {
			logctx.Add(c, "reason", "too_many_letters")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Vous ne pouvez pas jouer plus de 7 lettres à la fois. Veuillez réduire le nombre de lettres et réessayer.",
			})
		} else if strings.Contains(err.Error(), "must be aligned") {
			logctx.Add(c, "reason", "letters_not_aligned")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Les lettres doivent être alignées. Veuillez vérifier la position de vos lettres et réessayer.",
			})
		} else if strings.Contains(err.Error(), "first move must cover the center cell") {
			logctx.Add(c, "reason", "first_move_not_centered")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Le premier coup doit couvrir la case centrale. Veuillez ajuster votre coup et réessayer.",
			})
		} else if strings.Contains(err.Error(), "word must connect to existing letters") {
			logctx.Add(c, "reason", "word_not_connected")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Le mot doit se connecter à des lettres existantes. Veuillez vérifier votre coup et réessayer.",
			})
		} else if strings.Contains(err.Error(), "invalid word played:") {
			logctx.Add(c, "reason", "invalid_word")
			word := strings.TrimSpace(strings.Split(err.Error(), ":")[1])
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": fmt.Sprintf("Le mot '%s' n'est pas valide. Veuillez vérifier l'orthographe et réessayer.", word),
			})
		}

		logctx.Merge(c, map[string]any{
			"reason": "failed_to_play_move",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to play move: %v", err),
			"message": "Erreur lors de la tentative de jouer le coup. Veuillez recharger la page ou réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "move played successfully",
	})
}

func GetNewRack(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour obtenir une nouvelle main",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour obtenir une nouvelle main",
		})
	}
	logctx.Add(c, "game_id", gameID)

	newRack, err := services.GetNewRack(userID, gameID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_new_rack",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to get new rack: %v", err),
			"message": "Erreur lors de la récupération de la nouvelle main. Veuillez recharger la page ou réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusOK, newRack)
}

func GetUserGames(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour accéder à vos parties",
		})
	}

	games, err := services.GetGamesByUserID(userID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_games",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to get games: %v", err),
			"message": "Erreur lors de la récupération de vos parties. Veuillez recharger la page ou réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.JSON(http.StatusOK, response.GamesListResponse{
		Games: games,
	})
}

func SimulateScore(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}
	logctx.Add(c, "game_id", gameID)

	var body struct {
		Letters []request.PlacedLetter `json:"letters"`
	}
	if err := c.Bind(&body); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	score, err := services.SimulateScore(gameID, userID, body.Letters)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_simulate_score",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]int{"score": score})
}

func PassTurn(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour passer votre tour",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour passer votre tour",
		})
	}
	logctx.Add(c, "game_id", gameID)

	err := services.PassTurn(userID, gameID)
	if err != nil {
		if strings.Contains(err.Error(), "not your turn") {
			logctx.Add(c, "reason", "not_your_turn")
			return c.JSON(http.StatusForbidden, echo.Map{
				"error":   fmt.Sprintf("failed to pass turn: %v", err),
				"message": "Ce n'est pas votre tour de jouer. Veuillez attendre votre tour.",
			})
		} else if strings.Contains(err.Error(), "game not found") {
			logctx.Add(c, "reason", "game_not_found")
			return c.JSON(http.StatusNotFound, echo.Map{
				"error":   fmt.Sprintf("failed to pass turn: %v", err),
				"message": "La partie n'existe pas ou a été supprimée. Veuillez recharger la page ou réessayer. Si le problème persiste, contactez le support.",
			})
		}
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_pass_turn",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("failed to pass turn: %v", err),
			"message": "Erreur lors de la tentative de passer votre tour. Veuillez recharger la page ou réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.NoContent(http.StatusOK)
}
