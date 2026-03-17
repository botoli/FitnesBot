package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/bot/state"
	stmodels "traningBot/bot/internal/storage/models"
)

func Done(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
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

		app.State.Set(tgID, state.Pending{Kind: state.PendingDoneReport, ReminderID: nil})
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Отлично! Напиши отчет текстом (что и сколько сделала).",
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
	}
}

func HandlePendingDoneReport(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingDoneReport {
			return
		}

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
			ReportText: update.Message.Text,
		})
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла записать отчет."})
			return
		}

		app.State.Clear(tgID)
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Записала! Так держать.",
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
	}
}

