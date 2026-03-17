package handlers

import (
	"context"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/bot/state"
)

func Settings(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
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
		st, err := app.Store.GetUserSettings(ctx, u.ID)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось загрузить настройки."})
			return
		}

		app.State.Set(tgID, state.Pending{Kind: state.PendingSettings})
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text: "**Настройки напоминаний:**\n" +
				"🔔 Частота: каждые " + strconv.Itoa(st.ReminderIntervalMinutes) + " минут\n" +
				"🕒 Тихие часы: " + shortTime(st.QuietStart) + " - " + shortTime(st.QuietEnd) + "\n\n" +
				"Чтобы поменять:\n" +
				"- `freq 10`\n" +
				"- `quiet 23:00 08:00`",
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: keyboard.MainMenuKeyboard(),
		})
	}
}

func HandlePendingSettings(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingSettings {
			return
		}

		u, err := app.Store.GetUserByTgID(ctx, tgID)
		if err != nil {
			app.State.Clear(tgID)
			return
		}

		txt := strings.TrimSpace(update.Message.Text)
		parts := strings.Fields(txt)
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
			_, err = app.Store.UpdateReminderInterval(ctx, u.ID, mins)
			if err != nil {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла обновить частоту."})
				return
			}
			app.State.Clear(tgID)
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Готово. Частота обновлена.", ReplyMarkup: keyboard.MainMenuKeyboard()})
			return

		case "quiet":
			if len(parts) != 3 {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Формат: `quiet 23:00 08:00`", ParseMode: models.ParseModeMarkdown})
				return
			}
			_, err = app.Store.UpdateQuietHours(ctx, u.ID, parts[1], parts[2])
			if err != nil {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла обновить тихие часы."})
				return
			}
			app.State.Clear(tgID)
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Готово. Тихие часы обновлены.", ReplyMarkup: keyboard.MainMenuKeyboard()})
			return
		default:
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Команда не распознана. Примеры: `freq 10`, `quiet 23:00 08:00`", ParseMode: models.ParseModeMarkdown})
		}
	}
}

func shortTime(dbTime string) string {
	// dbTime like "23:00:00"
	if len(dbTime) >= 5 {
		return dbTime[:5]
	}
	return dbTime
}

