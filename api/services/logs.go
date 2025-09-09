package services

import (
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/response"
)

// GetLogsStats returns labels (hour strings) and counts per hour for last 48 hours
func GetLogsStats() (*response.LogsStatsResponse, error) {
	// query counts grouped by hour
	rows, err := database.Query(`
        SELECT date_trunc('hour', received_at) AS hr, COUNT(*)
        FROM logs
        WHERE received_at >= now() - interval '48 hours'
        GROUP BY hr
        ORDER BY hr
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[time.Time]int{}
	for rows.Next() {
		var hr time.Time
		var cnt int
		if err := rows.Scan(&hr, &cnt); err != nil {
			return nil, err
		}
		counts[hr.UTC()] = cnt
	}

	// build labels and data for the last 48 hours (older -> newer)
	labels := make([]string, 48)
	data := make([]int, 48)
	now := time.Now().UTC()
	// start at the hour 47 hours ago
	start := now.Add(-47 * time.Hour)
	// normalize to the hour
	start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, time.UTC)

	for i := 0; i < 48; i++ {
		t := start.Add(time.Duration(i) * time.Hour)
		// ISO timestamp in UTC
		labels[i] = t.Format(time.RFC3339)
		if v, ok := counts[t]; ok {
			data[i] = v
		} else {
			data[i] = 0
		}
	}

	return &response.LogsStatsResponse{Labels: labels, Data: data}, nil
}
