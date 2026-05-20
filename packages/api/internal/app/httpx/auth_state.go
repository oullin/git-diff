package httpx

import (
	"sync"
	"time"
)

const (
	bcryptCost        = 12
	sessionTTL        = 90 * 24 * time.Hour
	minPasswordLength = 6
)

type AuthState struct {
	mu            sync.RWMutex
	osUsername    string
	currentUserID int64
	currentToken  string
}

func NewAuthState(osUsername string) *AuthState {
	return &AuthState{osUsername: osUsername}
}

func (a *AuthState) OSUsername() string {
	a.mu.RLock()

	defer a.mu.RUnlock()

	return a.osUsername
}

func (a *AuthState) CurrentUserID() int64 {
	a.mu.RLock()

	defer a.mu.RUnlock()

	return a.currentUserID
}

func (a *AuthState) CurrentToken() string {
	a.mu.RLock()

	defer a.mu.RUnlock()

	return a.currentToken
}

func (a *AuthState) Set(userID int64, token string) {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.currentUserID = userID
	a.currentToken = token
}

func (a *AuthState) Clear() {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.currentUserID = 0
	a.currentToken = ""
}
