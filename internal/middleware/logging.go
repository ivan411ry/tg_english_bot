package middleware

import (
	"github.com/ivan411ry-afk/english_bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func WithLogging(next func(tgbotapi.Update)) func(tgbotapi.Update) {
	return func(update tgbotapi.Update) {

		logger.Log.Info(
			"incoming update",
			zap.String("text", getText(update)),
			zap.String("callback", getCallback(update)),
		)
		// Передаем дальше в основной обработчик
		next(update)
	}
}
func getText(update tgbotapi.Update) string {
	if update.Message != nil {
		return update.Message.Text
	}
	return ""
}
func getCallback(update tgbotapi.Update) string {
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Data
	}
	return ""
}
