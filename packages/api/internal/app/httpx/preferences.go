package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gocanto/git-diff/internal/service"
)

type savePreferencesRequest struct {
	Values map[string]string `json:"values"`
}

func (s Server) handlePreferenceError(w http.ResponseWriter, err error, wrap string) {
	if errors.Is(err, service.ErrAuthenticationRequired) {
		writeError(w, http.StatusUnauthorized, err)

		return
	}

	writeError(w, http.StatusInternalServerError, fmt.Errorf("%s: %w", wrap, err))
}

func (s Server) getPreferences(w http.ResponseWriter, r *http.Request) {
	prefs, err := s.PreferenceService.Get(r.Context(), s.Auth.CurrentUserID())

	if err != nil {
		s.handlePreferenceError(w, err, "read ui preferences")

		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

func (s Server) savePreferences(w http.ResponseWriter, r *http.Request) {
	var req savePreferencesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	prefs, err := s.PreferenceService.Save(r.Context(), s.Auth.CurrentUserID(), req.Values)

	if err != nil {
		s.handlePreferenceError(w, err, "save ui preferences")

		return
	}

	writeJSON(w, http.StatusOK, prefs)
}
