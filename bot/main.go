package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbot "github.com/go-telegram/bot"
	_ "github.com/lib/pq"

	"traningBot/bot/config"
	botapp "traningBot/bot/internal/bot"
	"traningBot/bot/internal/bot/handlers"
	"traningBot/bot/internal/bot/state"
	"traningBot/bot/internal/scheduler"
	"traningBot/bot/internal/storage/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := postgres.New(db)
	if err := store.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b, err := tgbot.New(cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	app := &botapp.App{
		Bot:   b,
		Store: store,
		State: state.New(),
		Ctx:   ctx,
	}

	// Commands
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, handlers.Start(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/plan", tgbot.MatchTypeExact, handlers.Plan(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/done", tgbot.MatchTypeExact, handlers.Done(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/remind", tgbot.MatchTypeExact, handlers.Remind(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/stats", tgbot.MatchTypeExact, handlers.Stats(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/settings", tgbot.MatchTypeExact, handlers.Settings(app))

	// Menu buttons (same as commands)
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "📋 Мой план", tgbot.MatchTypeExact, handlers.Plan(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "✅ Я позанималась", tgbot.MatchTypeExact, handlers.Done(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "📊 Прогресс", tgbot.MatchTypeExact, handlers.Stats(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "⏰ Напомнить", tgbot.MatchTypeExact, handlers.Remind(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "⚙️ Настройки", tgbot.MatchTypeExact, handlers.Settings(app))

	// Callback buttons from reminders
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "snooze_", tgbot.MatchTypePrefix, handlers.HandleReminderCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "done_remind_", tgbot.MatchTypePrefix, handlers.HandleReminderCallbacks(app))

	// Default text handler (pending flows + unknown text)
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "", tgbot.MatchTypePrefix, handlers.Text(app))

	go scheduler.RunReminderLoop(ctx, store, b)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	b.Start(ctx)
}

