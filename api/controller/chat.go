package controller

import (
	"fmt"
	"net/http"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

// CreateMessage handles POST /game/:id/message
func CreateMessage(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour envoyer un message",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis pour envoyer un message",
		})
	}

	var req request.CreateMessageRequest
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

	if req.Content == "" {
		logctx.Add(c, "reason", "empty_content")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "empty content",
			"message": "Le contenu du message ne peut pas être vide",
		})
	}

	// validate user is in game
	inGame, err := services.IsUserInGame(userID, gameID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_validate_user_in_game",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to validate user in game: %v", err),
			"message": "Erreur lors de la validation des permissions",
		})
	}
	if !inGame {
		logctx.Add(c, "reason", "user_not_in_game")
		return c.JSON(http.StatusForbidden, echo.Map{
			"error":   "forbidden",
			"message": "Vous ne faites pas partie de cette partie",
		})
	}

	msg, err := services.CreateMessage(userID, gameID, req.Content, req.Meta)
	if err != nil {
		// logctx.Add(c, "reason", "failed_to_create_message")
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_create_message",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create message: %v", err),
			"message": "Erreur lors de l'envoi du message, veuillez réessayer",
		})
	}

	return c.JSON(http.StatusCreated, msg)
}

// GetMessages handles GET /game/:id/messages
func GetMessages(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour récupérer les messages",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis",
		})
	}

	msgs, err := services.GetMessages(userID, gameID)
	if err != nil {
		if err.Error() == "user not in game" {
			logctx.Add(c, "reason", "user_not_in_game")
			return c.JSON(http.StatusForbidden, echo.Map{
				"error":   "forbidden",
				"message": "Vous ne faites pas partie de cette partie",
			})
		}
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_messages",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to get messages: %v", err),
			"message": "Erreur lors de la récupération des messages",
		})
	}

	return c.JSON(http.StatusOK, msgs)
}

// DeleteMessage handles DELETE /game/:id/messages/:msg_id
func DeleteMessage(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   "unauthorized, no user_id",
			"message": "Vous devez être connecté pour supprimer un message",
		})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing game id",
			"message": "L'ID de la partie est requis",
		})
	}

	msgID := c.Param("msg_id")
	if msgID == "" {
		logctx.Add(c, "reason", "missing_message_id")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "missing message id",
			"message": "L'ID du message est requis",
		})
	}

	if err := services.DeleteMessage(userID, gameID, msgID); err != nil {
		switch err.Error() {
		case "not found":
			logctx.Add(c, "reason", "message_not_found")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found", "message": "Message introuvable"})
		case "forbidden":
			logctx.Add(c, "reason", "forbidden")
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden", "message": "Vous n'êtes pas autorisé à supprimer ce message"})
		default:
			logctx.Merge(c, map[string]any{
				"reason": "failed_to_delete_message",
				"error":  err.Error(),
			})
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": fmt.Sprintf("failed to delete message: %v", err), "message": "Erreur lors de la suppression du message"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// MarkMessagesReadHandler handles POST /game/:id/messages/read
func MarkMessagesReadHandler(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized, no user_id"})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing game id"})
	}

	var body struct {
		LastMessageID *int64 `json:"last_message_id,omitempty"`
	}
	if err := c.Bind(&body); err != nil {
		// ignore body parse errors and continue with zero -> server will take latest
	}

	var lastID int64
	if body.LastMessageID != nil {
		lastID = *body.LastMessageID
	}

	if err := services.MarkMessagesRead(userID, gameID, lastID); err != nil {
		if err.Error() == "user not in game" {
			logctx.Add(c, "reason", "user_not_in_game")
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden"})
		}
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_mark_messages_read",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// GetUnreadCountForGameHandler handles GET /game/:id/unread_messages_count
func GetUnreadCountForGameHandler(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized, no user_id"})
	}

	gameID := c.Param("id")
	if gameID == "" {
		logctx.Add(c, "reason", "missing_game_id")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing game id"})
	}

	cnt, err := services.GetUnreadCountForUserInGame(userID, gameID)
	if err != nil {
		if err.Error() == "user not in game" {
			logctx.Add(c, "reason", "user_not_in_game")
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden"})
		}
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_unread_count",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"unread_count": cnt})
}
