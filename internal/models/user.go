package models

import "time"

type User struct {
	ID         int       `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	CreatedAt  time.Time `db:"created_at"`
}
