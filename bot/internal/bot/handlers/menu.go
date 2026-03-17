package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
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

		// Important: callback is triggered by USER, but the message "From" is the BOT.
		// Handlers expect Message.From to be the user, so we override it here.
		msg := *update.CallbackQuery.Message.Message
		msg.From = &update.CallbackQuery.From

		switch update.CallbackQuery.Data {
		case "menu_home":
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      msg.Chat.ID,
				Text:        "Выбирай действие:",
				ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
			})
		case "menu_plan":
			Plan(app)(ctx, b, &models.Update{Message: &msg})
		case "menu_done":
			Done(app)(ctx, b, &models.Update{Message: &msg})
		case "menu_stats":
			Stats(app)(ctx, b, &models.Update{Message: &msg})
		case "menu_remind":
			Remind(app)(ctx, b, &models.Update{Message: &msg})
		case "menu_settings":
			Settings(app)(ctx, b, &models.Update{Message: &msg})
		}
	}
}

