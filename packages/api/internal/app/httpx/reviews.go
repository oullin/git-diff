package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	store, closeStore, err := s.Store(r.Context())

	if err == nil {
		defer closeStore()

		if s.Auth != nil && s.Auth.CurrentUserID() != 0 {
			userID := s.Auth.CurrentUserID()
			_, _ = store.SaveUIPreferences(r.Context(), userID, map[string]string{
				storage.PrefKeyLastRepoRoot: state.Root,
			})
			_, _ = store.UpsertRepository(r.Context(), userID, state.Root, "")
		}
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryRefresh(w http.ResponseWriter, r *http.Request) {
	s.repositoryOpen(w, r)
}

func (s Server) repositoryCommit(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	sha := r.URL.Query().Get("sha")

	if path == "" {
		path = s.Repo
	}

	if sha == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("sha is required"))

		return
	}

	state, err := review.ReadCommitState(r.Context(), path, sha)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryLog(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	commits, err := review.ListCommitLog(r.Context(), path, limit)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"commits": commits})
}

func (s Server) repositoryPullRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	prs, err := review.ListPullRequests(r.Context(), path, limit)

	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, review.ErrGhUnavailable) {
			status = http.StatusPreconditionFailed
		}

		writeError(w, status, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pullRequests": prs})
}

func (s Server) repositoryPullRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	numberRaw := r.URL.Query().Get("number")

	if path == "" {
		path = s.Repo
	}

	number, err := strconv.Atoi(numberRaw)

	if err != nil || number <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid number: %q", numberRaw))

		return
	}

	state, err := review.ReadPullRequestState(r.Context(), path, number)

	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, review.ErrGhUnavailable) {
			status = http.StatusPreconditionFailed
		}

		writeError(w, status, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryBranches(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	names, err := review.ListBranches(r.Context(), path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, storeErr := s.Store(r.Context())

	if storeErr != nil {
		writeJSON(w, http.StatusOK, map[string]any{"branches": names})

		return
	}

	defer closeStore()

	root, rootErr := review.ResolveRoot(r.Context(), path)

	if rootErr == nil {
		_ = store.SyncBranches(r.Context(), root, names)

		if records, listErr := store.ListBranches(r.Context(), root); listErr == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"branches": names,
				"records":  records,
			})

			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"branches": names})
}

func (s Server) repositoryCheckout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	if err := review.CheckoutBranch(r.Context(), body.Path, body.Branch); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	state, err := review.ReadRepositoryState(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryCreateBranch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	if err := review.CreateBranch(r.Context(), body.Path, body.Name); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	state, err := review.ReadRepositoryState(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryFile(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")

	if root == "" {
		root = s.Repo
	}

	if root == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("root is required"))

		return
	}

	if path == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("path is required"))

		return
	}

	file, err := review.ReadRepositoryFile(r.Context(), root, path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, file)
}

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

func randomID(prefix string) string {
	var bytes [12]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s-fallback", prefix)
	}

	return prefix + "-" + hex.EncodeToString(bytes[:])
}
