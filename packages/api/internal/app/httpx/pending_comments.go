package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) handlePendingCommentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAuthenticationRequired):
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrInvalidCommentInput),
		errors.Is(err, service.ErrReviewIDRequired):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func (s Server) listPendingComments(w http.ResponseWriter, r *http.Request) {
	repoRoot := r.URL.Query().Get("path")

	if repoRoot == "" {
		repoRoot = s.Repo
	}

	comments, err := s.pending.List(
		r.Context(),
		s.Session.CurrentUserID(),
		repoRoot,
		r.URL.Query().Get("kind"),
		r.URL.Query().Get("sha"),
	)

	if err != nil {
		s.handlePendingCommentError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"comments": comments})
}

func (s Server) createPendingComment(w http.ResponseWriter, r *http.Request) {
	var input storage.PendingCommentInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	comment, err := s.pending.Create(
		r.Context(),
		s.Session.CurrentUserID(),
		s.Repo,
		input,
	)

	if err != nil {
		s.handlePendingCommentError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (s Server) updatePendingComment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	var body struct {
		BodyHTML string `json:"bodyHtml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	comment, err := s.pending.Update(
		r.Context(),
		s.Session.CurrentUserID(),
		id,
		body.BodyHTML,
	)

	if err != nil {
		if errors.Is(err, service.ErrAuthenticationRequired) {
			writeError(w, http.StatusUnauthorized, err)

			return
		}

		writeError(w, http.StatusNotFound, err)

		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (s Server) deletePendingComment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt64(w, r, "id")

	if !ok {
		return
	}

	if err := s.pending.Delete(
		r.Context(),
		s.Session.CurrentUserID(),
		id,
	); err != nil {
		s.handlePendingCommentError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s Server) promotePendingComments(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ReviewID int64 `json:"reviewId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	promoted, err := s.pending.Promote(
		r.Context(),
		s.Session.CurrentUserID(),
		body.ReviewID,
	)

	if err != nil {
		s.handlePendingCommentError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"promoted": promoted})
}
