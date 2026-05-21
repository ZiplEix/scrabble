package request

type SubmitPuzzleAttemptRequest struct {
	PuzzleID    string                `json:"puzzle_id"`
	WordsPlayed []PuzzleWordForSubmit `json:"words_played"`
	Letters     []PlacedLetter        `json:"letters,omitempty"`
	// time_used n'est plus envoyé par le client — calculé côté serveur depuis started_at
}

type PuzzleWordForSubmit struct {
	Word      string `json:"word"`
	Position  string `json:"position"`  // e.g., "7,7"
	Direction string `json:"direction"` // "horizontal" or "vertical"
}
