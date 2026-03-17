package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"traningBot/bot/internal/storage/models"
)

func (s *Store) EnsureUser(ctx context.Context, tgID int64, username string) (models.User, error) {
	var u models.User
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, `
INSERT INTO users (tg_id, username)
VALUES ($1, $2)
ON CONFLICT (tg_id) DO UPDATE SET username = EXCLUDED.username
RETURNING id, tg_id, COALESCE(username,''), created_at
`, tgID, username).Scan(&u.ID, &u.TgID, &u.Username, &u.CreatedAt)
	if err != nil {
		return models.User{}, err
	}

	_, _ = s.EnsureUserSettings(ctx, u.ID)
	return u, nil
}

func (s *Store) GetUserByTgID(ctx context.Context, tgID int64) (models.User, error) {
	var u models.User
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := s.DB.QueryRowContext(ctx, `
SELECT id, tg_id, COALESCE(username,''), created_at
FROM users
WHERE tg_id = $1
`, tgID).Scan(&u.ID, &u.TgID, &u.Username, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, errors.New("user not found")
	}
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (s *Store) GetTgIDByUserID(ctx context.Context, userID int64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var tgID int64
	err := s.DB.QueryRowContext(ctx, `SELECT tg_id FROM users WHERE id = $1`, userID).Scan(&tgID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	}
	return tgID, err
}

func (s *Store) EnsureUserSettings(ctx context.Context, userID int64) (models.UserSettings, error) {
	var st models.UserSettings
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, `
INSERT INTO user_settings (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING user_id, reminder_interval_minutes, quiet_start::text, quiet_end::text, created_at, updated_at
`, userID).Scan(&st.UserID, &st.ReminderIntervalMinutes, &st.QuietStart, &st.QuietEnd, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		return models.UserSettings{}, err
	}
	return st, nil
}

func (s *Store) GetUserSettings(ctx context.Context, userID int64) (models.UserSettings, error) {
	var st models.UserSettings
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := s.DB.QueryRowContext(ctx, `
SELECT user_id, reminder_interval_minutes, quiet_start::text, quiet_end::text, created_at, updated_at
FROM user_settings
WHERE user_id = $1
`, userID).Scan(&st.UserID, &st.ReminderIntervalMinutes, &st.QuietStart, &st.QuietEnd, &st.CreatedAt, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return s.EnsureUserSettings(ctx, userID)
	}
	if err != nil {
		return models.UserSettings{}, err
	}
	return st, nil
}


