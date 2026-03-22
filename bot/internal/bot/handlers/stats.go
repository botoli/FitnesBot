package handlers

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/keyboard"
	stmodels "traningBot/bot/internal/storage/models"
)

var statsNumberRe = regexp.MustCompile(`[-+]?\d+(?:[\.,]\d+)?`)

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

		reports, err := app.Store.ListReports(ctx, u.ID, 20)
		if err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не получилось загрузить историю."})
			return
		}

		if len(reports) == 0 {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Пока нет отчётов. После тренировки нажми «✅ Я позанималась» под быстрыми кнопками.",
				ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
			})
			return
		}

		parsed := parseStatsReports(reports)
		allTitles := statsTopTitles(parsed, 3)
		text := renderStatsMessage(parsed, "")

		_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ReplyMarkup: statsFiltersInlineKeyboard(allTitles, ""),
		})
	}
}

func HandleStatsCallbacks(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
			return
		}
		data := update.CallbackQuery.Data
		if !strings.HasPrefix(data, "statsf_") {
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

		reports, err := app.Store.ListReports(ctx, u.ID, 20)
		if err != nil || len(reports) == 0 {
			return
		}

		parsed := parseStatsReports(reports)
		titles := statsTopTitles(parsed, 3)

		selectedTitle := ""
		if data != "statsf_all" {
			idx, convErr := strconv.Atoi(strings.TrimPrefix(data, "statsf_"))
			if convErr == nil && idx >= 1 && idx <= len(titles) {
				selectedTitle = titles[idx-1]
			}
		}

		text := renderStatsMessage(parsed, selectedTitle)
		if _, err = b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   msgID,
			Text:        text,
			ReplyMarkup: statsFiltersInlineKeyboard(titles, selectedTitle),
		}); err != nil {
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:      chatID,
				Text:        text,
				ReplyMarkup: statsFiltersInlineKeyboard(titles, selectedTitle),
			})
		}
	}
}

type statsExercise struct {
	Name     string
	Actual   string
	Plan     string
	HasDelta bool
	Delta    float64
}

type statsReportView struct {
	CreatedAt time.Time
	Title     string
	Exercises []statsExercise
	RawText   string
}

func parseStatsReports(reports []stmodels.Report) []statsReportView {
	out := make([]statsReportView, 0, len(reports))
	for _, r := range reports {
		view := statsReportView{
			CreatedAt: r.CreatedAt,
			Title:     extractWorkoutTitle(r.ReportText),
			Exercises: parseReportExercises(r.ReportText),
			RawText:   r.ReportText,
		}
		if view.Title == "" {
			view.Title = "Тренировка"
		}
		out = append(out, view)
	}
	return out
}

func extractWorkoutTitle(text string) string {
	lines := strings.Split(text, "\n")
	for _, l := range lines {
		line := strings.TrimSpace(l)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Тренировка:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Тренировка:"))
		}
		return trimToRunes(line, 24)
	}
	return ""
}

func parseReportExercises(text string) []statsExercise {
	lines := strings.Split(text, "\n")
	out := make([]statsExercise, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, "-") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}

		name := strings.TrimSpace(line[:idx])
		rest := strings.TrimSpace(line[idx+1:])
		actual := rest
		plan := ""
		if pStart := strings.Index(rest, "(план:"); pStart >= 0 {
			actual = strings.TrimSpace(rest[:pStart])
			planPart := strings.TrimSpace(rest[pStart+len("(план:"):])
			planPart = strings.TrimSuffix(planPart, ")")
			plan = strings.TrimSpace(planPart)
		}

		ex := statsExercise{Name: name, Actual: actual, Plan: plan}
		a, okA := parseStatsFloat(actual)
		p, okP := parseStatsFloat(plan)
		if okA && okP {
			ex.HasDelta = true
			ex.Delta = a - p
		}
		out = append(out, ex)
	}
	return out
}

