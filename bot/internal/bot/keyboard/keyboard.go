package keyboard

import "github.com/go-telegram/bot/models"

func MainMenuInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📋 Мой план", CallbackData: "menu_plan"},
				{Text: "✅ Я позанималась", CallbackData: "menu_done"},
			},
			{
				{Text: "📊 Прогресс", CallbackData: "menu_stats"},
				{Text: "⏰ Напомнить", CallbackData: "menu_remind"},
			},
			{
				{Text: "⚙️ Настройки", CallbackData: "menu_settings"},
			},
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

func EmptyInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}}
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


