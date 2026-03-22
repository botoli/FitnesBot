package handlers

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/bot/state"
	stmodels "traningBot/bot/internal/storage/models"
)

var numberRe = regexp.MustCompile(`[-+]?\d+(?:[\.,]\d+)?`)

func Done(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		chatID := update.Message.Chat.ID
		tgID := update.Message.From.ID
		username := update.Message.From.Username

		_, err := app.Store.EnsureUser(ctx, tgID, username)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось открыть профиль."})
			return
		}

		plans, err := app.Store.ListPlansForWeek(ctx, mustUserID(ctx, app, tgID))
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла загрузить план тренировки."})
			return
		}

		todayPlans := plansForToday(plans)
		selectionText := "🏋️ Главное › Я позанималась\n\n🏋️ Что сегодня делала?\n\nВыбери тренировку из плана:"
		if len(todayPlans) == 0 {
			selectionText = "🏋️ Главное › Я позанималась\n\n🏋️ Что сегодня делала?\n\nНа сегодня плана нет, выбери «Своя тренировка ✍️»."
		}
		sent, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        selectionText,
			ReplyMarkup: donePickInlineKeyboard(todayPlans),
		})
		if err != nil {
			return
		}

		app.State.Set(tgID, state.Pending{
			Kind: state.PendingDoneFlow,
			DoneFlow: &state.DoneFlowSession{
				PromptMessageID: sent.ID,
			},
		})
	}
}

func HandleDoneFlowCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
			return
		}

		data := update.CallbackQuery.Data
		if !strings.HasPrefix(data, "doneflow_") {
			return
		}
		_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

		msg := update.CallbackQuery.Message.Message
		chatID := msg.Chat.ID
		msgID := msg.ID
		tgID := update.CallbackQuery.From.ID

		if data == "doneflow_custom" {
			app.State.Set(tgID, state.Pending{Kind: state.PendingDoneReport})
			if _, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: msgID,
				Text:      "🏋️ Главное › Я позанималась › Своя тренировка\n\nОпиши тренировку текстом: что делала и сколько.",
			}); err != nil {
				_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
					ChatID: chatID,
					Text:   "🏋️ Главное › Я позанималась › Своя тренировка\n\nОпиши тренировку текстом: что делала и сколько.",
				})
			}
			return
		}

		if data == "doneflow_stats" {
			statsMsg := *msg
			statsMsg.From = &update.CallbackQuery.From
			Stats(app)(ctx, b, &models.Update{Message: &statsMsg})
			return
		}

		if !strings.HasPrefix(data, "doneflow_plan_") {
			return
		}

		planID, err := strconv.ParseInt(strings.TrimPrefix(data, "doneflow_plan_"), 10, 64)
		if err != nil {
			return
		}

		u, err := app.Store.EnsureUser(ctx, tgID, update.CallbackQuery.From.Username)
		if err != nil {
			return
		}

		plans, err := app.Store.ListPlansForWeek(ctx, u.ID)
		if err != nil {
			return
		}

		var selected *stmodels.Plan
		for i := range plans {
			if plans[i].ID == planID {
				selected = &plans[i]
				break
			}
		}
		if selected == nil {
			return
		}

		exercises := parsePlanExercises(selected.Details)
		if len(exercises) == 0 {
			exercises = []state.DoneExercise{{Name: selected.Title, Plan: "в свободной форме"}}
		}

		session := &state.DoneFlowSession{
			WorkoutTitle:    selected.Title,
			Exercises:       exercises,
			Answers:         make([]state.DoneAnswer, 0, len(exercises)),
			CurrentIndex:    0,
			PromptMessageID: msgID,
		}
		if _, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: msgID,
			Text:      doneExercisePrompt(session),
		}); err != nil {
			sent, sendErr := b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID: chatID,
				Text:   doneExercisePrompt(session),
			})
			if sendErr == nil {
				session.PromptMessageID = sent.ID
			}
		}
		app.State.Set(tgID, state.Pending{Kind: state.PendingDoneFlow, DoneFlow: session})
	}
}