func parseStatsFloat(text string) (float64, bool) {
	m := statsNumberRe.FindString(strings.ReplaceAll(text, ",", "."))
	if m == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func renderStatsMessage(parsed []statsReportView, selectedTitle string) string {
	filtered := make([]statsReportView, 0, len(parsed))
	for _, p := range parsed {
		if selectedTitle == "" || p.Title == selectedTitle {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) > 5 {
		filtered = filtered[:5]
	}

	if len(filtered) == 0 {
		if selectedTitle == "" {
			return "📊 Последние 5 тренировок\n\nПока нет данных."
		}
		return "📊 Последние 5 тренировок\n\nПо фильтру «" + selectedTitle + "» пока нет данных."
	}

	var sb strings.Builder
	sb.WriteString("📊 Последние 5 тренировок\n\n")
	separator := "━━━━━━━━━━━━━━━━━━━━━━━━━━━"

	for i, r := range filtered {
		sb.WriteString(separator)
		sb.WriteString("\n")
		sb.WriteString("**")
		sb.WriteString(r.CreatedAt.Format("02.01 15:04"))
		sb.WriteString(" · ")
		sb.WriteString(r.Title)
		sb.WriteString("**\n")
		sb.WriteString(separator)
		sb.WriteString("\n")

		total := len(r.Exercises)
		if total == 0 {
			sb.WriteString("✅ Запись добавлена\n")
			sb.WriteString("📝 ")
			sb.WriteString(trimToRunes(strings.ReplaceAll(r.RawText, "\n", " "), 80))
			sb.WriteString("\n\n")
			continue
		}

		pos := 0.0
		neg := 0.0
		bestName := ""
		bestDelta := -1e9
		for _, ex := range r.Exercises {
			if !ex.HasDelta {
				continue
			}
			if ex.Delta > 0 {
				pos += ex.Delta
			}
			if ex.Delta < 0 {
				neg += ex.Delta
			}
			if ex.Delta > bestDelta {
				bestDelta = ex.Delta
				bestName = ex.Name
			}
		}

		if i > 0 && total > 3 {
			sb.WriteString("✅ ")
			sb.WriteString(strconv.Itoa(total))
			sb.WriteString(" упражнений (показаны первые 3)\n")
		} else {
			sb.WriteString("✅ Выполнено: ")
			sb.WriteString(strconv.Itoa(total))
			sb.WriteString("/")
			sb.WriteString(strconv.Itoa(total))
			sb.WriteString(" упражнений\n")
		}
		if pos > 0.001 {
			sb.WriteString("📈 Прогресс: +")
			sb.WriteString(shortStatsFloat(pos))
			sb.WriteString(" к плану\n")
		}
		if neg < -0.001 {
			sb.WriteString("📉 Отставание: ")
			sb.WriteString(shortStatsFloat(neg))
			sb.WriteString("\n")
		}
		if bestName != "" && bestDelta > 0.001 {
			sb.WriteString("🏆 Лучшее: ")
			sb.WriteString(bestName)
			sb.WriteString(" (+")
			sb.WriteString(shortStatsFloat(bestDelta))
			sb.WriteString(")\n")
		}
		sb.WriteString("\n")

		limit := total
		if i > 0 && total > 3 {
			limit = 3
		}

		for j := 0; j < limit; j++ {
			ex := r.Exercises[j]
			sb.WriteString("▸ ")
			sb.WriteString(padRightRunes(trimToRunes(ex.Name, 22), 22))
			sb.WriteString("  ")
			sb.WriteString(ex.Actual)
			if ex.HasDelta {
				sb.WriteString(" (")
				sb.WriteString(statsDeltaLabel(ex.Delta))
				sb.WriteString(")")
			}
			sb.WriteString("\n")
		}
		if limit < total {
			sb.WriteString("▸ ... и ещё ")
			sb.WriteString(strconv.Itoa(total - limit))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}

func statsTopTitles(parsed []statsReportView, max int) []string {
	seen := map[string]bool{}
	out := make([]string, 0, max)
	for _, p := range parsed {
		if p.Title == "" || seen[p.Title] {
			continue
		}
		seen[p.Title] = true
		out = append(out, p.Title)
		if len(out) == max {
			break
		}
	}
	return out
}

func statsFiltersInlineKeyboard(titles []string, selected string) *models.InlineKeyboardMarkup {
	row := make([]models.InlineKeyboardButton, 0, len(titles)+1)
	for i, t := range titles {
		label := "📎 " + t
		if t == selected {
			label = "• " + label
		}
		row = append(row, models.InlineKeyboardButton{Text: trimToRunes(label, 24), CallbackData: "statsf_" + strconv.Itoa(i+1)})
	}
	allLabel := "📎 Всё"
	if selected == "" {
		allLabel = "• " + allLabel
	}
	row = append(row, models.InlineKeyboardButton{Text: allLabel, CallbackData: "statsf_all"})
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{row}}
}

func statsDeltaLabel(delta float64) string {
	if math.Abs(delta) < 0.001 {
		return "~"
	}
	if delta > 0 {
		return "+" + shortStatsFloat(delta)
	}
	return shortStatsFloat(delta)
}

func shortStatsFloat(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func trimToRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func padRightRunes(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
}
