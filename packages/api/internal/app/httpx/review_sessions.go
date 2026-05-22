package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) createReview(w http.ResponseWriter, r *http.Request) {
	var input storage.ReviewSessionStart

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	created, err := s.Services.reviews.Create(r.Context(), s.Auth.CurrentUserID(), input)

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

	reviews, err := s.Services.reviews.List(r.Context(), s.Auth.CurrentUserID(), limit)

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
	detail, err := s.Services.reviews.Detail(r.Context(), r.PathValue("id"))

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

	event, err := s.Services.reviews.AddEvent(r.Context(), r.PathValue("id"), input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, event)
}
