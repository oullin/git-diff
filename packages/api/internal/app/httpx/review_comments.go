package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) createReviewComment(w http.ResponseWriter, r *http.Request) {
	var input storage.ReviewCommentInput

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

	comment, err := store.CreateReviewComment(r.Context(), r.PathValue("id"), randomID("comment"), input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (s Server) updateReviewComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BodyHTML string `json:"bodyHtml"`
	}

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

	comment, err := store.UpdateReviewComment(r.Context(), r.PathValue("id"), r.PathValue("commentId"), input.BodyHTML)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (s Server) deleteReviewComment(w http.ResponseWriter, r *http.Request) {
	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	if err := store.DeleteReviewComment(r.Context(), r.PathValue("id"), r.PathValue("commentId")); err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
