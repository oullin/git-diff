package httpx

import (
	"errors"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/service"
)

type fileSearchResult struct {
	RepoPath string `json:"repoPath"`
	RepoName string `json:"repoName"`
	FilePath string `json:"filePath"`
	Score    int    `json:"score"`
}

const (
	defaultSearchLimit = 50
	maxSearchLimit     = 200
	maxScanFiles       = 100_000
)

func (s Server) searchRepositoryFiles(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if query == "" {
		writeJSON(w, http.StatusOK, map[string]any{"results": []fileSearchResult{}})

		return
	}

	limit := defaultSearchLimit

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)

		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, errors.New("invalid limit"))

			return
		}

		if parsed > maxSearchLimit {
			parsed = maxSearchLimit
		}

		limit = parsed
	}

	repos, err := s.repos.List(r.Context(), s.Session.CurrentUserID())

	if err != nil {
		if errors.Is(err, service.ErrAuthenticationRequired) {
			writeError(w, http.StatusUnauthorized, err)

			return
		}

		writeError(w, http.StatusInternalServerError, err)

		return
	}

	needle := strings.ToLower(query)
	results := make([]fileSearchResult, 0, limit)
	scanned := 0

	for _, repo := range repos {
		if scanned >= maxScanFiles {
			break
		}

		files, err := review.ListRepositoryFiles(r.Context(), repo.Path)

		if err != nil {
			continue
		}

		repoName := repo.Name

		if repoName == "" {
			repoName = filepath.Base(repo.Path)
		}

		for _, file := range files {
			scanned++

			if scanned > maxScanFiles {
				break
			}

			score := scoreFileMatch(file, needle)

			if score <= 0 {
				continue
			}

			results = append(results, fileSearchResult{
				RepoPath: repo.Path,
				RepoName: repoName,
				FilePath: file,
				Score:    score,
			})
		}

		if err := r.Context().Err(); err != nil {
			break
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}

		if len(results[i].FilePath) != len(results[j].FilePath) {
			return len(results[i].FilePath) < len(results[j].FilePath)
		}

		return results[i].FilePath < results[j].FilePath
	})

	if len(results) > limit {
		results = results[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

// scoreFileMatch favours basename matches and exact prefixes; zero means
// no match. `needle` must already be lower-cased.
func scoreFileMatch(path, needle string) int {
	lowerPath := strings.ToLower(path)
	idx := strings.Index(lowerPath, needle)

	if idx < 0 {
		return 0
	}

	base := lowerPath

	if slash := strings.LastIndexByte(lowerPath, '/'); slash >= 0 {
		base = lowerPath[slash+1:]
	}

	score := 100

	if baseIdx := strings.Index(base, needle); baseIdx >= 0 {
		score += 200

		if baseIdx == 0 {
			score += 100
		}

		if base == needle {
			score += 200
		}
	} else if idx == 0 {
		score += 50
	}

	score -= idx

	if score < 1 {
		score = 1
	}

	return score
}
