package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var nonImplementedHandler = func(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}

func SetupRoutes(e *echo.Echo) {
	setupAuthRoutes(e)
	setupGameRoutes(e)
	setupReportRoutes(e)
	setupUsersRoutes(e)
}
