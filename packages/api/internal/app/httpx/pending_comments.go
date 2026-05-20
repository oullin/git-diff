package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) listPendingComments(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	repoRoot := r.URL.Query().Get("path")

	if repoRoot == "" {
		repoRoot = s.Repo
	}

	kind := r.URL.Query().Get("kind")
	sha := r.URL.Query().Get("sha")

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	comments, err := store.ListPendingComments(r.Context(), userID, repoRoot, kind, sha)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"comments": comments})
}

func (s Server) createPendingComment(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	var input storage.PendingCommentInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if strings.TrimSpace(input.RepoRoot) == "" {
		input.RepoRoot = s.Repo
	}

	if strings.TrimSpace(input.FilePath) == "" || strings.TrimSpace(input.DiffSection) == "" {
		writeError(w, http.StatusBadRequest, errors.New("filePath and diffSection are required"))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	id := "pending-" + randID(8)
	comment, err := store.CreatePendingComment(r.Context(), userID, id, input)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (s Server) updatePendingComment(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	id := r.PathValue("id")

	var body struct {
		BodyHTML string `json:"bodyHtml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	comment, err := store.UpdatePendingComment(r.Context(), userID, id, body.BodyHTML)

	if err != nil {
		writeError(w, http.StatusNotFound, err)

		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (s Server) deletePendingComment(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	id := r.PathValue("id")
	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	if err := store.DeletePendingComment(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s Server) promotePendingComments(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	var body struct {
		ReviewID string `json:"reviewId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if strings.TrimSpace(body.ReviewID) == "" {
		writeError(w, http.StatusBadRequest, errors.New("reviewId is required"))

		return
	}

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	promoted, err := store.PromotePendingComments(r.Context(), userID, body.ReviewID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"promoted": promoted})
}

func randID(n int) string {
	buf := make([]byte, n)

	if _, err := rand.Read(buf); err != nil {
		return "0"
	}

	return hex.EncodeToString(buf)
}
