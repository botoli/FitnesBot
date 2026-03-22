package handlers

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/bot/state"
	stmodels "traningBot/bot/internal/storage/models"
)

var planInputNumberRe = regexp.MustCompile(`[-+]?\d+(?:[\.,]\d+)?`)

func AddPlan(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
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

		draft := &state.PlanDraftSession{Step: state.PlanDraftSelectDays, Days: []int{}, Exercises: []state.PlanDraftExercise{}}
		sent, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "🗓 Выбери дни для тренировки\n\nМожно выбрать несколько дней:",
			ReplyMarkup: planAddDaysInlineKeyboard(draft.Days),
		})
		if err != nil {
			return
		}
		draft.PromptMsg = sent.ID
		app.State.Set(tgID, state.Pending{Kind: state.PendingPlanAdd, PlanDraft: draft})
	}
}

func HandlePendingPlanAdd(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingPlanAdd || p.PlanDraft == nil {
			return
		}
		draft := p.PlanDraft

		switch draft.Step {
		case state.PlanDraftSelectDays:
			_ = editPlanDraftMessage(ctx, b, chatID, draft.PromptMsg, "🗓 Выбери дни для тренировки\n\nМожно выбрать несколько дней:", planAddDaysInlineKeyboard(draft.Days))
			return

		case state.PlanDraftTitle:
			draft.Title = text
			draft.Step = state.PlanDraftExercises
			app.State.Set(tgID, p)
			_ = editPlanDraftMessage(ctx, b, chatID, draft.PromptMsg, "✏️ Тренировка \""+draft.Title+"\"\n\nОпиши первое упражнение и его план:\n\nПример: Планка 1 минута\nили: Скручивания 20 раз", planAddExerciseActionsKeyboard())
			return

		case state.PlanDraftExercises:
			name, plan := splitExerciseInput(text)
			draft.Exercises = append(draft.Exercises, state.PlanDraftExercise{Name: name, Plan: plan})
			app.State.Set(tgID, p)
			_ = editPlanDraftMessage(ctx, b, chatID, draft.PromptMsg, draftExercisesPromptText(draft), planAddExerciseActionsKeyboard())
			return
		}
	}
}

func HandlePlanAddCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
			return
		}
		data := update.CallbackQuery.Data
		if !strings.HasPrefix(data, "planadd_") {
			return
		}
		_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})

		msg := update.CallbackQuery.Message.Message
		chatID := msg.Chat.ID
		msgID := msg.ID
		tgID := update.CallbackQuery.From.ID

		p, ok := app.State.Get(tgID)
		if !ok || p.Kind != state.PendingPlanAdd || p.PlanDraft == nil {
			return
		}

		draft := p.PlanDraft

		if strings.HasPrefix(data, "planadd_day_") {
			if draft.Step != state.PlanDraftSelectDays {
				return
			}
			day, err := strconv.Atoi(strings.TrimPrefix(data, "planadd_day_"))
			if err != nil || day < 1 || day > 7 {
				return
			}
			draft.Days = toggleDay(draft.Days, day)
			app.State.Set(tgID, p)
			targetMsg := draft.PromptMsg
			if targetMsg == 0 {
				targetMsg = msgID
				draft.PromptMsg = msgID
				app.State.Set(tgID, p)
			}
			_ = editPlanDraftMessage(ctx, b, chatID, targetMsg, "🗓 Выбери дни для тренировки\n\nМожно выбрать несколько дней:", planAddDaysInlineKeyboard(draft.Days))
			return
		}

		switch data {
		case "planadd_done_days":
			if draft.Step != state.PlanDraftSelectDays {
				return
			}
			if len(draft.Days) == 0 {
				_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "Выбери хотя бы один день",
					ShowAlert:       true,
				})
				return
			}
			draft.Step = state.PlanDraftTitle
			app.State.Set(tgID, p)
			targetMsg := draft.PromptMsg
			if targetMsg == 0 {
				targetMsg = msgID
				draft.PromptMsg = msgID
				app.State.Set(tgID, p)
			}
			_ = editPlanDraftMessage(ctx, b, chatID, targetMsg, "📝 Отлично! Тренировка будет в:\n"+fullDaysText(draft.Days)+"\n\nНапиши название тренировки:\n(например: \"Пресс и кор\" или \"Утренняя зарядка\")", nil)
			return

		case "planadd_add_more":
			if draft.Step != state.PlanDraftExercises {
				return
			}
			if len(draft.Exercises) == 0 {
				_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "Сначала введи первое упражнение",
					ShowAlert:       true,
				})
				return
			}
			targetMsg := draft.PromptMsg
			if targetMsg == 0 {
				targetMsg = msgID
				draft.PromptMsg = msgID
				app.State.Set(tgID, p)
			}
			_ = editPlanDraftMessage(ctx, b, chatID, targetMsg, draftExercisesPromptText(draft), planAddExerciseActionsKeyboard())
			return

		case "planadd_save":
			if draft.Step != state.PlanDraftExercises {
				return
			}
			if len(draft.Exercises) == 0 {
				_, _ = b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "Добавь хотя бы одно упражнение",
					ShowAlert:       true,
				})
				return
			}

			u, err := app.Store.GetUserByTgID(ctx, tgID)
			if err != nil {
				app.State.Clear(tgID)
				return
			}
			details := formatPlanDetails(draft.Exercises)
			for _, d := range draft.Days {
				_, err := app.Store.UpsertPlan(ctx, stmodels.Plan{UserID: u.ID, DayOfWeek: d, Title: draft.Title, Details: details})
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла сохранить тренировку."})
					return
				}
			}

			finalText := "🎉 Тренировка сохранена!\n\n\"" + draft.Title + "\"\n" +
				"📅 Дни: " + shortDaysText(draft.Days) + "\n" +
				"📝 " + strconv.Itoa(len(draft.Exercises)) + " упражнения:\n" +
				exerciseSummaryText(draft.Exercises) +
				"\n\nТеперь я буду напоминать о ней в эти дни! ✨"

			targetMsg := draft.PromptMsg
			if targetMsg == 0 {
				targetMsg = msgID
			}
			app.State.Clear(tgID)
			_ = editPlanDraftMessage(ctx, b, chatID, targetMsg, finalText, keyboard.PlanSavedInlineKeyboard())
			return
		}
	}
}

