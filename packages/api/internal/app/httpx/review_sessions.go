package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/oullin/git-diff/internal/service"
	"github.com/oullin/git-diff/internal/storage"
)

func (s Server) createReview(w http.ResponseWriter, r *http.Request) {
	var input storage.ReviewSessionStart

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	created, err := s.reviews.Create(r.Context(), s.Session.CurrentUserID(), input)

	switch {
	case errors.Is(err, service.ErrAuthenticationRequired):
		writeError(w, http.StatusUnauthorized, err)

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s Server) listReviews(w http.ResponseWriter, r *http.Request) {
	limit := int64(50)

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)

		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid limit: %w", err))

			return
		}

		limit = parsed
	}

	reviews, err := s.reviews.List(r.Context(), s.Session.CurrentUserID(), limit)

	switch {
	case errors.Is(err, service.ErrAuthenticationRequired):
		writeError(w, http.StatusUnauthorized, err)

		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"reviews": reviews})
}

func (s Server) reviewDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	detail, err := s.reviews.Detail(r.Context(), id)

	if err != nil {
		writeError(w, http.StatusNotFound, err)

		return
	}

	writeJSON(w, http.StatusOK, detail)
}

func (s Server) addReviewEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	var input storage.ReviewEventInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	event, err := s.reviews.AddEvent(r.Context(), id, input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, event)
}
