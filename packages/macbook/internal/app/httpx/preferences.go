package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type savePreferencesRequest struct {
	Values map[string]string `json:"values"`
}

func (s Server) getPreferences(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil || s.Auth.CurrentUserID() == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open workflow log database: %w", err))

		return
	}

	defer closeStore()

	prefs, err := store.GetUIPreferences(r.Context(), s.Auth.CurrentUserID())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("read ui preferences: %w", err))

		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

func (s Server) savePreferences(w http.ResponseWriter, r *http.Request) {
	if s.Auth == nil || s.Auth.CurrentUserID() == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	var req savePreferencesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("open workflow log database: %w", err))

		return
	}

	defer closeStore()

	prefs, err := store.SaveUIPreferences(r.Context(), s.Auth.CurrentUserID(), req.Values)

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("save ui preferences: %w", err))

		return
	}

	writeJSON(w, http.StatusOK, prefs)
}
