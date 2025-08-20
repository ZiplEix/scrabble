package request

// CreateMessageRequest représente le payload pour créer un message de partie
type CreateMessageRequest struct {
	Content string                 `json:"content"`
	Meta    map[string]interface{} `json:"meta"`
}
