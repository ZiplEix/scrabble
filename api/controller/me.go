package controller

import (
	"net/http"

	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

// GetMe retourne les informations basiques de l'utilisateur connecté
func GetMe(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	me, err := services.GetMeInfo(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch user info")
	}

	return c.JSON(http.StatusOK, me)
}
