package handlers

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivan411ry-afk/english_bot/internal/logger"
	"github.com/ivan411ry-afk/english_bot/internal/state"
	"github.com/ivan411ry-afk/english_bot/internal/storage"
	"go.uber.org/zap"
)

// Start обрабатывает команду /start.
// Логика такая:
// 1. получаем Telegram ID и username пользователя
// 2. пытаемся найти пользователя в БД
// 3. если пользователя нет — создаём
// 4. сохраняем начальное состояние в FSM
// 5. отправляем приветственное сообщение с клавиатурой
func (h *Handler) Start(ctx context.Context, update tgbotapi.Update) {

	// Получаем данные пользователя из Telegram update.
	telegramID := update.Message.From.ID
	username := update.Message.From.UserName

	// Пытаемся найти пользователя в БД.
	userID, err := h.userStorage.GetUserIdByTelegramID(ctx, telegramID)
	// Если пользователь не найден — создаём нового.
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			logger.Log.Error("failed to get user by telegram id", zap.Error(err))

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка, попробуйте позже")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("failed to send registration error message", zap.Error(sendErr))
			}
			return
		}

		userID, err = h.userStorage.CreateUser(ctx, telegramID, username)
		if err != nil {
			logger.Log.Error("failed to create user", zap.Error(err))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка, попробуйте позже")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("failed to send registration error message", zap.Error(sendErr))
			}
			return
		}
	}

	logger.Log.Info("user authorized", zap.Int("user_id", userID))

	// Устанавливаем базовое состояние пользователя в FSM.
	// StateIdle означает, что пользователь сейчас не находится внутри
	//add-flow или другого сценария.
	h.state.SetState(telegramID, state.StateIdle, state.Data{}, userID)

	// Создаём клавиатуру с основными действиями.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что пожелаете, хозяин?")
	if username == h.specialUsername {
		msg.Text = "Привет, моя королева 👑. Что пожелаешь?"
	}

	button1 := tgbotapi.NewKeyboardButton("Добавить слово")
	button2 := tgbotapi.NewKeyboardButton("Мои слова")
	button3 := tgbotapi.NewKeyboardButton("Начать тренировку")

	row := tgbotapi.NewKeyboardButtonRow(button1, button2, button3)
	keyboard := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = keyboard

	// Отправляем приветственное сообщение.
	if _, err := h.bot.Send(msg); err != nil {
		logger.Log.Error("helloMessage sending error", zap.Error(err))
	}
}
