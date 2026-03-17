package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	BotToken string
	DSN      string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:     os.Getenv("PORT"),
		BotToken: os.Getenv("BOT_TOKEN"),
		DSN:      os.Getenv("DATABASE_URL"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.BotToken == "" {
		return Config{}, errors.New("BOT_TOKEN is required")
	}
	if cfg.DSN == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

