package handlers

import (
	"context"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
)

func Stats(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		chatID := update.Message.Chat.ID
		tgID := update.Message.From.ID
		username := update.Message.From.Username

		u, err := app.Store.EnsureUser(ctx, tgID, username)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось открыть профиль."})
			return
		}

		reports, err := app.Store.ListReports(ctx, u.ID, 10)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось загрузить историю."})
			return
		}

		if len(reports) == 0 {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Пока нет отчетов. Нажми “✅ Я позанималась” после тренировки.",
				ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
			})
			return
		}

		var sb strings.Builder
		sb.WriteString("📊 Последние отчеты:\n\n")
		for i, r := range reports {
			sb.WriteString(strconv.Itoa(i + 1))
			sb.WriteString(") ")
			sb.WriteString(r.CreatedAt.Format("02.01 15:04"))
			sb.WriteString("\n")
			sb.WriteString(r.ReportText)
			sb.WriteString("\n\n")
		}

		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        sb.String(),
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
	}
}

