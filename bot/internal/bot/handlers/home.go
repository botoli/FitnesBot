package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"

	"traningBot/bot/internal/bot/keyboard"
)

// SendHomeMessages отправляет главный экран: ответная клавиатура + быстрые inline-действия.
// В Telegram одно сообщение не может одновременно нести reply- и inline-клавиатуру — поэтому два сообщения.
func SendHomeMessages(ctx context.Context, b *tgbot.Bot, chatID int64, intro string) {
	if intro == "" {
		intro = "🏠 Главное меню\n\n" +
			"Снизу — основные разделы. Под следующим сообщением — быстрые действия: записать тренировку и напоминание."
	}
	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        intro,
		ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
	})
	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        "⚡ Быстрые действия:",
		ReplyMarkup: keyboard.QuickActionsInlineKeyboard(),
	})
}
