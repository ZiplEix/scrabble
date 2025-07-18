package controller

import (
	"encoding/json"
	"net/http"

	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func PushSubscribe(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var sub utils.Subscription
	if err := c.Bind(&sub); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid subscription"})
	}

	subBytes, err := json.Marshal(sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to serialize subscription"})
	}

	err = services.PushSubscribe(userID, subBytes)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to subscribe"})
	}

	return c.NoContent(http.StatusOK)
}

func SendTest(c echo.Context) error {
	payload, _ := json.Marshal(map[string]string{
		"title": "Test Notification",
		"body":  "Hello depuis le serveur 😄",
	})

	err := utils.SendNotificationToUserByID(3, string(payload))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to send notification"})
	}

	return c.NoContent(http.StatusOK)
}
