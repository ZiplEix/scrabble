package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func PushSubscribe(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var sub utils.Subscription
	if err := c.Bind(&sub); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid subscription"})
	}

	subBytes, err := json.Marshal(sub)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_serialize_subscription",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to serialize subscription"})
	}

	err = services.PushSubscribe(userID, subBytes)

	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_subscribe",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to subscribe"})
	}

	return c.NoContent(http.StatusOK)
}

func SendTest(c echo.Context) error {
	payload := utils.NotificationPayload{
		Title: "Test Notification",
		Body:  "Hello depuis le serveur 😄",
	}

	err := utils.SendNotificationToUserByID(1, payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": fmt.Sprintf("failed to send test notification: %v", err)})
	}

	return c.NoContent(http.StatusOK)
}

func PushUnsubscribe(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	if err := services.PushUnsubscribe(userID); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_unsubscribe",
			"error":  err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to unsubscribe"})
	}

	return c.NoContent(http.StatusOK)
}
