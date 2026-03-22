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
	stmodels "traningBot/bot/internal/storage/models"
)

func settingsSummaryText(st stmodels.UserSettings) string {
	return "⚙️ Настройки\n\n" +
		"🔔 Частота напоминаний: каждые " + strconv.Itoa(st.ReminderIntervalMinutes) + " мин\n" +
		"🕒 Тихие часы: " + shortTime(st.QuietStart) + " – " + shortTime(st.QuietEnd) + "\n\n" +
		"Меняй кнопками ниже или текстом:\n" +
		"• freq 10 — интервал в минутах (1–120)\n" +
		"• quiet 23:00 08:00 — не беспокоить"
}

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
			ChatID:      chatID,
			Text:        settingsSummaryText(st),
			ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
		})
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "👇 Нажми, чтобы изменить:",
			ReplyMarkup: keyboard.SettingsInlineKeyboard(),
		})
	}
}

func SettingsCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

		if update.CallbackQuery.Message.Message == nil {
			return
		}
		msg := update.CallbackQuery.Message.Message
		chatID := msg.Chat.ID
		msgID := msg.ID
		tgID := update.CallbackQuery.From.ID
		username := update.CallbackQuery.From.Username
		data := update.CallbackQuery.Data

		u, err := app.Store.EnsureUser(ctx, tgID, username)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось открыть профиль."})
			return
		}

		switch data {
		case "set_freq_5":
			_, _ = app.Store.UpdateReminderInterval(ctx, u.ID, 5)
		case "set_freq_10":
			_, _ = app.Store.UpdateReminderInterval(ctx, u.ID, 10)
		case "set_freq_15":
			_, _ = app.Store.UpdateReminderInterval(ctx, u.ID, 15)
		case "set_quiet_23_08":
			_, _ = app.Store.UpdateQuietHours(ctx, u.ID, "23:00", "08:00")
		default:
			return
		}

		st, err := app.Store.GetUserSettings(ctx, u.ID)
		if err != nil {
			return
		}

		text := settingsSummaryText(st)
		if _, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   msgID,
			Text:        "✅ Сохранено.\n\n" + text,
			ReplyMarkup: keyboard.SettingsInlineKeyboard(),
		}); err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        "✅ Сохранено.\n\n" + text,
				ReplyMarkup: keyboard.SettingsInlineKeyboard(),
			})
		}
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
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Формат: freq 10"})
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
			st, _ := app.Store.GetUserSettings(ctx, u.ID)
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Готово.\n\n" + settingsSummaryText(st), ReplyMarkup: keyboard.MainMenuReplyKeyboard()})
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "👇", ReplyMarkup: keyboard.SettingsInlineKeyboard()})
			return

		case "quiet":
			if len(parts) != 3 {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Формат: quiet 23:00 08:00"})
				return
			}
			_, err = app.Store.UpdateQuietHours(ctx, u.ID, parts[1], parts[2])
			if err != nil {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла обновить тихие часы."})
				return
			}
			app.State.Clear(tgID)
			st, _ := app.Store.GetUserSettings(ctx, u.ID)
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Готово.\n\n" + settingsSummaryText(st), ReplyMarkup: keyboard.MainMenuReplyKeyboard()})
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "👇", ReplyMarkup: keyboard.SettingsInlineKeyboard()})
			return
		default:
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не поняла. Примеры: freq 10, quiet 23:00 08:00"})
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
