package storage

import (
	"context"
	"errors"

	"github.com/ivan411ry-afk/english_bot/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStorage - хранилище данных в PostgreSQL
type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{db: pool}
}

// GetUserIdByTelegramID ищет пользователя по telegram_id.
// Если пользователь не найден, возвращает ErrUserNotFound.
func (s *PostgresStorage) GetUserIdByTelegramID(ctx context.Context, telegramID int64) (int, error) {
	const query = `
	SELECT id
	FROM users
	WHERE telegram_id = $1
	`
	row := s.db.QueryRow(ctx, query, telegramID)
	var userID int
	err := row.Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return userID, nil
}

// CreateUser  создаёт пользователя и возвращает его внутренний id.
func (s *PostgresStorage) CreateUser(ctx context.Context, telegramID int64, username string) (int, error) {
	// Безопасно при конкуренции:
	// если два вызова /start выполнятся одновременно, ON CONFLICT вернёт id существующего пользователя.
	// Позже это уйдет в логику service
	const query = `
INSERT INTO users (telegram_id, username) 
VALUES ($1, $2)
ON CONFLICT (telegram_id) 
DO UPDATE SET username = EXCLUDED.username
RETURNING id
`
	var newUserID int
	err := s.db.QueryRow(ctx, query, telegramID, username).Scan(&newUserID)
	if err != nil {
		return 0, err
	}
	return newUserID, nil
}

// CreateCard создаёт новую карточку и возвращает её id.
func (s *PostgresStorage) CreateCard(ctx context.Context, userID int, word string, translation string) (int, error) {
	const query = `
INSERT INTO cards (user_id, word, translation)
VALUES ($1, $2, $3)
RETURNING id
`
	var cardID int
	err := s.db.QueryRow(ctx, query, userID, word, translation).Scan(&cardID)
	if err != nil {
		return 0, err
	}
	return cardID, nil
}

// GetRandomCard Выдает рандомную карту во время тренировки
func (s *PostgresStorage) GetRandomCard(ctx context.Context, userID int) (models.Card, error) {
	const query = `
SELECT id, word, translation
FROM cards
WHERE user_id = $1
ORDER BY random() 
LIMIT 1
`
	row := s.db.QueryRow(ctx, query, userID)
	var card models.Card
	err := row.Scan(&card.ID, &card.Word, &card.Translation)
	if err != nil {
		return models.Card{}, err
	}
	card.UserID = userID
	return card, nil
}

// GetCardByID Находит карту по ID. Нужна, чтобы выдавать правильный перевод
func (s *PostgresStorage) GetCardByID(ctx context.Context, cardID int, userID int) (models.Card, error) {
	const query = `
SELECT id, user_id, word, translation
FROM cards
WHERE id = $1 and user_id = $2
`
	row := s.db.QueryRow(ctx, query, cardID, userID)
	var card models.Card
	err := row.Scan(&card.ID, &card.UserID, &card.Word, &card.Translation)
	if err != nil {
		return models.Card{}, err
	}
	return card, nil
}
