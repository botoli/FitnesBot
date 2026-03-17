package postgres

import (
	"context"
	"time"

	"traningBot/bot/internal/storage/models"
)

func (s *Store) UpdateReminderInterval(ctx context.Context, userID int64, minutes int) (models.UserSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var st models.UserSettings
	err := s.DB.QueryRowContext(ctx, `
UPDATE user_settings
SET reminder_interval_minutes = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING user_id, reminder_interval_minutes, quiet_start::text, quiet_end::text, created_at, updated_at
`, userID, minutes).Scan(&st.UserID, &st.ReminderIntervalMinutes, &st.QuietStart, &st.QuietEnd, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		return models.UserSettings{}, err
	}
	return st, nil
}

func (s *Store) UpdateQuietHours(ctx context.Context, userID int64, startHHMM string, endHHMM string) (models.UserSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var st models.UserSettings
	err := s.DB.QueryRowContext(ctx, `
UPDATE user_settings
SET quiet_start = $2::time, quiet_end = $3::time, updated_at = NOW()
WHERE user_id = $1
RETURNING user_id, reminder_interval_minutes, quiet_start::text, quiet_end::text, created_at, updated_at
`, userID, startHHMM, endHHMM).Scan(&st.UserID, &st.ReminderIntervalMinutes, &st.QuietStart, &st.QuietEnd, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		return models.UserSettings{}, err
	}
	return st, nil
}

