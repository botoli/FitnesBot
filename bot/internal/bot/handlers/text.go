package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/bot/state"
	stmodels "traningBot/bot/internal/storage/models"
	"traningBot/bot/internal/utils"
)

func Text(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		p, ok := app.State.Get(tgID)
		if ok {
			switch p.Kind {
			case state.PendingDoneReport:
				u, err := app.Store.GetUserByTgID(ctx, tgID)
				if err != nil {
					app.State.Clear(tgID)
					return
				}
				var reminderID *int64
				if p.ReminderID != nil {
					reminderID = p.ReminderID
				}
				_, err = app.Store.CreateReport(ctx, stmodels.Report{
					UserID:     u.ID,
					ReminderID: reminderID,
					ReportText: text,
				})
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла записать отчет."})
					return
				}
				app.State.Clear(tgID)
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Записала! Так держать.", ReplyMarkup: keyboard.MainMenuInlineKeyboard()})
				return

			case state.PendingRemind:
				u, err := app.Store.GetUserByTgID(ctx, tgID)
				if err != nil {
					app.State.Clear(tgID)
					return
				}
				st, _ := app.Store.GetUserSettings(ctx, u.ID)
				tm, msg, err := utils.ParseUserReminderInput(text, time.Now(), time.Local)
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
						ChatID:      chatID,
						Text:        "Не распознала. Формат: DD.MM.YYYY HH:MM текст",
						ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
					})
					return
				}
				_, err = app.Store.CreateReminder(ctx, stmodels.Reminder{
					UserID:      u.ID,
					RemindAt:    tm,
					Message:     msg,
					IsRecurring: false,
					IntervalMin: st.ReminderIntervalMinutes,
				})
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла сохранить напоминание."})
					return
				}
				app.State.Clear(tgID)
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Ок, напомню в указанное время.", ReplyMarkup: keyboard.MainMenuInlineKeyboard()})
				return

			case state.PendingSettings:
				u, err := app.Store.GetUserByTgID(ctx, tgID)
				if err != nil {
					app.State.Clear(tgID)
					return
				}
				parts := strings.Fields(text)
				if len(parts) == 0 {
					return
				}
				switch parts[0] {
				case "freq":
					if len(parts) != 2 {
						_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Формат: `freq 10`", ParseMode: models.ParseModeMarkdown})
						return
					}
					mins, err := strconv.Atoi(parts[1])
					if err != nil || mins < 1 || mins > 120 {
						_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Минуты должны быть 1..120."})
						return
					}
					if _, err := app.Store.UpdateReminderInterval(ctx, u.ID, mins); err != nil {
						_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла обновить частоту."})
						return
					}
					app.State.Clear(tgID)
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Готово. Частота обновлена.", ReplyMarkup: keyboard.MainMenuInlineKeyboard()})
					return
				case "quiet":
					if len(parts) != 3 {
						_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Формат: `quiet 23:00 08:00`", ParseMode: models.ParseModeMarkdown})
						return
					}
					if _, err := app.Store.UpdateQuietHours(ctx, u.ID, parts[1], parts[2]); err != nil {
						_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла обновить тихие часы."})
						return
					}
					app.State.Clear(tgID)
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Готово. Тихие часы обновлены.", ReplyMarkup: keyboard.MainMenuInlineKeyboard()})
					return
				default:
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Команда не распознана. Примеры: `freq 10`, `quiet 23:00 08:00`", ParseMode: models.ParseModeMarkdown})
					return
				}
			}
		}

		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Я рядом. Выбирай действие кнопками внизу.",
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
	}
}

