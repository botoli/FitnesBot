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
		ResizeKeyboard:        true,
		IsPersistent:          true,
		InputFieldPlaceholder: "Число или кнопка ↓",
		Keyboard: [][]models.KeyboardButton{
			{{Text: BtnPlan}, {Text: BtnStats}},
			{{Text: BtnDone}, {Text: BtnRemind}},
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
			{{Text: "🏠 В меню", CallbackData: "doneflow_home"}},
		},
	}
}

// DoneFlowCancelRow — отмена записи тренировки (одна строка).
func DoneFlowCancelRow() [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{{Text: "❌ Отменить запись", CallbackData: "doneflow_cancel"}},
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
			{{Text: "✅ Начать сегодняшнюю", CallbackData: "menu_done"}, {Text: "✏️ Изменить план", CallbackData: "menu_addplan"}},
		},
	}
}

// PlanDayPickerInlineKeyboard — выбор дня или вся неделя.
func PlanDayPickerInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Пн", CallbackData: "planview_day_1"},
				{Text: "Вт", CallbackData: "planview_day_2"},
				{Text: "Ср", CallbackData: "planview_day_3"},
				{Text: "Чт", CallbackData: "planview_day_4"},
			},
			{
				{Text: "Пт", CallbackData: "planview_day_5"},
				{Text: "Сб", CallbackData: "planview_day_6"},
				{Text: "Вс", CallbackData: "planview_day_7"},
			},
			{{Text: "📅 Вся неделя", CallbackData: "planview_week"}},
		},
	}
}

// PlanDayNavInlineKeyboard — после выбора дня: назад и действия.
func PlanDayNavInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "◀️ К дням", CallbackData: "planview_back"}, {Text: "📅 Вся неделя", CallbackData: "planview_week"}},
			{{Text: "✅ Я позанималась", CallbackData: "menu_done"}, {Text: "✏️ Изменить план", CallbackData: "menu_addplan"}},
		},
	}
}

// RemindWorkoutInlineKeyboard — быстрые напоминания о тренировке + отмена.
func RemindWorkoutInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Сегодня 18:00", CallbackData: "remind_quick_td18"},
				{Text: "Сегодня 20:00", CallbackData: "remind_quick_td20"},
			},
			{
				{Text: "Завтра 08:00", CallbackData: "remind_quick_tm8"},
				{Text: "Завтра 19:00", CallbackData: "remind_quick_tm19"},
			},
			{{Text: "❌ Отмена", CallbackData: "remind_cancel"}},
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
