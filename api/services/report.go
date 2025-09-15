package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/models/response"
)

func CreateReport(userID int64, title, content, rType string) (int64, error) {
	var reportID int64
	err := database.DB.QueryRow(`
		INSERT INTO reports (user_id, title, content, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userID, title, content, rType, time.Now(), time.Now()).Scan(&reportID)

	if err != nil {
		return 0, fmt.Errorf("failed to insert report: %w", err)
	}

	return reportID, nil
}

func GetReportByID(reportID string) (*response.Report, error) {
	query := `
		SELECT r.id, r.title, r.content, r.status, r.priority, r.type,
		       COALESCE(u.username, 'Utilisateur supprimé') AS username,
		       r.created_at, r.updated_at
		FROM reports r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.id = $1
	`

	var report response.Report
	err := database.DB.QueryRow(query, reportID).Scan(
		&report.ID,
		&report.Title,
		&report.Content,
		&report.Status,
		&report.Priority,
		&report.Type,
		&report.Username,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &report, nil
}

func GetAllReports() ([]response.Report, error) {
	query := `
		SELECT r.id, r.title, r.content, r.status, r.priority, r.type,
		       COALESCE(u.username, 'Utilisateur supprimé') AS username,
		       r.created_at, r.updated_at
		FROM reports r
		LEFT JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var reports []response.Report
	for rows.Next() {
		var r response.Report
		err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Content,
			&r.Status,
			&r.Priority,
			&r.Type,
			&r.Username,
			&r.CreatedAt,
			&r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		reports = append(reports, r)
	}

	return reports, nil
}

func UpdateReportStatus(reportID string, status string) error {
	query := `
		UPDATE reports
		SET status = $1
		WHERE id = $2
	`

	result, err := database.DB.Exec(query, status, reportID)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no report found with id %s", reportID)
	}

	return nil
}

func DeleteReport(reportID string) error {
	result, err := database.DB.Exec(`DELETE FROM reports WHERE id = $1`, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no report found with id %s", reportID)
	}

	return nil
}

func UpdateReport(reportID string, req request.UpdateReportRequest) error {
	// Construction dynamique de la requête SQL
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != "" {
		updates = append(updates, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, req.Title)
		argIndex++
	}
	if req.Content != "" {
		updates = append(updates, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, req.Content)
		argIndex++
	}
	if req.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.Type != "" {
		updates = append(updates, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, req.Type)
		argIndex++
	}

	if len(updates) == 0 {
		return fmt.Errorf("nothing to update")
	}

	args = append(args, reportID)
	query := fmt.Sprintf(`
		UPDATE reports
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no report found with id %s", reportID)
	}

	return nil
}

func GetReportsByUserID(userID int64) ([]map[string]any, error) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.title, r.content, r.status, r.priority, r.type, r.created_at, r.updated_at, u.username
		FROM reports r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []map[string]any

	for rows.Next() {
		var r struct {
			ID        int
			Title     string
			Content   string
			Status    string
			Priority  string
			Type      string
			CreatedAt time.Time
			UpdatedAt time.Time
			Username  *string
		}

		if err := rows.Scan(
			&r.ID, &r.Title, &r.Content, &r.Status,
			&r.Priority, &r.Type, &r.CreatedAt, &r.UpdatedAt, &r.Username,
		); err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}

		reports = append(reports, map[string]any{
			"id":         r.ID,
			"title":      r.Title,
			"content":    r.Content,
			"status":     r.Status,
			"priority":   r.Priority,
			"type":       r.Type,
			"created_at": r.CreatedAt,
			"updated_at": r.UpdatedAt,
			"username":   r.Username,
		})
	}

	return reports, nil
}
