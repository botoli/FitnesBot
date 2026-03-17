package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
)

func MenuCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		// ack always
		_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

		if update.CallbackQuery.Message.Message == nil {
			return
		}

		switch update.CallbackQuery.Data {
		case "menu_plan":
			Plan(app)(ctx, b, &models.Update{Message: update.CallbackQuery.Message.Message})
		case "menu_done":
			Done(app)(ctx, b, &models.Update{Message: update.CallbackQuery.Message.Message})
		case "menu_stats":
			Stats(app)(ctx, b, &models.Update{Message: update.CallbackQuery.Message.Message})
		case "menu_remind":
			Remind(app)(ctx, b, &models.Update{Message: update.CallbackQuery.Message.Message})
		case "menu_settings":
			Settings(app)(ctx, b, &models.Update{Message: update.CallbackQuery.Message.Message})
		}
	}
}

