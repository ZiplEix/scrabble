package request

type CreateGameRequest struct {
	Name    string   `json:"name"`
	Players []string `json:"players"` // liste des usernames, ex: ["Alice", "Bob"]
}

type RenameGameRequest struct {
	NewName string `json:"new_name"`
}

type PlayMoveRequest struct {
	Word      string         `json:"word"` // ex: "CHAT"
	StartX    int            `json:"x"`    // position de départ
	StartY    int            `json:"y"`
	Direction string         `json:"dir"`     // "H" ou "V"
	Letters   []PlacedLetter `json:"letters"` // lettres posées ce tour
	Score     int            `json:"score"`   // score du coup
}

type PlacedLetter struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Char string `json:"char"` // toujours en majuscules
}
