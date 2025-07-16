package controller

import (
	"net/http"
	"os"

	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	user, err := services.CreateUser(req.Username, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error creating user")
	}

	tokenString, err := utils.GenerateToken(*user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create token")
	}

	return c.JSON(http.StatusCreated, response.AuthResponse{Token: tokenString})
}

func Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	user, err := services.VerifyUser(req.Username, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, failed to verify user")
	}

	tokenString, err := utils.GenerateToken(*user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create token")
	}

	return c.JSON(http.StatusOK, response.AuthResponse{Token: tokenString})
}
