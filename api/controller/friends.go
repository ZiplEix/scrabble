package controller

import (
	"fmt"
	"net/http"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

// AddFriend handles POST /users/friends/:id
func AddFriend(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	friendIDParam := c.Param("id")
	if friendIDParam == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing friend id")
	}

	var friendID int64
	if _, err := fmt.Sscanf(friendIDParam, "%d", &friendID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid friend id")
	}

	if userID == friendID {
		return echo.NewHTTPError(http.StatusBadRequest, "you cannot add yourself as a friend")
	}

	err := services.AddFriend(userID, friendID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_add_friend",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add friend")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "friend added successfully",
	})
}

// RemoveFriend handles DELETE /users/friends/:id
func RemoveFriend(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	friendIDParam := c.Param("id")
	if friendIDParam == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing friend id")
	}

	var friendID int64
	if _, err := fmt.Sscanf(friendIDParam, "%d", &friendID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid friend id")
	}

	err := services.RemoveFriend(userID, friendID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_remove_friend",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove friend")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "friend removed successfully",
	})
}

// GetFriends handles GET /users/friends
func GetFriends(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	friends, err := services.GetFriends(userID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_friends",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get friends")
	}

	return c.JSON(http.StatusOK, friends)
}

// GetRecentOpponents handles GET /users/recent-opponents
func GetRecentOpponents(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	opponents, err := services.GetRecentOpponents(userID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_recent_opponents",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get recent opponents")
	}

	return c.JSON(http.StatusOK, opponents)
}
