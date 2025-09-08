package controller

import (
	"fmt"
	"net/http"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func SuggestUsers(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	query := c.QueryParam("q")
	if len(query) < 2 {
		return c.JSON(http.StatusOK, response.SuggestUsersResponse{})
	}

	suggestions, err := services.SuggestUsers(userID, query)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_fetch_suggestions",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch suggestions")
	}

	return c.JSON(http.StatusOK, suggestions)
}

// GetUserPublic retourne les informations publiques d'un utilisateur par id
func GetUserPublic(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "" {
		logctx.Add(c, "reason", "missing_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	// Convert id to int64
	var uid int64
	_, err := fmt.Sscanf(idParam, "%d", &uid)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_parse_id",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	user, err := services.GetUserPublicByID(uid)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_fetch_user",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch user")
	}
	if user == nil {
		logctx.Merge(c, map[string]any{
			"reason":  "user_not_found",
			"user_id": uid,
		})
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, user)
}
