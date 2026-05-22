package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
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

func userToResponse(user storage.User) authUserResponse {
	return authUserResponse{
		ID:          user.ID,
		OSUsername:  user.OSUsername,
		DisplayName: user.DisplayName,
	}
}

func (s Server) requireAuthSetup(w http.ResponseWriter) bool {
	if s.Auth == nil || s.Services.auth == nil {
		writeError(w, http.StatusInternalServerError, errors.New("auth not initialized"))

		return false
	}

	return true
}

func (s Server) authState(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuthSetup(w) {
		return
	}

	user, err := s.Services.auth.State(r.Context(), s.Auth.OSUsername())

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
	if !s.requireAuthSetup(w) {
		return
	}

	var req setupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	user, session, err := s.Services.auth.Setup(r.Context(), s.Auth.OSUsername(), req.Password)

	switch {
	case errors.Is(err, service.ErrPasswordTooShort):
		writeError(w, http.StatusBadRequest, fmt.Errorf("password must be at least %d characters", minPasswordLength))

		return
	case errors.Is(err, service.ErrPasswordAlreadySet):
		writeError(w, http.StatusConflict, errors.New("password already set; use the login flow"))

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	s.Auth.Set(user.ID, session.RawToken)

	writeJSON(w, http.StatusOK, authLoginResponse{
		Token: session.RawToken,
		User:  userToResponse(user),
	})
}

func (s Server) authLogin(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuthSetup(w) {
		return
	}

	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	user, session, err := s.Services.auth.Login(r.Context(), s.Auth.OSUsername(), req.Password, req.Remember)

	switch {
	case errors.Is(err, service.ErrPasswordNotSet):
		writeError(w, http.StatusConflict, errors.New("password not set; complete setup first"))

		return
	case errors.Is(err, service.ErrInvalidPassword):
		writeError(w, http.StatusUnauthorized, errors.New("incorrect password"))

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	resp := authLoginResponse{User: userToResponse(user)}

	if session != nil {
		resp.Token = session.RawToken
		s.Auth.Set(user.ID, session.RawToken)
	} else {
		s.Auth.Set(user.ID, "")
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s Server) authResume(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuthSetup(w) {
		return
	}

	var req resumeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	user, err := s.Services.auth.Resume(r.Context(), req.Token)

	switch {
	case errors.Is(err, storage.ErrSessionNotFound):
		writeError(w, http.StatusUnauthorized, errors.New("session expired"))

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, fmt.Errorf("resume session: %w", err))

		return
	}

	s.Auth.Set(user.ID, req.Token)

	writeJSON(w, http.StatusOK, map[string]any{"user": userToResponse(user)})
}

func (s Server) authLogout(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuthSetup(w) {
		return
	}

	_ = s.Services.auth.Logout(r.Context(), s.Auth.CurrentToken())
	s.Auth.Clear()

	w.WriteHeader(http.StatusNoContent)
}

func (s Server) authWipe(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuthSetup(w) {
		return
	}

	var req wipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	err := s.Services.auth.Wipe(r.Context(), s.Auth.OSUsername(), req.OSUsername)

	switch {
	case errors.Is(err, service.ErrCannotWipeOtherUser):
		writeError(w, http.StatusForbidden, err)

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	s.Auth.Clear()

	w.WriteHeader(http.StatusNoContent)
}
