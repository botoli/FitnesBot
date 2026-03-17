package keyboard

import "github.com/go-telegram/bot/models"

func MainMenuKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "📋 Мой план"},
				{Text: "✅ Я позанималась"},
			},
			{
				{Text: "📊 Прогресс"},
				{Text: "⏰ Напомнить"},
			},
			{
				{Text: "⚙️ Настройки"},
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


