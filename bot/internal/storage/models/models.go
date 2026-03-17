package models

import "time"

type User struct {
	ID        int64
	TgID      int64
	Username  string
	CreatedAt time.Time
}

type UserSettings struct {
	UserID                  int64
	ReminderIntervalMinutes int
	QuietStart              string // HH:MM:SS in DB, keep as string for simplicity
	QuietEnd                string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type Plan struct {
	ID        int64
	UserID    int64
	DayOfWeek int // 1=Mon..7=Sun
	Title     string
	Details   string
	CreatedAt time.Time
}

type Reminder struct {
	ID             int64
	UserID         int64
	RemindAt       time.Time
	Message        string
	IsActive       bool
	IsRecurring    bool
	IntervalMin    int
	LastSentAt     *time.Time
	ExercisePlanID *int64
	CreatedAt      time.Time
}

type Report struct {
	ID        int64
	UserID    int64
	ReminderID *int64
	ReportText string
	CreatedAt time.Time
}

