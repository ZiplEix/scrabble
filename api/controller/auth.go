package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   "RegisterRequest",
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	logctx.Add(c, "username", username)
	if username == "" {
		logctx.Add(c, "reason", "username_required")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "username is required",
			"message": "Le nom d'utilisateur est requis",
		})
	}

	user, err := services.CreateUser(username, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			logctx.Add(c, "reason", "username_already_exists")
			return c.JSON(http.StatusConflict, echo.Map{
				"error":   fmt.Sprintf("username %s already exists: %v", username, err),
				"message": "Le nom d'utilisateur existe déjà, veuillez en choisir un autre",
			})
		}

		logctx.Add(c, "reason", "user_creation_failed")
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create user: %v", err),
			"message": "Erreur lors de la création de l'utilisateur, veuillez vérifier",
		})
	}

	tokenString, err := utils.GenerateToken(*user)
	if err != nil {
		logctx.Add(c, "reason", "token_generation_failed")
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create token for user %s: %v", username, err),
			"message": "Erreur interne du serveur lors de la création du token d'authentification, veuillez réessayer plus tard",
		})
	}

	return c.JSON(http.StatusCreated, response.AuthResponse{Token: tokenString})
}

func Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   "LoginRequest",
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	logctx.Add(c, "username", username)

	user, err := services.VerifyUser(username, req.Password)
	if err != nil {
		logctx.Add(c, "reason", "invalid_credentials")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   fmt.Sprintf("failed to verify user %s: %v", username, err),
			"message": "Mot de passe ou nom d'utilisateur incorrect",
		})
	}

	tokenString, err := utils.GenerateToken(*user)
	if err != nil {
		logctx.Add(c, "reason", "token_generation_failed")
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create token for user %s: %v", username, err),
			"message": "Erreur interne du serveur lors de la création du token d'authentification, veuillez réessayer plus tard",
		})
	}

	return c.JSON(http.StatusOK, response.AuthResponse{Token: tokenString})
}

func AdminLogin(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   "LoginRequest",
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	logctx.Add(c, "username", username)

	user, err := services.VerifyAdmin(username, req.Password)
	if err != nil {
		logctx.Add(c, "reason", "invalid_credentials")
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error":   fmt.Sprintf("failed to verify user %s: %v", username, err),
			"message": "Mot de passe ou nom d'utilisateur incorrect",
		})
	}

	tokenString, err := utils.GenerateToken(*user)
	if err != nil {
		logctx.Add(c, "reason", "token_generation_failed")
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create token for user %s: %v", username, err),
			"message": "Erreur interne du serveur lors de la création du token d'authentification, veuillez réessayer plus tard",
		})
	}

	return c.JSON(http.StatusOK, response.AuthResponse{Token: tokenString})
}

func ChangePassword(c echo.Context) error {
	var req request.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   "ChangePasswordRequest",
		})
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	if username == "" {
		logctx.Add(c, "reason", "username_required")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "username is required",
			"message": "Le nom d'utilisateur est requis",
		})
	}

	err := services.UpdateUserPassword(username, req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to change password for user %s: %v", username, err),
			"message": "Erreur lors du changement de mot de passe, veuillez réessayer plus tard",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Mot de passe changé avec succès"})
}

func ConnectAS(c echo.Context) error {
	// get the user from the query parameter "user"
	username := strings.ToLower(strings.TrimSpace(c.QueryParam("user")))

	user, err := services.GetUserByUsername(username)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error":   fmt.Sprintf("user %s not found: %v", username, err),
			"message": "Utilisateur non trouvé",
		})
	}

	tokenString, err := utils.GenerateToken(*user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   fmt.Sprintf("failed to create token for user %s: %v", username, err),
			"message": "Erreur interne du serveur lors de la création du token d'authentification, veuillez réessayer plus tard",
		})
	}
	return c.JSON(http.StatusOK, response.AuthResponse{Token: tokenString})
}
