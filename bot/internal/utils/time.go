package utils

import (
	"errors"
	"strings"
	"time"
)

// DefaultWorkoutReminderMessage — текст напоминания, если пользователь указал только дату и время.
const DefaultWorkoutReminderMessage = "🏋️ Время тренировки! Открой бота и нажми «Я позанималась», когда закончишь."

// ParseUserReminderInput parses: "DD.MM.YYYY HH:MM some text"
func ParseUserReminderInput(s string, now time.Time, loc *time.Location) (time.Time, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, "", errors.New("empty input")
	}

	parts := strings.Fields(s)
	if len(parts) < 2 {
		return time.Time{}, "", errors.New("expected: DD.MM.YYYY HH:MM [текст]")
	}

	dtStr := parts[0] + " " + parts[1]
	t, err := time.ParseInLocation("02.01.2006 15:04", dtStr, loc)
	if err != nil {
		return time.Time{}, "", err
	}

	msg := ""
	if len(parts) >= 3 {
		msg = strings.TrimSpace(strings.Join(parts[2:], " "))
	}
	if msg == "" {
		msg = DefaultWorkoutReminderMessage
	}

	// Prevent creating reminders in the past by a lot; allow a small drift.
	if t.Before(now.Add(-30 * time.Second)) {
		return time.Time{}, "", errors.New("time is in the past")
	}

	return t, msg, nil
}

// NextClockOnOrAfter — ближайшее сегодня hour:min в loc; если уже прошло — завтра в то же время.
func NextClockOnOrAfter(now time.Time, hour, min int, loc *time.Location) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	if !t.After(now.Add(30 * time.Second)) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// TomorrowAt — завтра в hour:min (локальное время).
func TomorrowAt(hour, min int, now time.Time, loc *time.Location) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	return t.Add(24 * time.Hour)
}

func InQuietHours(now time.Time, quietStart time.Time, quietEnd time.Time) bool {
	// If quietStart <= quietEnd: quiet is [start, end)
	// If wraps over midnight: quiet is [start, 24h) U [0, end)
	if !quietStart.After(quietEnd) {
		return !now.Before(quietStart) && now.Before(quietEnd)
	}
	return !now.Before(quietStart) || now.Before(quietEnd)
}


