// Package copy — короткие тексты бота в одном месте (тон и разделители).
package copy

import (
	"strings"
	"time"
)

// Divider — визуальный разделитель блоков в сообщениях.
const Divider = "───────────────"

const (
	// MsgQuickActions — подпись ко второму сообщению с inline-кнопками.
	MsgQuickActions = "⚡ Быстрые действия"

	MsgCancelAck = "Окей, отменила.\n\n"
	MsgResetAck  = "Сбросила сценарий.\n\n"
	MsgUnknown   = "Не расслышала 🤔\n\nВыбери кнопку внизу или напиши /help."
)

// HomeDefaultIntro — если в SendHomeMessages передать пустую строку.
func HomeDefaultIntro() string {
	return "🏠 Главное меню\n\n" +
		"Снизу — разделы. Следующее сообщение — быстрые действия."
}

// Welcome приветствие; имя можно передать пустым.
func Welcome(displayName string) string {
	var sb strings.Builder
	sb.WriteString("Привет")
	if t := strings.TrimSpace(displayName); t != "" {
		sb.WriteString(", ")
		sb.WriteString(t)
	}
	sb.WriteString(" 👋\n\n")
	sb.WriteString("План, тренировки, прогресс и напоминания — в одном чате.\n\n")
	sb.WriteString("Меню — кнопки внизу. Ещё быстрее — «Быстрые действия» ниже.")
	return sb.String()
}

// RandomTip — короткая подсказка (ротация по времени).
func RandomTip() string {
	tips := []string{
		"Для напоминания достаточно даты и времени — текст о тренировке подставится сам.",
		"В «Мой план» можно смотреть один день кнопками Пн–Вс.",
		"Напиши «отмена» или /cancel — выйдем из любого сценария.",
		"Команда /help покажет список команд.",
	}
	if len(tips) == 0 {
		return ""
	}
	i := int(time.Now().UnixNano() % int64(len(tips)))
	return "💡 " + tips[i]
}

// Help возвращает текст справки (plain text).
func Help() string {
	return "📖 Справка\n" +
		Divider + "\n\n" +
		"Команды:\n" +
		"/start — главный экран\n" +
		"/plan — план тренировок\n" +
		"/done — записать тренировку\n" +
		"/stats — прогресс\n" +
		"/remind — напоминание о тренировке\n" +
		"/addplan — добавить тренировку в план\n" +
		"/settings — настройки напоминаний\n" +
		"/cancel — отменить текущий шаг\n" +
		"/help — эта подсказка\n\n" +
		"Совет: основные кнопки всегда под полем ввода."
}
