package keyboard

import "github.com/go-telegram/bot/models"

const (
	BtnPlan     = "📋 Мой план"
	BtnDone     = "✅ Я позанималась"
	BtnStats    = "📊 Прогресс"
	BtnRemind   = "⏰ Напомнить"
	BtnAddPlan  = "➕ Добавить тренировку"
	BtnSettings = "⚙️ Настройки"
)

func MainMenuReplyKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: [][]models.KeyboardButton{
			{{Text: BtnPlan}, {Text: BtnStats}},
			{{Text: BtnAddPlan}},
			{{Text: BtnSettings}},
		},
	}
}

func QuickActionsInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: BtnDone, CallbackData: "menu_done"}, {Text: BtnRemind, CallbackData: "menu_remind"}},
		},
	}
}

func ReminderInlineKeyboard(reminderID int64) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Сделала!", CallbackData: "done_remind_" + itoa64(reminderID)},
				{Text: "⏳ Еще нет", CallbackData: "snooze_" + itoa64(reminderID)},
			},
		},
	}
}

func DoneFinalInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Посмотреть статистику 📈", CallbackData: "doneflow_stats"}},
		},
	}
}

func PlanSavedInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "📋 Мой план", CallbackData: "menu_plan"}, {Text: "🏋️ Начать тренировку сейчас", CallbackData: "menu_done"}},
		},
	}
}

func PlanViewInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "✅ Начать сегодняшнюю", CallbackData: "menu_done"}, {Text: "✏️ Редактировать план", CallbackData: "menu_addplan"}},
		},
	}
}

func EmptyInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}}
}

func SettingsInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Частота 5м", CallbackData: "set_freq_5"},
				{Text: "10м", CallbackData: "set_freq_10"},
				{Text: "15м", CallbackData: "set_freq_15"},
			},
			{
				{Text: "Тихие 23:00-08:00", CallbackData: "set_quiet_23_08"},
			},
			{
				{Text: "⬅️ Меню", CallbackData: "menu_home"},
			},
		},
	}
}

func itoa64(v int64) string {
	// small local helper to avoid strconv import across packages; keyboard is tiny so keep it here
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v%10]
		v /= 10
	}
	return string(buf[i:])
}
