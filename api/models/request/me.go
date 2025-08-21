package request

type UpdatePrefsRequest struct {
	Turn     *bool `json:"turn,omitempty"`
	Messages *bool `json:"messages,omitempty"`
}
