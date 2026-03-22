package handlers

import (
	"context"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
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
				Text:        "Пока нет плана. Нажми «✏️ Редактировать план», чтобы создать первую тренировку.",
				ReplyMarkup: keyboard.PlanViewInlineKeyboard(),
			})
			return
		}

		today := dayOfWeekISO(time.Now())
		text := renderWeeklyPlanMessage(plans, today)
		sendTextInChunks(ctx, b, update.Message.Chat.ID, text, keyboard.PlanViewInlineKeyboard())
	}
}

func renderWeeklyPlanMessage(plans []smodels.Plan, today int) string {
	const sep = "━━━━━━━━━━━━━━━━━━━━━━━━━━━"
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

func weekdayRu(d int) string {
	switch d {
	case 1:
		return "Пн"
	case 2:
		return "Вт"
	case 3:
		return "Ср"
	case 4:
		return "Чт"
	case 5:
		return "Пт"
	case 6:
		return "Сб"
	case 7:
		return "Вс"
	default:
		return "?"
	}
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

