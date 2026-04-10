package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	DatabaseURL    string
	UpdatesTimeout time.Duration
}

func Load() (*Config, error) {
	// Пытаемся загрузить .env
	// Если файла нет, не падаем, переменные могут уже прийти из окружения
	//например через docker compose env_file
	_ = godotenv.Load()
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN empty")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not found")
	}

	timeout := 60 * time.Second
	if raw := os.Getenv("BOT_UPDATES_TIMEOUT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("BOT_UPDATES_TIMEOUT must be positive number")
		}
		timeout = time.Duration(parsed) * time.Second
	}

	return &Config{
		BotToken:       botToken,
		DatabaseURL:    databaseURL,
		UpdatesTimeout: timeout,
	}, nil
}
