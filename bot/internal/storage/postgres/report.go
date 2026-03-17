package postgres

import (
	"context"
	"time"

	"traningBot/bot/internal/storage/models"
)

func (s *Store) CreateReport(ctx context.Context, r models.Report) (models.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var out models.Report
	err := s.DB.QueryRowContext(ctx, `
INSERT INTO reports (user_id, reminder_id, report_text)
VALUES ($1,$2,$3)
RETURNING id, user_id, reminder_id, report_text, created_at
`, r.UserID, r.ReminderID, r.ReportText).
		Scan(&out.ID, &out.UserID, &out.ReminderID, &out.ReportText, &out.CreatedAt)
	if err != nil {
		return models.Report{}, err
	}
	return out, nil
}

func (s *Store) ListReports(ctx context.Context, userID int64, limit int) ([]models.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, user_id, reminder_id, report_text, created_at
FROM reports
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Report
	for rows.Next() {
		var r models.Report
		if err := rows.Scan(&r.ID, &r.UserID, &r.ReminderID, &r.ReportText, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}


