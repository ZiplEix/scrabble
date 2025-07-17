package response

type Report struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
	Type      string `json:"type"`
	Username  string `json:"username"` // le username au lieu de user_id
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
