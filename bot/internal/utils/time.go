package utils

import (
	"errors"
	"strings"
	"time"
)

// ParseUserReminderInput parses: "DD.MM.YYYY HH:MM some text"
func ParseUserReminderInput(s string, now time.Time, loc *time.Location) (time.Time, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, "", errors.New("empty input")
	}

	parts := strings.Fields(s)
	if len(parts) < 3 {
		return time.Time{}, "", errors.New("expected: DD.MM.YYYY HH:MM text")
	}

	dtStr := parts[0] + " " + parts[1]
	t, err := time.ParseInLocation("02.01.2006 15:04", dtStr, loc)
	if err != nil {
		return time.Time{}, "", err
	}

	msg := strings.TrimSpace(strings.Join(parts[2:], " "))
	if msg == "" {
		return time.Time{}, "", errors.New("message is empty")
	}

	// Prevent creating reminders in the past by a lot; allow a small drift.
	if t.Before(now.Add(-30 * time.Second)) {
		return time.Time{}, "", errors.New("time is in the past")
	}

	return t, msg, nil
}

func InQuietHours(now time.Time, quietStart time.Time, quietEnd time.Time) bool {
	// If quietStart <= quietEnd: quiet is [start, end)
	// If wraps over midnight: quiet is [start, 24h) U [0, end)
	if !quietStart.After(quietEnd) {
		return !now.Before(quietStart) && now.Before(quietEnd)
	}
	return !now.Before(quietStart) || now.Before(quietEnd)
}


