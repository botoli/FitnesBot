package scheduler

import (
	"context"
	"log"
	"time"

	tgbot "github.com/go-telegram/bot"

	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/storage/postgres"
	"traningBot/bot/internal/utils"
)

func RunReminderLoop(ctx context.Context, store *postgres.Store, b *tgbot.Bot) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			due, err := store.ListDueReminders(ctx, now, 50)
			if err != nil {
				log.Println("scheduler: list due reminders:", err)
				continue
			}

			for _, r := range due {
				tgID, err := store.GetTgIDByUserID(ctx, r.UserID)
				if err != nil {
					continue
				}

				st, err := store.GetUserSettings(ctx, r.UserID)
				if err == nil {
					qs, qse := parseQuiet(now, st.QuietStart, st.QuietEnd)
					if utils.InQuietHours(now, qs, qse) {
						// do not send in quiet hours; try again on next ticks
						continue
					}
				}

				text := "⏰ Напоминание!\n" + r.Message
				params := &tgbot.SendMessageParams{
					ChatID:    tgID,
					Text:      text,
				}

				// Spam-until-confirmed: always show inline buttons.
				params.ReplyMarkup = keyboard.ReminderInlineKeyboard(r.ID)

				_, err = b.SendMessage(ctx, params)
				if err != nil {
					continue
				}

				_ = store.TouchReminderSent(ctx, r.ID, now)
			}
		}
	}
}

func parseQuiet(now time.Time, startText string, endText string) (time.Time, time.Time) {
	loc := now.Location()
	qs, _ := time.ParseInLocation("15:04:05", startText, loc)
	qe, _ := time.ParseInLocation("15:04:05", endText, loc)
	qs = time.Date(now.Year(), now.Month(), now.Day(), qs.Hour(), qs.Minute(), 0, 0, loc)
	qe = time.Date(now.Year(), now.Month(), now.Day(), qe.Hour(), qe.Minute(), 0, 0, loc)
	return qs, qe
}

