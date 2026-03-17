package handlers

import (
	"context"
	"fmt"
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
				ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
			})
			return
		}

		if len(plans) == 0 {
			seedDefaultPlan(ctx, app, u.ID)
			plans, _ = app.Store.ListPlansForWeek(ctx, u.ID)
		}

		var sb strings.Builder
	sb.WriteString("Твой план на неделю:\n")
		byDay := map[int]string{}
		for _, p := range plans {
		byDay[p.DayOfWeek] = fmt.Sprintf("%s: %s\n%s", weekdayRu(p.DayOfWeek), p.Title, p.Details)
		}
		for d := 1; d <= 7; d++ {
			if v, ok := byDay[d]; ok {
				sb.WriteString(v)
				sb.WriteString("\n\n")
			}
		}

		today := dayOfWeekISO(time.Now())
		if v, ok := byDay[today]; ok {
		sb.WriteString("Сегодня:\n")
			sb.WriteString(v)
		}

		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        sb.String(),
			ReplyMarkup: keyboard.MainMenuInlineKeyboard(),
		})
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

func seedDefaultPlan(ctx context.Context, app *botapp.App, userID int64) {
	// quiet best-effort defaults
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 1, Title: "Пресс и Кор", Details: "Планка 1 минута | Скручивания 20 раз"})
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 2, Title: "Спина", Details: "Гиперэкстензия 3×12"})
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 3, Title: "Кардио", Details: "20 минут быстрым шагом"})
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 4, Title: "Ноги", Details: "Приседания 3×12 | Выпады 3×10"})
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 5, Title: "Растяжка", Details: "10 минут"})
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 6, Title: "Отдых", Details: "Легкая прогулка"})
	_, _ = app.Store.UpsertPlan(ctx, smodels.Plan{UserID: userID, DayOfWeek: 7, Title: "Отдых", Details: "Сон и восстановление"})
}


