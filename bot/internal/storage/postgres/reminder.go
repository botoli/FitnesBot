package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"traningBot/bot/internal/storage/models"
)

func (s *Store) CreateReminder(ctx context.Context, r models.Reminder) (models.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var out models.Reminder
	err := s.DB.QueryRowContext(ctx, `
INSERT INTO reminders (user_id, remind_at, message, is_active, is_recurring, interval_minutes, last_sent_at, exercise_plan_id)
VALUES ($1,$2,$3,TRUE,$4,$5,$6,$7)
RETURNING id, user_id, remind_at, message, is_active, is_recurring, interval_minutes, last_sent_at, exercise_plan_id, created_at
`, r.UserID, r.RemindAt, r.Message, r.IsRecurring, r.IntervalMin, r.LastSentAt, r.ExercisePlanID).
		Scan(&out.ID, &out.UserID, &out.RemindAt, &out.Message, &out.IsActive, &out.IsRecurring, &out.IntervalMin, &out.LastSentAt, &out.ExercisePlanID, &out.CreatedAt)
	if err != nil {
		return models.Reminder{}, err
	}
	return out, nil
}

func (s *Store) GetReminderByID(ctx context.Context, id int64) (models.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var r models.Reminder
	err := s.DB.QueryRowContext(ctx, `
SELECT id, user_id, remind_at, message, is_active, is_recurring, interval_minutes, last_sent_at, exercise_plan_id, created_at
FROM reminders
WHERE id = $1
`, id).Scan(&r.ID, &r.UserID, &r.RemindAt, &r.Message, &r.IsActive, &r.IsRecurring, &r.IntervalMin, &r.LastSentAt, &r.ExercisePlanID, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Reminder{}, errors.New("reminder not found")
	}
	if err != nil {
		return models.Reminder{}, err
	}
	return r, nil
}

func (s *Store) DeactivateReminder(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := s.DB.ExecContext(ctx, `UPDATE reminders SET is_active = FALSE WHERE id = $1`, id)
	return err
}

func (s *Store) TouchReminderSent(ctx context.Context, id int64, sentAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := s.DB.ExecContext(ctx, `UPDATE reminders SET last_sent_at = $2 WHERE id = $1`, id, sentAt)
	return err
}

func (s *Store) SnoozeReminder(ctx context.Context, id int64, nextAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := s.DB.ExecContext(ctx, `
UPDATE reminders
SET remind_at = $2, last_sent_at = NULL
WHERE id = $1
`, id, nextAt)
	return err
}

// ListDueReminders returns reminders that should be sent right now.
// For non-recurring: remind_at <= now and not sent (last_sent_at is NULL) and active.
// For recurring: if last_sent_at is NULL then remind_at <= now; else last_sent_at + interval <= now.
func (s *Store) ListDueReminders(ctx context.Context, now time.Time, limit int) ([]models.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, user_id, remind_at, message, is_active, is_recurring, interval_minutes, last_sent_at, exercise_plan_id, created_at
FROM reminders
WHERE is_active = TRUE
AND (
  (last_sent_at IS NULL AND remind_at <= $1)
  OR
  (last_sent_at IS NOT NULL AND (last_sent_at + (interval_minutes || ' minutes')::interval) <= $1)
)
ORDER BY remind_at ASC
LIMIT $2
`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Reminder
	for rows.Next() {
		var r models.Reminder
		if err := rows.Scan(&r.ID, &r.UserID, &r.RemindAt, &r.Message, &r.IsActive, &r.IsRecurring, &r.IntervalMin, &r.LastSentAt, &r.ExercisePlanID, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}


