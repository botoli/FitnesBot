package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/copy"
	"traningBot/bot/internal/bot/keyboard"
	smodels "traningBot/bot/internal/storage/models"
)

func Plan(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *tgmodels.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *tgmodels.Update) {
		if update.Message == nil {
			return
		}
		tgID := update.Message.From.ID
		username := update.Message.From.Username

		u, err := app.Store.EnsureUser(ctx, tgID, username)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Не смогла открыть профиль. Попробуй позже.",
			})
			return
		}

		plans, err := app.Store.ListPlansForWeek(ctx, u.ID)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не смогла прочитать план из базы.",
				ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
			})
			return
		}

		if len(plans) == 0 {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Пока нет плана. Нажми «✏️ Изменить план», чтобы создать первую тренировку.",
				ReplyMarkup: keyboard.PlanViewInlineKeyboard(),
			})
			return
		}

		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: "📋 План тренировок\n\n" +
				"Выбери день недели или посмотри весь план на неделю — кнопки ниже.",
			ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
		})
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "👇",
			ReplyMarkup: keyboard.PlanDayPickerInlineKeyboard(),
		})
	}
}

// HandlePlanViewCallbacks — выбор дня / вся неделя / назад в просмотре плана.
func HandlePlanViewCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *tgmodels.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *tgmodels.Update) {
		if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
			return
		}
		data := update.CallbackQuery.Data
		if !strings.HasPrefix(data, "planview_") {
			return
		}
		_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

		msg := update.CallbackQuery.Message.Message
		chatID := msg.Chat.ID
		msgID := msg.ID
		tgID := update.CallbackQuery.From.ID

		u, err := app.Store.EnsureUser(ctx, tgID, update.CallbackQuery.From.Username)
		if err != nil {
			return
		}
		plans, err := app.Store.ListPlansForWeek(ctx, u.ID)
		if err != nil {
			return
		}

		today := dayOfWeekISO(time.Now())

		switch data {
		case "planview_back":
			pickerText := "📋 План тренировок\n\nВыбери день недели или «Вся неделя»:"
			_, _ = b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:      chatID,
				MessageID:   msgID,
				Text:        pickerText,
				ReplyMarkup: keyboard.PlanDayPickerInlineKeyboard(),
			})
			return

		case "planview_week":
			if len(plans) == 0 {
				_, _ = b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
					ChatID:      chatID,
					MessageID:   msgID,
					Text:        "Плана пока нет.",
					ReplyMarkup: keyboard.PlanDayPickerInlineKeyboard(),
				})
				return
			}
			_, _ = b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:      chatID,
				MessageID:   msgID,
				Text:        "📅 Отправляю план на всю неделю — смотри сообщения ниже.",
				ReplyMarkup: keyboard.PlanDayPickerInlineKeyboard(),
			})
			weekText := renderWeeklyPlanMessage(plans, today)
			sendTextInChunks(ctx, b, chatID, weekText, keyboard.PlanViewInlineKeyboard())
			return
		}

		if strings.HasPrefix(data, "planview_day_") {
			day, err := strconv.Atoi(strings.TrimPrefix(data, "planview_day_"))
			if err != nil || day < 1 || day > 7 {
				return
			}
			if len(plans) == 0 {
				_, _ = b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
					ChatID:      chatID,
					MessageID:   msgID,
					Text:        "Плана пока нет.",
					ReplyMarkup: keyboard.PlanDayPickerInlineKeyboard(),
				})
				return
			}
			text := renderSingleDayPlanMessage(plans, day, today)
			if _, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:      chatID,
				MessageID:   msgID,
				Text:        text,
				ReplyMarkup: keyboard.PlanDayNavInlineKeyboard(),
			}); err != nil {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: text, ReplyMarkup: keyboard.PlanDayNavInlineKeyboard()})
			}
		}
	}
}

