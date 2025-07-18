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
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no user_id")
	}

	var req request.CreateGameRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "game name is required")
	}

	var usernames []string
	for _, player := range req.Players {
		usernames = append(usernames, strings.ToLower(strings.TrimSpace(player)))
	}

	gameID, err := services.CreateGame(userID, req.Name, usernames)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to create game: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Game created",
		"game_id": gameID,
	})
}

func DeleteGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no user_id")
	}

	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}

	if err := services.DeleteGame(userID, gameID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete game: "+err.Error())
	}

	return c.JSON(http.StatusNoContent, map[string]string{
		"message": "Game deleted successfully",
	})
}

func RenameGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no user_id")
	}

	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}

	var req request.RenameGameRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := services.RenameGame(userID, gameID, req.NewName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to rename game: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Game renamed successfully",
	})
}

func GetGame(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no user_id")
	}

	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}

	game, err := services.GetGameDetails(userID, gameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "failed to load game: "+err.Error())
	}

	return c.JSON(http.StatusOK, game)
}

func PlayMove(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}

	var req request.PlayMoveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := services.PlayMove(gameID, userID, req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to play move: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "move played successfully",
	})
}

func GetNewRack(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}

	newRack, err := services.GetNewRack(userID, gameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get new rack: "+err.Error())
	}

	return c.JSON(http.StatusOK, newRack)
}

func GetUserGames(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	games, err := services.GetGamesByUserID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get games: "+err.Error())
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
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing game id")
	}

	err := services.PassTurn(userID, gameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
