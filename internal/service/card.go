package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type CardStorage interface {
	CreateCard(ctx context.Context, userID int, word string, translation string) (int, error)
}
type CardService struct {
	cardStorage CardStorage
}

var ErrWordIsEmpty = errors.New("word is empty")
var ErrTranslationIsEmpty = errors.New("translation is empty")
var ErrCardExists = errors.New("card already exists")

func WordValidate(word string) error {
	if word == "" {
		return ErrWordIsEmpty
	}
	return nil
}

func TranslationValidate(translation string) error {
	if translation == "" {
		return ErrTranslationIsEmpty
	}
	return nil
}

func (c *CardService) AddCard(ctx context.Context, userID int, word string, translation string) error {
	err := WordValidate(word)
	if err != nil {
		return err
	}
	err = TranslationValidate(translation)
	if err != nil {
		return err
	}
	_, err = c.cardStorage.CreateCard(ctx, userID, word, translation)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrCardExists
			}
		}
		return err
	}
	return nil
}

func NewCardService(cardStorage CardStorage) *CardService {
	return &CardService{
		cardStorage: cardStorage,
	}
}
