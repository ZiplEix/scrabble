package request

type CreateReportRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateReportRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Status  string `json:"status,omitempty"` // open, in_progress, resolved, rejected
}
