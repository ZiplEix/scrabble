package controller

import (
	"net/http"

	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func SuggestUsers(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	query := c.QueryParam("q")
	if len(query) < 2 {
		return c.JSON(http.StatusOK, response.SuggestUsersResponse{})
	}

	suggestions, err := services.SuggestUsers(userID, query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch suggestions")
	}

	return c.JSON(http.StatusOK, suggestions)
}
