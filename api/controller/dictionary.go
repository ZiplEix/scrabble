package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/labstack/echo/v4"
)

type SaveDefinitionRequest struct {
	Word        string          `json:"word"`
	Definitions json.RawMessage `json:"definitions"`
}

func GetDictionaryDefinition(c echo.Context) error {
	word := c.Param("word")
	if word == "" {
		logctx.Add(c, "reason", "missing_word")
		return echo.NewHTTPError(http.StatusBadRequest, "missing word")
	}

	definitions, err := services.GetDictionaryDefinition(word)
	if err != nil {
		if err == sql.ErrNoRows {
			logctx.Add(c, "reason", "not_found")
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not_found"})
		}
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_fetch_definition",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch definition")
	}

	return c.Blob(http.StatusOK, "application/json", []byte(definitions))
}

func SaveDictionaryDefinition(c echo.Context) error {
	var req SaveDefinitionRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_bind_request",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Word == "" {
		logctx.Add(c, "reason", "missing_word")
		return echo.NewHTTPError(http.StatusBadRequest, "word is required")
	}

	definitionsJson, err := json.Marshal(req.Definitions)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_marshal_definitions",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process definitions")
	}

	err = services.SaveDictionaryDefinition(req.Word, string(definitionsJson))
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_save_definition",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save definition")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "saved"})
}
