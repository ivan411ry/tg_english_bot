package handlers

import (
	"context"

	"github.com/ivan411ry-afk/english_bot/internal/models"
	"github.com/ivan411ry-afk/english_bot/internal/state"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}
type userStorage interface {
	GetUserIdByTelegramID(ctx context.Context, telegramID int64) (int, error)
	CreateUser(ctx context.Context, telegramID int64, username string) (int, error)
}
type cardStorage interface {
	CreateCard(ctx context.Context, userID int, word string, translation string) (int, error)
	GetRandomCard(ctx context.Context, userID int) (models.Card, error)
	GetCardByID(ctx context.Context, cardID int, userID int) (models.Card, error)
}
type stateManager interface {
	SetState(telegramID int64, state state.State, data state.Data, userID int)
	GetState(telegramID int64) (state.UserState, error)
	ClearState(telegramID int64)
}
type cardService interface {
	AddCard(ctx context.Context, userID int, word string, translation string) error
}

// Handler хранит общие зависимости для всех обработчиков.
type Handler struct {
	bot             bot
	userStorage     userStorage
	cardStorage     cardStorage
	state           stateManager
	cardService     cardService
	specialUsername string
}

func NewHandler(
	bot bot,
	userStorage userStorage,
	cardStorage cardStorage,
	stateManager stateManager,
	cardService cardService,
	specialUsername string,
) *Handler {
	return &Handler{
		bot:             bot,
		userStorage:     userStorage,
		cardStorage:     cardStorage,
		state:           stateManager,
		cardService:     cardService,
		specialUsername: specialUsername,
	}
}
