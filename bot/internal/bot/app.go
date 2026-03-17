package botapp

import (
	"context"

	tgbot "github.com/go-telegram/bot"

	"traningBot/bot/internal/bot/state"
	"traningBot/bot/internal/storage/postgres"
)

type App struct {
	Bot   *tgbot.Bot
	Store *postgres.Store
	State *state.Store
	Ctx   context.Context
}

