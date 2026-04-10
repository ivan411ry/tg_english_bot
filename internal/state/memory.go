package state

import "sync"

type MemoryStateManager struct {
	mu     sync.RWMutex
	states map[int64]UserState
}

func (r *MemoryStateManager) SetState(telegramID int64, state State, data Data, userID int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[telegramID] = UserState{
		State:  state,
		Data:   data,
		UserID: userID,
	}
}

func (r *MemoryStateManager) GetState(telegramID int64) (UserState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nowState, ok := r.states[telegramID]
	if !ok {
		return UserState{}, ErrStateNotFound
	}
	return nowState, nil
}

func (r *MemoryStateManager) ClearState(telegramID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.states, telegramID)
}

func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		states: make(map[int64]UserState),
	}
}
