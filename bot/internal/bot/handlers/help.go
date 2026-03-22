package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/copy"
	"traningBot/bot/internal/bot/keyboard"
)

// Help показывает краткую справку.
func Help(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		_ = app
		if update.Message == nil {
			return
		}
		chatID := update.Message.Chat.ID
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        copy.Help(),
			ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
		})
	}
}
