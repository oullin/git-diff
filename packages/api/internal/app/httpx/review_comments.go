package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/oullin/git-diff/internal/storage"
)

func (s Server) createReviewComment(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	var input storage.ReviewCommentInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	comment, err := s.reviews.CreateComment(r.Context(), reviewID, input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (s Server) updateReviewComment(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	commentID, ok := pathInt64(w, r, "commentId")

	if !ok {
		return
	}

	var input struct {
		BodyHTML string `json:"bodyHtml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	comment, err := s.reviews.UpdateComment(r.Context(), reviewID, commentID, input.BodyHTML)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (s Server) setReviewCommentResolved(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	commentID, ok := pathInt64(w, r, "commentId")

	if !ok {
		return
	}

	var input struct {
		Resolved bool `json:"resolved"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	comment, err := s.reviews.SetCommentResolved(r.Context(), reviewID, commentID, input.Resolved)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (s Server) deleteReviewComment(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	commentID, ok := pathInt64(w, r, "commentId")

	if !ok {
		return
	}

	if err := s.reviews.DeleteComment(r.Context(), reviewID, commentID); err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
