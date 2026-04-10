package storage

import (
	"errors"
)

// ErrUserNotFound возвращается, когда Get(...) не находит пользователя.
var ErrUserNotFound = errors.New("user not found")
