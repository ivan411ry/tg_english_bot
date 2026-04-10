package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/ivan411ry-afk/english_bot/internal/logger"
	"github.com/ivan411ry-afk/english_bot/internal/state"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Practice
// 1. проверяет, что пользователь прошёл /start и есть в FSM
// 2. получает случайную карточку пользователя
// 3. переводит FSM в состояние ожидания ответа
// 4. отправляет слово с inline-кнопками
func (h *Handler) Practice(ctx context.Context, update tgbotapi.Update) {
	telegramID := update.Message.From.ID
	// Получаем состояние пользователя.
	// userID должен был сохраниться после /start.
	userState, stateErr := h.state.GetState(telegramID)
	if stateErr != nil {
		if errors.Is(stateErr, state.ErrStateNotFound) {
			logger.Log.Error(
				" no state found for user",
				zap.Int64("telegram_id", telegramID),
				zap.Error(stateErr),
			)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала нажмите /start")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error(
					"sending /practice precondition error",
					zap.Error(sendErr),
				)
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
			logger.Log.Error(
				"sending state error message failed",
				zap.Error(sendErr),
			)
		}
		return
	}

	userID := userState.UserID
	// Получаем случайную карточку пользователя.
	card, err := h.cardStorage.GetRandomCard(ctx, userID)
	if err != nil {
		logger.Log.Error(
			"getting card error",
			zap.Error(err),
			zap.Int("user_id", userID),
		)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет карточек. Добавьте через /add")
		if _, sendErr := h.bot.Send(msg); sendErr != nil {
			logger.Log.Error(
				"sending no-cards message error",
				zap.Error(sendErr),
			)
		}
		return
	}
	h.state.SetState(telegramID, state.StateWaitingForAnswer, state.Data{CardID: card.ID}, userID)
	// Создаём inline-кнопки.
	keyboard := buildPracticeKeyboard(card.ID)
	// Отправляем пользователю слово с кнопками.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, card.Word)
	msg.ReplyMarkup = keyboard
	if _, sendErr := h.bot.Send(msg); sendErr != nil {
		logger.Log.Error(
			"sending practice card error",
			zap.Error(sendErr),
		)
	}
}

// PracticeCallback обрабатывает нажатия inline-кнопок в режиме тренировки.
func (h *Handler) PracticeCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	if callback == nil || callback.Message == nil {
		logger.Log.Error("callback message is nil")
		if callback != nil {
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Сообщение устарело")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
		}
		return
	}
	telegramID := callback.From.ID
	// Получаем состояние пользователя из FSM.
	userState, stateErr := h.state.GetState(telegramID)
	if stateErr != nil {
		if errors.Is(stateErr, state.ErrStateNotFound) {
			logger.Log.Error(
				"no state in callback",
				zap.Int64("telegram_id", telegramID),
				zap.Error(stateErr),
			)

			msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Ошибка, начните заново /practice")
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				logger.Log.Error(
					"sending callback state error",
					zap.Error(sendErr),
				)
			}
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
			return
		}
		logger.Log.Error(
			"getting callback state error",
			zap.Error(stateErr),
			zap.Int64("telegram_id", telegramID),
		)

		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Ошибка, попробуйте позже")
		if _, sendErr := h.bot.Send(msg); sendErr != nil {
			logger.Log.Error(
				"sending callback state error",
				zap.Error(sendErr),
			)
		}
		if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "")); err != nil {
			logger.Log.Error(
				"answering callback error",
				zap.Error(err),
			)
		}
		return
	}
	userID := userState.UserID
	switch {
	case callback.Data == "next":
		if userState.State != state.StateWaitingForNext {
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Сначала покажите перевод")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
			return
		}
		card, err := h.cardStorage.GetRandomCard(ctx, userID)
		if err != nil {
			logger.Log.Error(
				"getting card error",
				zap.Error(err),
				zap.Int("user_id", userID),
			)
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Нет карточек")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
			return
		}
		h.state.SetState(telegramID, state.StateWaitingForAnswer, state.Data{CardID: card.ID}, userID)
		// Пересоздаём кнопки для новой карточки.
		keyboard := buildPracticeKeyboard(card.ID)
		// Редактируем старое сообщение, показываем новое слово.
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, card.Word)
		edit.ReplyMarkup = &keyboard
		if _, err := h.bot.Send(edit); err != nil {
			logger.Log.Error(
				"editing practice message error",
				zap.Error(err),
			)
		}
	// Формат callback.Data = "show_123".
	// Проверяем не полное равенство, а префикс, потому что ID карточки меняется.
	case strings.HasPrefix(callback.Data, "show_"):
		// Перевод можно показывать только когда мы действительно ждём ответ по карточке.
		if userState.State != state.StateWaitingForAnswer {
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Сначала нажмите 'Следующее'")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
			return
		}
		// Достаём cardID из строки вида "show_123".
		cardID, err := strconv.Atoi(strings.TrimPrefix(callback.Data, "show_"))
		if err != nil {
			logger.Log.Error(
				"parsing cardID error",
				zap.Error(err),
				zap.String("callback_data", callback.Data),
			)
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Ошибка")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
			return
		}
		// Получаем карточку по ID из БД.
		card, err := h.cardStorage.GetCardByID(ctx, cardID)
		if err != nil {
			logger.Log.Error(
				"getting card error",
				zap.Error(err),
				zap.Int("card_id", cardID),
			)
			if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Ошибка")); err != nil {
				logger.Log.Error(
					"answering callback error",
					zap.Error(err),
				)
			}
			return
		}
		// Оставляем те же кнопки.
		keyboard := buildPracticeKeyboard(card.ID)
		// Редактируем сообщение: теперь показываем слово + перевод.
		newText := card.Word + "\n\nПеревод: " + card.Translation
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, newText)
		edit.ReplyMarkup = &keyboard
		if _, sendErr := h.bot.Send(edit); sendErr != nil {
			logger.Log.Error(
				"editing translation message error",
				zap.Error(sendErr),
			)
		}
		// После показа перевода разрешаем переход к следующей карточке.
		h.state.SetState(telegramID, state.StateWaitingForNext, state.Data{CardID: cardID}, userID)
	default:
		// Любое неизвестное действие считаем ошибочным.
		if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "Что-то ты не то мутишь")); err != nil {
			logger.Log.Error(
				"answering callback error",
				zap.Error(err),
			)
		}
		return
	}
	if _, err := h.bot.Request(tgbotapi.NewCallback(callback.ID, "")); err != nil {
		logger.Log.Error(
			"answering callback error",
			zap.Error(err),
		)
	}
}

// buildPracticeKeyboard создает кнопки чтобы не дублировать кнопки постоянно
func buildPracticeKeyboard(cardID int) tgbotapi.InlineKeyboardMarkup {
	buttonShow := tgbotapi.NewInlineKeyboardButtonData("Показать перевод", "show_"+strconv.Itoa(cardID))
	buttonNext := tgbotapi.NewInlineKeyboardButtonData("Следующее", "next")
	row := tgbotapi.NewInlineKeyboardRow(buttonShow, buttonNext)
	return tgbotapi.NewInlineKeyboardMarkup(row)
}
