package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gocanto/git-diff/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type authUserResponse struct {
	ID          int64  `json:"id"`
	OSUsername  string `json:"osUsername"`
	DisplayName string `json:"displayName"`
}

type authStateResponse struct {
	OSUsername      string `json:"osUsername"`
	NeedsSetup      bool   `json:"needsSetup"`
	IsAuthenticated bool   `json:"isAuthenticated"`
}

type authLoginResponse struct {
	Token string           `json:"token,omitempty"`
	User  authUserResponse `json:"user"`
}

type setupRequest struct {
	Password string `json:"password"`
}

type loginRequest struct {
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type resumeRequest struct {
	Token string `json:"token"`
}

type wipeRequest struct {
	OSUsername string `json:"osUsername"`
}

func (s Server) authState(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth state not initialized"))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open review database: %w", err))

		return
	}

	defer closeStore()

	user, err := store.GetUserByOSUsername(r.Context(), s.Auth.OSUsername())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("read user: %w", err))

		return
	}

	writeJSON(w, http.StatusOK, authStateResponse{
		OSUsername:      user.OSUsername,
		NeedsSetup:      !user.HasPassword,
		IsAuthenticated: s.Auth.CurrentUserID() == user.ID,
	})
}

func (s Server) authSetup(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth state not initialized"))

		return
	}

	var req setupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	if len(strings.TrimSpace(req.Password)) < minPasswordLength {
		writeError(w, http.StatusBadRequest, fmt.Errorf("password must be at least %d characters", minPasswordLength))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open review database: %w", err))

		return
	}

	defer closeStore()

	user, err := store.GetUserByOSUsername(r.Context(), s.Auth.OSUsername())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("read user: %w", err))

		return
	}

	if user.HasPassword {
		writeError(w, http.StatusConflict, errors.New("password already set; use the login flow"))

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("hash password: %w", err))

		return
	}

	if err := store.SetPasswordHash(r.Context(), user.ID, string(hash)); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("save password hash: %w", err))

		return
	}

	session, err := store.CreateSession(r.Context(), user.ID, sessionTTL)

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("create session: %w", err))

		return
	}

	_ = store.TouchUserLogin(r.Context(), user.ID)
	s.Auth.Set(user.ID, session.RawToken)

	writeJSON(w, http.StatusOK, authLoginResponse{
		Token: session.RawToken,
		User: authUserResponse{
			ID:          user.ID,
			OSUsername:  user.OSUsername,
			DisplayName: user.DisplayName,
		},
	})
}

func (s Server) authLogin(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth state not initialized"))

		return
	}

	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open review database: %w", err))

		return
	}

	defer closeStore()

	user, err := store.GetUserByOSUsername(r.Context(), s.Auth.OSUsername())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("read user: %w", err))

		return
	}

	if !user.HasPassword {
		writeError(w, http.StatusConflict, errors.New("password not set; complete setup first"))

		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, errors.New("incorrect password"))

		return
	}

	resp := authLoginResponse{
		User: authUserResponse{
			ID:          user.ID,
			OSUsername:  user.OSUsername,
			DisplayName: user.DisplayName,
		},
	}

	if req.Remember {
		session, err := store.CreateSession(r.Context(), user.ID, sessionTTL)

		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("create session: %w", err))

			return
		}

		resp.Token = session.RawToken
		s.Auth.Set(user.ID, session.RawToken)
	} else {
		s.Auth.Set(user.ID, "")
	}

	_ = store.TouchUserLogin(r.Context(), user.ID)

	writeJSON(w, http.StatusOK, resp)
}

func (s Server) authResume(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth state not initialized"))

		return
	}

	var req resumeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open review database: %w", err))

		return
	}

	defer closeStore()

	user, err := store.ResumeSession(r.Context(), req.Token)

	switch {
	case errors.Is(err, storage.ErrSessionNotFound):
		writeError(w, http.StatusUnauthorized, errors.New("session expired"))

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, fmt.Errorf("resume session: %w", err))

		return
	}

	s.Auth.Set(user.ID, req.Token)

	writeJSON(w, http.StatusOK, map[string]any{
		"user": authUserResponse{
			ID:          user.ID,
			OSUsername:  user.OSUsername,
			DisplayName: user.DisplayName,
		},
	})
}

func (s Server) authLogout(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth state not initialized"))

		return
	}

	token := s.Auth.CurrentToken()

	if token != "" {
		store, closeStore, err := s.Store(r.Context())

		if err == nil {
			defer closeStore()

			_ = store.DeleteSession(r.Context(), token)
		}
	}

	s.Auth.Clear()

	w.WriteHeader(http.StatusNoContent)
}

func (s Server) authWipe(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth state not initialized"))

		return
	}

	var req wipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	osUsername := strings.TrimSpace(req.OSUsername)

	if osUsername == "" {
		osUsername = s.Auth.OSUsername()
	}

	if osUsername != s.Auth.OSUsername() {
		writeError(w, http.StatusForbidden, errors.New("can only wipe the active OS user"))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open review database: %w", err))

		return
	}

	defer closeStore()

	user, err := store.GetUserByOSUsername(r.Context(), osUsername)

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("read user: %w", err))

		return
	}

	if err := store.WipeUser(r.Context(), user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("wipe user: %w", err))

		return
	}

	if _, err := store.EnsureUser(r.Context(), osUsername); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("reseed user: %w", err))

		return
	}

	s.Auth.Clear()

	w.WriteHeader(http.StatusNoContent)
}
