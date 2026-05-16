package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) repositoryState(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	state, err := review.ReadRepositoryState(r.Context(), path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryOpen(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	state, err := review.ReadRepositoryState(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err == nil {
		defer closeStore()

		prefs, _ := store.GetUserPreferences(r.Context())
		prefs.LastRepoRoot = state.Root
		_, _ = store.SaveUserPreferences(r.Context(), prefs)
		_, _ = store.UpsertRepository(r.Context(), state.Root, "")
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryRefresh(w http.ResponseWriter, r *http.Request) {
	s.repositoryOpen(w, r)
}

func (s Server) createReview(w http.ResponseWriter, r *http.Request) {
	var input storage.ReviewSessionStart

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if input.ID == "" {
		input.ID = randomID("review")
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	created, err := store.CreateReview(r.Context(), input)

	if err != nil {
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

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	reviews, err := store.ListReviews(r.Context(), limit)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"reviews": reviews})
}

func (s Server) reviewDetail(w http.ResponseWriter, r *http.Request) {
	store, closeStore, err := s.WorkflowStore(r.Context())

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

	store, closeStore, err := s.WorkflowStore(r.Context())

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

func (s Server) createReviewComment(w http.ResponseWriter, r *http.Request) {
	var input storage.ReviewCommentInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

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

	store, closeStore, err := s.WorkflowStore(r.Context())

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
	store, closeStore, err := s.WorkflowStore(r.Context())

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

func randomID(prefix string) string {
	var bytes [12]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s-fallback", prefix)
	}

	return prefix + "-" + hex.EncodeToString(bytes[:])
}
