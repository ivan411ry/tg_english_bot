package telegram

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivan411ry-afk/english_bot/internal/logger"
	"github.com/ivan411ry-afk/english_bot/internal/state"
	"go.uber.org/zap"
)

type command string

const (
	commandStart    command = "start"
	commandAdd      command = "add"
	commandPractice command = "practice"
)

type bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}
type AppHandlers interface {
	Start(ctx context.Context, update tgbotapi.Update)
	Add(ctx context.Context, update tgbotapi.Update)
	Practice(ctx context.Context, update tgbotapi.Update)
	PracticeCallback(ctx context.Context, callback *tgbotapi.CallbackQuery)
}
type stateManager interface {
	SetState(telegramID int64, state state.State, data state.Data, userID int)
	GetState(telegramID int64) (state.UserState, error)
	ClearState(telegramID int64)
}

type Router struct {
	bot          bot
	appHandlers  AppHandlers
	stateManager stateManager
}

func NewRouter(bot bot, appHandlers AppHandlers, stateManager stateManager) *Router {
	return &Router{
		bot:          bot,
		appHandlers:  appHandlers,
		stateManager: stateManager,
	}
}

// HandleUpdate - главная точка входа для всех апдейтов
func (r *Router) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Если это callback (нажатие на inline-кнопку)
	if update.CallbackQuery != nil {
		r.appHandlers.PracticeCallback(ctx, update.CallbackQuery)
		return
	}
	// Если это не сообщение — дальше не идем
	if update.Message == nil || update.Message.From == nil {
		return
	}
	// Если это команда (/start, /add...)
	if update.Message.IsCommand() {
		cmd := command(update.Message.Command())
		switch cmd {
		case commandStart:
			r.appHandlers.Start(ctx, update)
		case commandAdd:
			r.appHandlers.Add(ctx, update)
		case commandPractice:
			r.appHandlers.Practice(ctx, update)
		default:
			r.sendText(
				update.Message.Chat.ID,
				"Неизвестная команда. Используйте /start, /add или /practice",
				"sending unknown command message error",
			)
		}
		return
	}
	// Если это обычный текст
	r.handleTextMessage(ctx, update)
}

// handleTextMessage обрабатывает обычные текстовые сообщения
func (r *Router) handleTextMessage(ctx context.Context, update tgbotapi.Update) {
	// Быстрые кнопки (reply keyboard)
	switch update.Message.Text {
	case "Добавить слово":
		r.appHandlers.Add(ctx, update)
		return
	case "Начать тренировку":
		r.appHandlers.Practice(ctx, update)
		return
	case "Мои слова":
		r.sendText(
			update.Message.Chat.ID,
			"Не гони коней. В разработке пока что, скоро все будет",
			"sending my-words placeholder error",
		)
		return
	}
	// FSM: если пользователь сейчас в add-flow
	// то обычный текст надо отдать в Add
	telegramID := update.Message.From.ID

	userState, stateErr := r.stateManager.GetState(telegramID)
	if stateErr == nil {
		if userState.State == state.StateWaitingForWord ||
			userState.State == state.StateWaitingForTranslation {
			r.appHandlers.Add(ctx, update)
			return
		}
	} else if !errors.Is(stateErr, state.ErrStateNotFound) {
		logger.Log.Error(
			"failed to get user state in router",
			zap.Error(stateErr),
			zap.Int64("telegram_id", telegramID),
		)
	}
	// Если ничего не подошло
	logger.Log.Info(
		"unknown text message",
		zap.String("text", update.Message.Text),
		zap.Int64("telegram_id", telegramID),
	)
	r.sendText(
		update.Message.Chat.ID,
		"Не понял сообщение. Используйте /start, /add или /practice",
		"sending unknown text message error",
	)
}

// sendText - helper, чтобы не дублировать отправку простого текста
func (r *Router) sendText(chatID int64, text string, logMsg string) {
	replyText := tgbotapi.NewMessage(chatID, text)
	if _, err := r.bot.Send(replyText); err != nil {
		logger.Log.Error(logMsg, zap.Error(err))
	}
}
