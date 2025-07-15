package controller

import (
	"encoding/json"
	"net/http"

	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/go-chi/chi/v5"
)

func CreateGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized, no user_id", http.StatusUnauthorized)
		return
	}

	var req request.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "game name is required", http.StatusBadRequest)
		return
	}

	gameID, err := services.CreateGame(userID, req.Name)
	if err != nil {
		http.Error(w, "failed to create game", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Game created",
		"game_id": gameID,
	})
}

func GetGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized, no user_id", http.StatusUnauthorized)
		return
	}

	gameID := chi.URLParam(r, "id")
	if gameID == "" {
		http.Error(w, "missing game id", http.StatusBadRequest)
		return
	}

	game, err := services.GetGameDetails(userID, gameID)
	if err != nil {
		http.Error(w, "failed to load game: "+err.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(game)
}

func PlayMove(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	gameID := chi.URLParam(r, "id")
	if gameID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing game id"})
		return
	}

	var req request.PlayMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if err := services.PlayMove(gameID, userID, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to play move: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "move played successfully",
	})
}

func GetUserGames(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	games, err := services.GetGamesByUserID(userID)
	if err != nil {
		http.Error(w, "Failed to get games: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := response.GamesListResponse{
		Games: games,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
