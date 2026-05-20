package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) createReview(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))

		return
	}

	var input storage.ReviewSessionStart

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if input.ID == "" {
		input.ID = randomID("review")
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	created, err := store.CreateReview(r.Context(), userID, input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s Server) listReviews(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))

		return
	}

	limit := int64(50)

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)

		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid limit: %w", err))

			return
		}

		limit = parsed
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	reviews, err := store.ListReviews(r.Context(), userID, limit)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"reviews": reviews})
}

func (s Server) reviewDetail(w http.ResponseWriter, r *http.Request) {
	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	detail, err := store.ReviewDetail(r.Context(), r.PathValue("id"))

	if err != nil {
		writeError(w, http.StatusNotFound, err)

		return
	}

	writeJSON(w, http.StatusOK, detail)
}

func (s Server) addReviewEvent(w http.ResponseWriter, r *http.Request) {
	var input storage.ReviewEventInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	event, err := store.AddReviewEvent(r.Context(), r.PathValue("id"), input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, event)
}

func randomID(prefix string) string {
	var bytes [12]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s-fallback", prefix)
	}

	return prefix + "-" + hex.EncodeToString(bytes[:])
}
