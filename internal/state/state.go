package state

import "errors"

type State string

const (
	StateIdle                  State = "idle"
	StateWaitingForWord        State = "waiting_for_word"
	StateWaitingForTranslation State = "waiting_for_translation"
	StateWaitingForAnswer      State = "waiting_for_answer"
	StateWaitingForNext        State = "waiting_for_next"
)

var ErrStateNotFound = errors.New("state not found")

type Data struct {
	Word   string
	CardID int
}

type StateManager interface {
	SetState(telegramID int64, state State, data Data, userID int)
	GetState(telegramID int64) (UserState, error)
	ClearState(telegramID int64)
}

type UserState struct {
	State  State
	Data   Data
	UserID int
}
