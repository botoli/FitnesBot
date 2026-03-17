package postgres

import (
	"context"
	"time"

	"traningBot/bot/internal/storage/models"
)

func (s *Store) UpsertPlan(ctx context.Context, p models.Plan) (models.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// simple upsert by (user_id, day_of_week) using a unique constraint emulated via delete+insert
	_, _ = s.DB.ExecContext(ctx, `DELETE FROM plans WHERE user_id = $1 AND day_of_week = $2`, p.UserID, p.DayOfWeek)

	var out models.Plan
	err := s.DB.QueryRowContext(ctx, `
INSERT INTO plans (user_id, day_of_week, title, details)
VALUES ($1,$2,$3,$4)
RETURNING id, user_id, day_of_week, title, details, created_at
`, p.UserID, p.DayOfWeek, p.Title, p.Details).
		Scan(&out.ID, &out.UserID, &out.DayOfWeek, &out.Title, &out.Details, &out.CreatedAt)
	if err != nil {
		return models.Plan{}, err
	}
	return out, nil
}

func (s *Store) ListPlansForWeek(ctx context.Context, userID int64) ([]models.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, user_id, day_of_week, title, details, created_at
FROM plans
WHERE user_id = $1
ORDER BY day_of_week ASC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Plan
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.UserID, &p.DayOfWeek, &p.Title, &p.Details, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

