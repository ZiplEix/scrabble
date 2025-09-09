package response

import "time"

// LogResumeEntry is a compact representation of a log line for admin UI
type LogResumeEntry struct {
	ID        int64          `json:"id"`
	Level     string         `json:"level"`
	Date      time.Time      `json:"date"`
	Route     string         `json:"route"`
	Message   string         `json:"message"`
	Raw       map[string]any `json:"raw"`
	RequestID string         `json:"request_id,omitempty"`
	Username  string         `json:"username,omitempty"`
	Method    string         `json:"method,omitempty"`
	Status    int            `json:"status,omitempty"`
	Reason    string         `json:"reason,omitempty"`
}
