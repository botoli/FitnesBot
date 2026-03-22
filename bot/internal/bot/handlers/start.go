package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
)

func Start(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		chatID := update.Message.Chat.ID
		tgID := update.Message.From.ID
		username := update.Message.From.Username

		_, _ = app.Store.EnsureUser(ctx, tgID, username)
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Привет! Я твой тренировочный бот.\nОсновные действия внизу на обычной клавиатуре.",
			ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
		})
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Быстрые кнопки:",
			ReplyMarkup: keyboard.QuickActionsInlineKeyboard(),
		})
	}
}
