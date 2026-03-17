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

func Remind(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		chatID := update.Message.Chat.ID
		tgID := update.Message.From.ID
		username := update.Message.From.Username

		_, err := app.Store.EnsureUser(ctx, tgID, username)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось открыть профиль."})
			return
		}

		app.State.Set(tgID, state.Pending{Kind: state.PendingRemind})
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Напиши дату, время и текст напоминания.\nПример: 20.03.2026 19:00 Купить протеин",
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
	}
}

func HandleReminderCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		data := update.CallbackQuery.Data
		if update.CallbackQuery.Message.Message == nil {
			return
		}
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		msgID := update.CallbackQuery.Message.Message.ID
		tgID := update.CallbackQuery.From.ID

		// Always ack callback to stop spinner
		_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

		if strings.HasPrefix(data, "snooze_") {
			id, err := strconv.ParseInt(strings.TrimPrefix(data, "snooze_"), 10, 64)
			if err != nil {
				return
			}
			r, err := app.Store.GetReminderByID(ctx, id)
			if err != nil || r.UserID == 0 {
				return
			}
			if r.UserID != mustUserID(ctx, app, tgID) {
				return
			}
			next := time.Now().Add(time.Duration(r.IntervalMin) * time.Minute)
			_ = app.Store.SnoozeReminder(ctx, id, next)
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Хорошо, напомню чуть позже. Жду отчета!",
				ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
			})
			return
		}

		if strings.HasPrefix(data, "done_remind_") {
			id, err := strconv.ParseInt(strings.TrimPrefix(data, "done_remind_"), 10, 64)
			if err != nil {
				return
			}

			r, err := app.Store.GetReminderByID(ctx, id)
			if err != nil {
				return
			}
			userID := mustUserID(ctx, app, tgID)
			if userID == 0 || r.UserID != userID {
				return
			}

			_ = app.Store.DeactivateReminder(ctx, id)
			_, _ = b.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
				ChatID:      chatID,
				MessageID:   msgID,
				ReplyMarkup: keyboard.EmptyInlineKeyboard(),
			})

			app.State.Set(tgID, state.Pending{Kind: state.PendingDoneReport, ReminderID: &id})
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Отлично! Сколько сегодня сделала? Напиши текстом.",
				ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
			})
			return
		}
	}
}

func HandlePendingRemindInput(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingRemind {
			return
		}

		u, err := app.Store.GetUserByTgID(ctx, tgID)
		if err != nil {
			app.State.Clear(tgID)
			return
		}
		st, _ := app.Store.GetUserSettings(ctx, u.ID)

		loc := time.Local
		tm, msg, err := utils.ParseUserReminderInput(update.Message.Text, time.Now(), loc)
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
			IsRecurring: true,
			IntervalMin: st.ReminderIntervalMinutes,
		})
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла сохранить напоминание."})
			return
		}

		app.State.Clear(tgID)
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Ок, напомню в указанное время.",
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
	}
}

func mustUserID(ctx context.Context, app *botapp.App, tgID int64) int64 {
	u, err := app.Store.GetUserByTgID(ctx, tgID)
	if err != nil {
		u2, err2 := app.Store.EnsureUser(ctx, tgID, "")
		if err2 != nil {
			return 0
		}
		return u2.ID
	}
	return u.ID
}

