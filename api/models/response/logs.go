package response

type LogsStatsResponse struct {
	Labels []string `json:"labels"`
	Data   []int    `json:"data"`
}
