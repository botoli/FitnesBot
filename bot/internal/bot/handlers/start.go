package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/copy"
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
		name := ""
		if update.Message.From.FirstName != "" {
			name = update.Message.From.FirstName
		}
		intro := copy.Welcome(name) + "\n\n" + copy.RandomTip()
		SendHomeMessages(ctx, b, chatID, intro)
	}
}
