package handlers

import (
	"context"

	tgbot "github.com/go-telegram/bot"

	"traningBot/bot/internal/bot/copy"
	"traningBot/bot/internal/bot/keyboard"
)

// SendHomeMessages — главный экран: reply-клавиатура + быстрые inline-кнопки (два сообщения — ограничение Telegram).
func SendHomeMessages(ctx context.Context, b *tgbot.Bot, chatID int64, intro string) {
	if intro == "" {
		intro = copy.HomeDefaultIntro()
	}
	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        intro,
		ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
	})
	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        copy.MsgQuickActions,
		ReplyMarkup: keyboard.QuickActionsInlineKeyboard(),
	})
}

// SendReplyWithQuickActions — ответ с основной клавиатурой и блоком быстрых действий (после успешных сценариев).
func SendReplyWithQuickActions(ctx context.Context, b *tgbot.Bot, chatID int64, text string) {
	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: keyboard.MainMenuReplyKeyboard(),
	})
	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        copy.MsgQuickActions,
		ReplyMarkup: keyboard.QuickActionsInlineKeyboard(),
	})
}
