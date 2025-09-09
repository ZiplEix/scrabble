package services

import (
	"database/sql"
	"encoding/json"
	"strconv"
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

// GetLogsResume returns the last N logs (default 10) with parsed fields from the raw JSONB
func GetLogsResume(limit int) ([]response.LogResumeEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := database.Query(`
		SELECT id, received_at, raw, req_id
		FROM logs
		ORDER BY received_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.LogResumeEntry
	for rows.Next() {
		var id int64
		var receivedAt time.Time
		var raw string
		var reqID sql.NullString

		if err := rows.Scan(&id, &receivedAt, &raw, &reqID); err != nil {
			return nil, err
		}

		// try to parse raw JSON as a map
		parsed := map[string]any{}
		_ = json.Unmarshal([]byte(raw), &parsed)

		// extract common fields
		level := "info"
		if v, ok := parsed["level"]; ok {
			if s, ok := v.(string); ok {
				level = s
			}
		}
		msg := ""
		if v, ok := parsed["msg"]; ok {
			if s, ok := v.(string); ok {
				msg = s
			}
		}
		route := ""
		if v, ok := parsed["route"]; ok {
			if s, ok := v.(string); ok {
				route = s
			}
		}

		entry := response.LogResumeEntry{
			ID:        id,
			Level:     level,
			Date:      receivedAt,
			Route:     route,
			Message:   msg,
			Raw:       parsed,
			RequestID: reqID.String,
		}
		out = append(out, entry)
	}

	return out, nil
}

// GetLogs renvoie les logs paginés (page=1 => logs 0-49, page=2 => 50-99, etc). Si page <= 0, renvoie tous les logs.
func GetLogs(page int) ([]response.LogResumeEntry, error) {
	const pageSize = 50
	var rows *sql.Rows
	var err error
	if page > 0 {
		offset := (page - 1) * pageSize
		rows, err = database.Query(`
			 SELECT id, received_at, raw, req_id
			 FROM logs
			 ORDER BY received_at DESC
			 LIMIT $1 OFFSET $2
		 `, pageSize, offset)
	} else {
		rows, err = database.Query(`
			 SELECT id, received_at, raw, req_id
			 FROM logs
			 ORDER BY received_at DESC
		 `)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.LogResumeEntry
	for rows.Next() {
		var id int64
		var receivedAt time.Time
		var raw string
		var reqID sql.NullString

		if err := rows.Scan(&id, &receivedAt, &raw, &reqID); err != nil {
			return nil, err
		}

		// try to parse raw JSON as a map
		parsed := map[string]any{}
		_ = json.Unmarshal([]byte(raw), &parsed)

		// extract common fields
		level := "info"
		if v, ok := parsed["level"]; ok {
			if s, ok := v.(string); ok {
				level = s
			}
		}
		msg := ""
		if v, ok := parsed["msg"]; ok {
			if s, ok := v.(string); ok {
				msg = s
			}
		}
		route := ""
		if v, ok := parsed["route"]; ok {
			if s, ok := v.(string); ok {
				route = s
			}
		}
		username := ""
		if v, ok := parsed["username"]; ok {
			if s, ok := v.(string); ok {
				username = s
			}
		}
		method := ""
		if v, ok := parsed["method"]; ok {
			if s, ok := v.(string); ok {
				method = s
			}
		}
		status := 0
		if v, ok := parsed["status"]; ok {
			switch vv := v.(type) {
			case float64:
				status = int(vv)
			case int:
				status = vv
			case string:
				// try to parse
				if n, err := strconv.Atoi(vv); err == nil {
					status = n
				}
			}
		}
		reason := ""
		if v, ok := parsed["reason"]; ok {
			if s, ok := v.(string); ok {
				reason = s
			}
		}

		entry := response.LogResumeEntry{
			ID:        id,
			Level:     level,
			Date:      receivedAt,
			Route:     route,
			Message:   msg,
			Raw:       parsed,
			RequestID: reqID.String,
			Username:  username,
			Method:    method,
			Status:    status,
			Reason:    reason,
		}
		out = append(out, entry)
	}

	return out, nil
}
