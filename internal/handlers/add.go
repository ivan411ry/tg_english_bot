package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/ivan411ry-afk/english_bot/internal/logger"
	"github.com/ivan411ry-afk/english_bot/internal/service"
	"github.com/ivan411ry-afk/english_bot/internal/state"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Add обрабатывает сценарий добавления карточки.
// Этот метод работает как FSM:
// 1. пользователь нажал /add -> просим ввести слово
// 2. пользователь ввёл слово -> просим перевод
// 3. пользователь ввёл перевод -> сохраняем карточку
func (h *Handler) Add(ctx context.Context, update tgbotapi.Update) {
	telegramID := update.Message.From.ID
	// Достаём текущее состояние пользователя из FSM.
	userState, stateErr := h.state.GetState(telegramID)
	if stateErr != nil {
		if errors.Is(stateErr, state.ErrStateNotFound) {
			logger.Log.Error("no state found for user", zap.Int64("telegram_id", telegramID))

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала нажмите /start")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("sending /add precondition error", zap.Error(sendErr))
			}
			return
		}
		logger.Log.Error(
			"getting user state error",
			zap.Error(stateErr),
			zap.Int64("telegram_id", telegramID),
		)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка, попробуйте позже")
		if _, sendErr := h.bot.Send(msg); sendErr != nil {
			logger.Log.Error("sending error message failed", zap.Error(sendErr))
		}
		return
	}

	userID := userState.UserID
	currentState := userState.State

	switch currentState {
	case state.StateIdle:
		h.state.SetState(telegramID, state.StateWaitingForWord, state.Data{}, userID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите слово")
		if _, err := h.bot.Send(msg); err != nil {
			logger.Log.Error("sending ask-word message error", zap.Error(err))
		}
		return
	case state.StateWaitingForWord:
		word := strings.TrimSpace(update.Message.Text)
		err := service.WordValidate(word)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Слово не может быть пустым. Введите слово")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("sending empty-word message error", zap.Error(sendErr))
			}
			return
		}
		h.state.SetState(telegramID, state.StateWaitingForTranslation, state.Data{Word: word}, userID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите перевод")
		if _, sendErr := h.bot.Send(msg); sendErr != nil {
			logger.Log.Error("sending ask-translation message error", zap.Error(sendErr))
		}
		return
	case state.StateWaitingForTranslation:
		word := userState.Data.Word
		translation := strings.TrimSpace(update.Message.Text)

		err := h.cardService.AddCard(ctx, userID, word, translation)

		if errors.Is(err, service.ErrCardExists) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такая карточка уже есть")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("sending duplicate-card message error", zap.Error(sendErr))
			}
			return
		}

		if errors.Is(err, service.ErrTranslationIsEmpty) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Перевод не может быть пустым. Введите перевод")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("sending empty-translation message error", zap.Error(sendErr))
			}
			return
		}
		if err != nil {
			logger.Log.Error("card saving error", zap.Error(err))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка, попробуйте позже")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error("sending save-card error message failed", zap.Error(sendErr))
			}
			return
		}
		h.state.SetState(telegramID, state.StateIdle, state.Data{}, userID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Слово сохранено!")
		if _, sendErr := h.bot.Send(msg); sendErr != nil {
			logger.Log.Error("sending save success message error", zap.Error(sendErr))
		}
		return
		// страховка от багов и неожиданных состояний
	default:
		logger.Log.Info("add flow: unknown state",
			zap.String("state", string(currentState)),
			zap.Int64("telegram_id", telegramID))
		h.state.SetState(telegramID, state.StateIdle, state.Data{}, userID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Состояние сброшено. Нажмите /add ещё раз")
		if _, sendErr := h.bot.Send(msg); sendErr != nil {
			logger.Log.Error("sending unknown-state message error", zap.Error(sendErr))
		}
		return

	}
}
