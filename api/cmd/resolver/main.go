package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ZiplEix/scrabble/api/services"
)

//go:embed web/*
var webFiles embed.FS

type SolveRequest struct {
	Board [15][15]string `json:"board"`
	Rack  string         `json:"rack"`
}

func main() {
	// Serve HTML UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := webFiles.ReadFile("web/index.html")
		if err != nil {
			http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// Serve CSS
	http.HandleFunc("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		data, err := webFiles.ReadFile("web/style.css")
		if err != nil {
			http.Error(w, "Failed to load style.css", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(data)
	})

	// Serve JS
	http.HandleFunc("/static/app.js", func(w http.ResponseWriter, r *http.Request) {
		data, err := webFiles.ReadFile("web/app.js")
		if err != nil {
			http.Error(w, "Failed to load app.js", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write(data)
	})

	// Solver API endpoint
	http.HandleFunc("/solve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req SolveRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
			return
		}

		// Clean the rack input (uppercase, only letters and '?')
		rack := strings.ToUpper(strings.TrimSpace(req.Rack))
		var cleanedRack []rune
		for _, r := range rack {
			if (r >= 'A' && r <= 'Z') || r == '?' {
				cleanedRack = append(cleanedRack, r)
			}
		}

		log.Printf("Solving for rack %q...", string(cleanedRack))

		// Normalize board to uppercase
		var normalizedBoard [15][15]string
		for y := 0; y < 15; y++ {
			for x := 0; x < 15; x++ {
				normalizedBoard[y][x] = strings.ToUpper(strings.TrimSpace(req.Board[y][x]))
			}
		}

		// Run Scrabble Solver!
		bestMove := services.FindBestMoveStandalone(normalizedBoard, string(cleanedRack))

		w.Header().Set("Content-Type", "application/json")
		if bestMove == nil {
			w.Write([]byte(`{}`))
			return
		}

		responseBytes, err := json.Marshal(bestMove)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to marshal response: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(responseBytes)
	})

	port := ":8080"
	log.Printf("Standalone Scrabble Solver server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