func editPlanDraftMessage(ctx context.Context, b *tgbot.Bot, chatID int64, messageID int, text string, markup models.ReplyMarkup) error {
	if messageID == 0 {
		return nil
	}
	_, err := b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ReplyMarkup: markup,
	})
	return err
}

func planAddDaysInlineKeyboard(selected []int) *models.InlineKeyboardMarkup {
	row := make([]models.InlineKeyboardButton, 0, 7)
	for day := 1; day <= 7; day++ {
		label := shortWeekday(day)
		if containsDay(selected, day) {
			label = "✅ " + label
		}
		row = append(row, models.InlineKeyboardButton{Text: label, CallbackData: "planadd_day_" + strconv.Itoa(day)})
	}
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			row,
			{{Text: "✅ Готово", CallbackData: "planadd_done_days"}},
		},
	}
}

func planAddExerciseActionsKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "➕ Добавить еще", CallbackData: "planadd_add_more"}, {Text: "✅ Сохранить тренировку", CallbackData: "planadd_save"}},
		},
	}
}

func toggleDay(days []int, day int) []int {
	if containsDay(days, day) {
		out := make([]int, 0, len(days)-1)
		for _, d := range days {
			if d != day {
				out = append(out, d)
			}
		}
		return out
	}
	out := append(days, day)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func containsDay(days []int, day int) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func shortWeekday(day int) string {
	switch day {
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

func fullWeekday(day int) string {
	switch day {
	case 1:
		return "Понедельник"
	case 2:
		return "Вторник"
	case 3:
		return "Среда"
	case 4:
		return "Четверг"
	case 5:
		return "Пятница"
	case 6:
		return "Суббота"
	case 7:
		return "Воскресенье"
	default:
		return "?"
	}
}

func fullDaysText(days []int) string {
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, fullWeekday(d))
	}
	return strings.Join(parts, " • ")
}

func shortDaysText(days []int) string {
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, shortWeekday(d))
	}
	return strings.Join(parts, ", ")
}

func splitExerciseInput(input string) (string, string) {
	item := strings.TrimSpace(input)
	idx := planInputNumberRe.FindStringIndex(item)
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

func formatPlanDetails(exercises []state.PlanDraftExercise) string {
	parts := make([]string, 0, len(exercises))
	for _, ex := range exercises {
		parts = append(parts, ex.Name+" "+ex.Plan)
	}
	return strings.Join(parts, " | ")
}

func exerciseSummaryText(exercises []state.PlanDraftExercise) string {
	var sb strings.Builder
	for i, ex := range exercises {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("• ")
		sb.WriteString(ex.Name)
		sb.WriteString(" (")
		sb.WriteString(ex.Plan)
		sb.WriteString(")")
	}
	return sb.String()
}

func draftExercisesPromptText(draft *state.PlanDraftSession) string {
	var sb strings.Builder
	sb.WriteString("📋 Тренировка \"")
	sb.WriteString(draft.Title)
	sb.WriteString("\"\n\nУпражнения:\n")
	for _, ex := range draft.Exercises {
		sb.WriteString("✅ ")
		sb.WriteString(ex.Name)
		sb.WriteString(" — ")
		sb.WriteString(ex.Plan)
		sb.WriteString("\n")
	}
	next := len(draft.Exercises) + 1
	sb.WriteString("\n➡️ Опиши ")
	sb.WriteString(ordinalRu(next))
	sb.WriteString(" упражнение:")
	return sb.String()
}

func ordinalRu(n int) string {
	switch n {
	case 1:
		return "первое"
	case 2:
		return "второе"
	case 3:
		return "третье"
	case 4:
		return "четвертое"
	default:
		return strconv.Itoa(n) + "-е"
	}
}