func HandlePendingDoneReport(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingDoneReport {
			return
		}

		u, err := app.Store.GetUserByTgID(ctx, tgID)
		if err != nil {
			app.State.Clear(tgID)
			return
		}

		var reminderID *int64
		if p.ReminderID != nil {
			reminderID = p.ReminderID
		}

		_, err = app.Store.CreateReport(ctx, stmodels.Report{
			UserID:     u.ID,
			ReminderID: reminderID,
			ReportText: update.Message.Text,
		})
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла записать отчет."})
			return
		}

		app.State.Clear(tgID)
		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Записала! Так держать.",
			ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
		})
	}
}

func HandlePendingDoneFlowInput(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		value := strings.TrimSpace(update.Message.Text)

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingDoneFlow || p.DoneFlow == nil {
			return
		}
		if len(p.DoneFlow.Exercises) == 0 || p.DoneFlow.CurrentIndex >= len(p.DoneFlow.Exercises) {
			return
		}
		if _, ok := parseFloatValue(value); !ok {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Введи число, например: 15 или 1.5"})
			return
		}

		ex := p.DoneFlow.Exercises[p.DoneFlow.CurrentIndex]
		p.DoneFlow.Answers = append(p.DoneFlow.Answers, state.DoneAnswer{Name: ex.Name, Plan: ex.Plan, Actual: value})

		if p.DoneFlow.CurrentIndex+1 < len(p.DoneFlow.Exercises) {
			p.DoneFlow.CurrentIndex++
			if _, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: p.DoneFlow.PromptMessageID,
				Text:      doneExercisePrompt(p.DoneFlow),
			}); err != nil {
				sent, sendErr := b.SendMessage(ctx, &tgbot.SendMessageParams{
					ChatID: chatID,
					Text:   doneExercisePrompt(p.DoneFlow),
				})
				if sendErr == nil {
					p.DoneFlow.PromptMessageID = sent.ID
				}
			}
			app.State.Set(tgID, p)
			return
		}

		u, err := app.Store.GetUserByTgID(ctx, tgID)
		if err != nil {
			app.State.Clear(tgID)
			return
		}

		reportText := reportTextFromDoneFlow(p.DoneFlow)
		_, err = app.Store.CreateReport(ctx, stmodels.Report{
			UserID:     u.ID,
			ReminderID: p.DoneFlow.SourceReminder,
			ReportText: reportText,
		})
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла записать отчет."})
			return
		}

		finalText := doneFinalSummary(p.DoneFlow)
		app.State.Clear(tgID)
		if _, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   p.DoneFlow.PromptMessageID,
			Text:        finalText,
			ReplyMarkup: keyboard.DoneFinalInlineKeyboard(),
		}); err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        finalText,
				ReplyMarkup: keyboard.DoneFinalInlineKeyboard(),
			})
		}
	}
}

func plansForToday(plans []stmodels.Plan) []stmodels.Plan {
	today := dayOfWeekISO(time.Now())
	out := make([]stmodels.Plan, 0)
	for _, p := range plans {
		if p.DayOfWeek == today {
			out = append(out, p)
		}
	}
	return out
}

