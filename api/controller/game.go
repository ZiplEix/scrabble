package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func CreateGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour créer une partie",
		})
	}

	var req request.CreateGameRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "game name is required",
			"message": "Le nom de la partie est requis",
		})
	}

	var usernames []string
	for _, player := range req.Players {
		usernames = append(usernames, strings.ToLower(strings.TrimSpace(player)))
	}

	gameID, err := services.CreateGame(userID, req.Name, usernames)
	if err != nil {
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
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour supprimer une partie",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour la suppression",
		})
	}

	if err := services.DeleteGame(userID, gameID); err != nil {
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
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour renommer une partie",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour le renommage",
		})
	}

	var req request.RenameGameRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	if err := services.RenameGame(userID, gameID, req.NewName); err != nil {
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
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour accéder aux détails de la partie",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour accéder aux détails",
		})
	}

	game, err := services.GetGameDetails(userID, gameID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error":   fmt.Sprintf("failed to load game: %v", err),
				"message": "Erreur lors du chargement de la partie, la partie n'existe pas. Si vous pensez qu'il s'agit d'une erreur, veuillez recharger la page. Si le problème persiste, contactez le support.",
			})
		}
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
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour jouer un coup",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour jouer un coup",
		})
	}

	var req request.PlayMoveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	if err := services.PlayMove(gameID, userID, req); err != nil {
		if strings.Contains(err.Error(), "not your turn") {
			return c.JSON(http.StatusForbidden, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Ce n'est pas votre tour de jouer. Veuillez attendre votre tour.",
			})
		} else if strings.Contains(err.Error(), "invalid move") {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Coup invalide. Veuillez vérifier votre coup et réessayer.",
			})
		} else if strings.Contains(err.Error(), "more than 7 letters") {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Vous ne pouvez pas jouer plus de 7 lettres à la fois. Veuillez réduire le nombre de lettres et réessayer.",
			})
		} else if strings.Contains(err.Error(), "must be aligned") {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Les lettres doivent être alignées. Veuillez vérifier la position de vos lettres et réessayer.",
			})
		} else if strings.Contains(err.Error(), "first move must cover the center cell") {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Le premier coup doit couvrir la case centrale. Veuillez ajuster votre coup et réessayer.",
			})
		} else if strings.Contains(err.Error(), "word must connect to existing letters") {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": "Le mot doit se connecter à des lettres existantes. Veuillez vérifier votre coup et réessayer.",
			})
		} else if strings.Contains(err.Error(), "invalid word played:") {
			word := strings.TrimSpace(strings.Split(err.Error(), ":")[1])
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   fmt.Sprintf("failed to play move: %v", err),
				"message": fmt.Sprintf("Le mot '%s' n'est pas valide. Veuillez vérifier l'orthographe et réessayer.", word),
			})
		}

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
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour obtenir une nouvelle main",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour obtenir une nouvelle main",
		})
	}

	newRack, err := services.GetNewRack(userID, gameID)
	if err != nil {
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
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour accéder à vos parties",
		})
	}

	games, err := services.GetGamesByUserID(userID)
	if err != nil {
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
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	gameID := c.Param("id")

	var body struct {
		Letters []request.PlacedLetter `json:"letters"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	score, err := services.SimulateScore(gameID, userID, body.Letters)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]int{"score": score})
}

func PassTurn(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		// return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour passer votre tour",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		// return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour passer votre tour",
		})
	}

	err := services.PassTurn(userID, gameID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("failed to pass turn: %v", err),
			"message": "Erreur lors de la tentative de passer votre tour. Veuillez recharger la page ou réessayer. Si le problème persiste, contactez le support.",
		})
	}

	return c.NoContent(http.StatusOK)
}
