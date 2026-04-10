package models

import "time"

type Card struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	Word        string    `db:"word"`
	Translation string    `db:"translation"`
	CreatedAt   time.Time `db:"created_at"`
}