func donePickInlineKeyboard(plans []stmodels.Plan) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(plans)+1)
	for _, p := range plans {
		rows = append(rows, []models.InlineKeyboardButton{{Text: p.Title, CallbackData: "doneflow_plan_" + strconv.FormatInt(p.ID, 10)}})
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Своя тренировка ✍️", CallbackData: "doneflow_custom"}})
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func parsePlanExercises(details string) []state.DoneExercise {
	parts := strings.FieldsFunc(details, func(r rune) bool {
		return r == '|' || r == '\n' || r == ';'
	})
	out := make([]state.DoneExercise, 0, len(parts))
	for _, p := range parts {
		item := strings.TrimSpace(p)
		if item == "" {
			continue
		}
		name, plan := splitExercisePlan(item)
		out = append(out, state.DoneExercise{Name: name, Plan: plan})
	}
	return out
}

func splitExercisePlan(item string) (string, string) {
	idx := numberRe.FindStringIndex(item)
	if idx == nil {
		return item, "по ощущениям"
	}
	name := strings.TrimSpace(strings.Trim(item[:idx[0]], "-: "))
	plan := strings.TrimSpace(item[idx[0]:])
	if name == "" {
		name = item
	}
	if plan == "" {
		plan = "по ощущениям"
	}
	return name, plan
}

func doneExercisePrompt(flow *state.DoneFlowSession) string {
	total := len(flow.Exercises)
	step := flow.CurrentIndex + 1
	if step < 1 {
		step = 1
	}
	if step > total {
		step = total
	}
	bar := progressBar(step, total)
	ex := flow.Exercises[flow.CurrentIndex]

	return "🏋️ Главное › Я позанималась › " + flow.WorkoutTitle + "\n\n" +
		bar + " Упражнение " + strconv.Itoa(step) + " из " + strconv.Itoa(total) + "\n\n" +
		ex.Name + "\n" +
		"📋 План: " + ex.Plan + "\n\n" +
		"Сколько сделала? (напиши число)"
}

func doneFinalSummary(flow *state.DoneFlowSession) string {
	total := len(flow.Answers)
	var sb strings.Builder
	sb.WriteString("✨ Тренировка завершена! ✨\n\n")
	sb.WriteString(progressBar(total, total))
	sb.WriteString(" ")
	sb.WriteString(strconv.Itoa(total))
	sb.WriteString(" из ")
	sb.WriteString(strconv.Itoa(total))
	sb.WriteString(" упражнений\n\n")
	sb.WriteString("📊 Твои результаты:\n")
	for _, a := range flow.Answers {
		sb.WriteString("• ")
		sb.WriteString(a.Name)
		sb.WriteString(": ")
		sb.WriteString(a.Actual)
		if a.Plan != "" {
			sb.WriteString(" (план: ")
			sb.WriteString(a.Plan)
			sb.WriteString(")")
		}
		if d, ok := deltaToPlan(a.Actual, a.Plan); ok {
			sb.WriteString(" ")
			sb.WriteString(d)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n💪 Ты умница! Результаты записаны в историю")
	return sb.String()
}

func reportTextFromDoneFlow(flow *state.DoneFlowSession) string {
	var sb strings.Builder
	sb.WriteString("Тренировка: ")
	sb.WriteString(flow.WorkoutTitle)
	sb.WriteString("\n")
	for _, a := range flow.Answers {
		sb.WriteString("- ")
		sb.WriteString(a.Name)
		sb.WriteString(": ")
		sb.WriteString(a.Actual)
		sb.WriteString(" (план: ")
		sb.WriteString(a.Plan)
		sb.WriteString(")\n")
	}
	return strings.TrimSpace(sb.String())
}

func progressBar(done int, total int) string {
	if total <= 0 {
		return ""
	}
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}
	var sb strings.Builder
	for i := 0; i < total; i++ {
		if i < done {
			sb.WriteString("🟩")
		} else {
			sb.WriteString("⬜️")
		}
	}
	return sb.String()
}

func deltaToPlan(actual string, plan string) (string, bool) {
	a, okA := parseFloatValue(actual)
	p, okP := parseFloatValue(plan)
	if !okA || !okP {
		return "", false
	}
	delta := a - p
	if math.Abs(delta) < 0.001 {
		return "(👍 норма)", true
	}
	if delta > 0 {
		return fmt.Sprintf("(+%s к плану) 🔥", shortFloat(delta)), true
	}
	return fmt.Sprintf("(%s до плана)", shortFloat(delta)), true
}

func parseFloatValue(text string) (float64, bool) {
	m := numberRe.FindString(strings.ReplaceAll(text, ",", "."))
	if m == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func shortFloat(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
