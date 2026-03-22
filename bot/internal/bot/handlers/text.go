package handlers

import (
	"context"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/copy"
	"traningBot/bot/internal/bot/keyboard"
	"traningBot/bot/internal/bot/state"
	stmodels "traningBot/bot/internal/storage/models"
	"traningBot/bot/internal/utils"
)

func isCancelIntent(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	switch t {
	case "отмена", "отменить", "назад", "cancel", "меню", "главное", "стоп", "хватит", "выход":
		return true
	default:
		return false
	}
}

func Text(app *botapp.App) func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		tgID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		if isCancelIntent(text) {
			app.State.Clear(tgID)
			SendHomeMessages(ctx, b, chatID, copy.MsgCancelAck)
			return
		}

		switch text {
		case keyboard.BtnPlan, "/plan":
			app.State.Clear(tgID)
			Plan(app)(ctx, b, update)
			return
		case keyboard.BtnAddPlan, "/addplan":
			app.State.Clear(tgID)
			AddPlan(app)(ctx, b, update)
			return
		case keyboard.BtnDone, "/done":
			app.State.Clear(tgID)
			Done(app)(ctx, b, update)
			return
		case keyboard.BtnStats, "/stats":
			app.State.Clear(tgID)
			Stats(app)(ctx, b, update)
			return
		case keyboard.BtnRemind, "/remind":
			app.State.Clear(tgID)
			Remind(app)(ctx, b, update)
			return
		case keyboard.BtnSettings, "/settings":
			app.State.Clear(tgID)
			Settings(app)(ctx, b, update)
			return
		case "/cancel":
			Cancel(app)(ctx, b, update)
			return
		case "/help":
			app.State.Clear(tgID)
			Help(app)(ctx, b, update)
			return
		}

		p, ok := app.State.Get(tgID)
		if ok {
			switch p.Kind {
			case state.PendingDoneFlow:
				HandlePendingDoneFlowInput(app)(ctx, b, update)
				return

			case state.PendingPlanAdd:
				HandlePendingPlanAdd(app)(ctx, b, update)
				return

			case state.PendingDoneReport:
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
					ReportText: text,
				})
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла записать отчет."})
					return
				}
				app.State.Clear(tgID)
				SendReplyWithQuickActions(ctx, b, chatID, "Записала! Так держать 💪")
				return

			case state.PendingRemind:
				u, err := app.Store.GetUserByTgID(ctx, tgID)
				if err != nil {
					app.State.Clear(tgID)
					return
				}
				st, _ := app.Store.GetUserSettings(ctx, u.ID)
				tm, msg, err := utils.ParseUserReminderInput(text, time.Now(), time.Local)
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
						ChatID:      chatID,
						Text:        "Не распознала. Формат: ДД.ММ.ГГГГ ЧЧ:ММ [текст]\nТекст можно не вводить — будет напоминание о тренировке.",
						ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
					})
					return
				}
				_, err = app.Store.CreateReminder(ctx, stmodels.Reminder{
					UserID:      u.ID,
					RemindAt:    tm,
					Message:     msg,
					IsRecurring: true,
					IntervalMin: st.ReminderIntervalMinutes,
				})
				if err != nil {
					_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: "Не смогла сохранить напоминание."})
					return
				}
				app.State.Clear(tgID)
				SendReplyWithQuickActions(ctx, b, chatID, "Ок, напомню о тренировке в "+tm.Format("02.01.2006 15:04")+".")
				return

			case state.PendingSettings:
				HandlePendingSettings(app)(ctx, b, update)
				return
			}
		}

		SendHomeMessages(ctx, b, chatID, copy.MsgUnknown)
	}
}
