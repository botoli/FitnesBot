package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
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

//go:embed migrations/*.sql
var migrationFS embed.FS

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

	// Auto-run migrations (idempotent via IF NOT EXISTS)
	if err := runMigrations(context.Background(), db); err != nil {
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
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/addplan", tgbot.MatchTypeExact, handlers.AddPlan(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/done", tgbot.MatchTypeExact, handlers.Done(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/remind", tgbot.MatchTypeExact, handlers.Remind(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/stats", tgbot.MatchTypeExact, handlers.Stats(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/settings", tgbot.MatchTypeExact, handlers.Settings(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/cancel", tgbot.MatchTypeExact, handlers.Cancel(app))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/help", tgbot.MatchTypeExact, handlers.Help(app))

	// Inline callbacks (prefix order: specific before generic menu)
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "planview_", tgbot.MatchTypePrefix, handlers.HandlePlanViewCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "remind_quick_", tgbot.MatchTypePrefix, handlers.HandleRemindQuickCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "menu_", tgbot.MatchTypePrefix, handlers.MenuCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "set_", tgbot.MatchTypePrefix, handlers.SettingsCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "planadd_", tgbot.MatchTypePrefix, handlers.HandlePlanAddCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "statsf_", tgbot.MatchTypePrefix, handlers.HandleStatsCallbacks(app))

	// Callback buttons from reminders
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "snooze_", tgbot.MatchTypePrefix, handlers.HandleReminderCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "done_remind_", tgbot.MatchTypePrefix, handlers.HandleReminderCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "doneflow_", tgbot.MatchTypePrefix, handlers.HandleDoneFlowCallbacks(app))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "remind_cancel", tgbot.MatchTypeExact, handlers.HandleRemindCancel(app))

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
	fmt.Println("Bot started")
	b.Start(ctx)

}

func runMigrations(ctx context.Context, db *sql.DB) error {
	sqlBytes, err := migrationFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(sqlBytes))
	return err
}
