package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
)

// Cancel сбрасывает активный сценарий и показывает главный экран.
func Cancel(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		app.State.Clear(update.Message.From.ID)
		SendHomeMessages(ctx, b, update.Message.Chat.ID, "Окей, сбросила то, что делали.\n\n")
	}
}