func renderSingleDayPlanMessage(plans []smodels.Plan, day int, today int) string {
	var p *smodels.Plan
	for i := range plans {
		if plans[i].DayOfWeek == day {
			p = &plans[i]
			break
		}
	}

	var sb strings.Builder
	sb.WriteString("📋 План на ")
	sb.WriteString(weekdayRuFullUpper(day))
	if day == today {
		sb.WriteString(" ⬅️ СЕГОДНЯ")
	}
	sb.WriteString("\n\n")

	if p == nil {
		sb.WriteString("В этот день тренировка в плане не задана.\nНажми «➕ Добавить тренировку» внизу.")
		return sb.String()
	}

	sb.WriteString(renderPlanDetailsSection(p.Title, p.Details))
	sb.WriteString("\n\n")
	sb.WriteString("📌 Условные обозначения:\n")
	sb.WriteString("🟥 День А — Грудь/Плечи/Руки\n")
	sb.WriteString("🟦 День Б — Спина/Ноги/Пресс\n")
	sb.WriteString("▸ Упражнение / план")
	return sb.String()
}

func renderWeeklyPlanMessage(plans []smodels.Plan, today int) string {
	sep := copy.Divider
	byDay := make(map[int]smodels.Plan, len(plans))
	for _, p := range plans {
		byDay[p.DayOfWeek] = p
	}

	var sb strings.Builder
	sb.WriteString("📋 МОЙ ПЛАН ТРЕНИРОВОК\n")

	for day := 1; day <= 7; day++ {
		plan, ok := byDay[day]
		if !ok {
			continue
		}

		sb.WriteString(sep)
		sb.WriteString("\n")
		sb.WriteString("📅 ")
		sb.WriteString(weekdayRuFullUpper(day))
		sb.WriteString(" — ")
		sb.WriteString(plan.Title)
		if day == today {
			sb.WriteString(" ⬅️ СЕГОДНЯ")
		}
		sb.WriteString("\n")
		sb.WriteString(sep)
		sb.WriteString("\n")
		sb.WriteString(renderPlanDetailsSection(plan.Title, plan.Details))
		sb.WriteString("\n\n")
	}

	sb.WriteString(sep)
	sb.WriteString("\n")
	sb.WriteString("📌 Условные обозначения:\n")
	sb.WriteString("🟥 День А — Грудь/Плечи/Руки\n")
	sb.WriteString("🟦 День Б — Спина/Ноги/Пресс\n")
	sb.WriteString("⏱️ Отдых между упражнениями\n")
	sb.WriteString("▸ Подход/упражнение")

	return sb.String()
}

func renderPlanDetailsSection(title string, details string) string {
	exercises := parsePlanExercises(details)
	if len(exercises) == 0 {
		return "▸ " + strings.TrimSpace(details)
	}

	blockEmoji := "🟦"
	lowerTitle := strings.ToLower(title)
	if strings.Contains(lowerTitle, "день а") {
		blockEmoji = "🟥"
	}

	var sb strings.Builder
	sb.WriteString(blockEmoji)
	sb.WriteString(" ")
	sb.WriteString(title)
	sb.WriteString("\n")

	for i, ex := range exercises {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("▸ ")
		sb.WriteString(ex.Name)
		sb.WriteString("\n")
		sb.WriteString("— ")
		sb.WriteString(ex.Plan)
	}

	return sb.String()
}

func sendTextInChunks(ctx context.Context, b *tgbot.Bot, chatID int64, text string, lastReplyMarkup tgmodels.ReplyMarkup) {
	const maxLen = 3900 // below Telegram 4096 hard limit, keeping margin for multibyte runes
	runes := []rune(text)
	if len(runes) == 0 {
		return
	}

	for start := 0; start < len(runes); start += maxLen {
		end := start + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		params := &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   string(runes[start:end]),
		}
		if end == len(runes) {
			params.ReplyMarkup = lastReplyMarkup
		}
		_, _ = b.SendMessage(ctx, params)
	}
}

func dayOfWeekISO(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func weekdayRuFullUpper(d int) string {
	switch d {
	case 1:
		return "ПОНЕДЕЛЬНИК"
	case 2:
		return "ВТОРНИК"
	case 3:
		return "СРЕДА"
	case 4:
		return "ЧЕТВЕРГ"
	case 5:
		return "ПЯТНИЦА"
	case 6:
		return "СУББОТА"
	case 7:
		return "ВОСКРЕСЕНЬЕ"
	default:
		return "?"
	}
}

