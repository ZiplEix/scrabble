package utils

import (
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func GetUserID(c echo.Context) (int64, bool) {
	val := c.Get(middleware.UserIDKey)
	id, ok := val.(int64)
	return id, ok
}
