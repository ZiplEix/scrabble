package utils

import (
	"github.com/labstack/echo/v4"
)

const UserIDKey = "user_id"

func GetUserID(c echo.Context) (int64, bool) {
	val := c.Get(UserIDKey)
	id, ok := val.(int64)
	return id, ok
}
