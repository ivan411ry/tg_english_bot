package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Внутренние пакеты проекта
	"github.com/ivan411ry-afk/english_bot/internal/config"
	"github.com/ivan411ry-afk/english_bot/internal/handlers"
	"github.com/ivan411ry-afk/english_bot/internal/logger"
	"github.com/ivan411ry-afk/english_bot/internal/middleware"
	"github.com/ivan411ry-afk/english_bot/internal/service"
	"github.com/ivan411ry-afk/english_bot/internal/state"
	"github.com/ivan411ry-afk/english_bot/internal/storage"
	"github.com/ivan411ry-afk/english_bot/internal/telegram"
	"github.com/ivan411ry-afk/english_bot/pkg/sanitize"

	// Telegram API
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	// PostgreSQL драйвер с пулом соединений
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	err := logger.Init()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Log.Sync()
	}()
	// Загружаем конфиг (токен бота, БД и т.д.)
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal("configuration loading error", zap.Error(err))
	}

	// Создаем контекст, который отменится при SIGINT (Ctrl+C) или SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Инициализируем Telegram-бота
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		logger.Log.Fatal("bot creation error", zap.String("error", sanitize.Error(err, cfg.BotToken)))
	}
	logger.Log.Info("bot created successfully")

	// Подключаемся к PostgreSQL через пул соединений
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatal("database connection error", zap.Error(err))
	}
	logger.Log.Info("database connected")
	defer pool.Close() // закрываем пул при завершении

	// Создаем реализацию storage (Postgres)
	pgStorage := storage.NewPostgresStorage(pool)
	// FSM (состояния пользователей) — в памяти
	stateManager := state.NewMemoryStateManager()

	userStorage := pgStorage
	cardStorage := pgStorage
	cardService := service.NewCardService(cardStorage)
	// Создаем хендлеры
	appHandlers := handlers.NewHandler(bot, userStorage, cardStorage, stateManager, cardService)
	logger.Log.Info("handlers initialized")

	// Настройка получения обновлений от Telegram
	updatesConfig := tgbotapi.NewUpdate(0)
	updatesConfig.Timeout = int(cfg.UpdatesTimeout / time.Second)
	// Канал с апдейтами
	getUpdates := bot.GetUpdatesChan(updatesConfig)
	// Создаём роутер
	router := telegram.NewRouter(bot, appHandlers, stateManager)
	// Оборачиваем обработчик в middleware (логирование)
	handleUpdateWithMiddleware := middleware.WithLogging(func(update tgbotapi.Update) {
		router.HandleUpdate(ctx, update)
	})

	// Горутина для graceful shutdown
	go func() {
		<-ctx.Done()
		logger.Log.Info("shutdown signal received, stopping updates")
		bot.StopReceivingUpdates() // останавливаем получение апдейтов
	}()

	logger.Log.Info("bot is running and waiting for updates")

	// Основной цикл обработки апдейтов
	for update := range getUpdates {
		handleUpdateWithMiddleware(update)
	}

	logger.Log.Info("updates loop stopped, shutting down gracefully")
}
