package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type ReviewSessionStart struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
	Branch       string `json:"branch"`
	HeadSHA      string `json:"headSha"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	FilesChanged int    `json:"filesChanged"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ContextKind  string `json:"contextKind"`
	ContextSHA   string `json:"contextSha"`
}

type ReviewSession struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
	UserID       int64  `json:"userId"`
	Branch       string `json:"branch"`
	HeadSHA      string `json:"headSha"`
	Status       string `json:"status"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	FilesChanged int    `json:"filesChanged"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	StartedAt    string `json:"startedAt"`
	CompletedAt  string `json:"completedAt,omitempty"`
	ContextKind  string `json:"contextKind"`
	ContextSHA   string `json:"contextSha,omitempty"`
}

type ReviewDetail struct {
	Review   ReviewSession   `json:"review"`
	Events   []ReviewEvent   `json:"events"`
	Comments []ReviewComment `json:"comments"`
}

type ReviewRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
	events  *ReviewEventRepo
}

func newReviewRepo(conn *sql.DB, queries *db.Queries, clk *clock, events *ReviewEventRepo) *ReviewRepo {
	return &ReviewRepo{db: conn, queries: queries, clk: clk, events: events}
}

func (r *ReviewRepo) CreateReview(ctx context.Context, userID int64, review ReviewSessionStart) (ReviewSession, error) {
	if userID == 0 {
		return ReviewSession{}, errors.New("user id is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	title := review.Title

	if title == "" {
		title = "Local review"
	}

	contextKind := review.ContextKind

	if contextKind == "" {
		contextKind = "working"
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO review_sessions (
			id, repo_root, user_id, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, context_kind, context_sha
		) VALUES (?, ?, ?, ?, ?, 'open', ?, ?, ?, ?, ?, ?, ?, ?)
	`, review.ID, review.RepoRoot, userID, review.Branch, review.HeadSHA, title, review.Summary, review.FilesChanged, review.Additions, review.Deletions, now, contextKind, nullString(review.ContextSHA))

	if err != nil {
		return ReviewSession{}, err
	}

	if _, err := r.events.Add(ctx, review.ID, ReviewEventInput{
		Type:    "review_started",
		Message: title,
	}); err != nil {
		return ReviewSession{}, err
	}

	return ReviewSession{
		ID:           review.ID,
		RepoRoot:     review.RepoRoot,
		UserID:       userID,
		Branch:       review.Branch,
		HeadSHA:      review.HeadSHA,
		Status:       "open",
		Title:        title,
		Summary:      review.Summary,
		FilesChanged: review.FilesChanged,
		Additions:    review.Additions,
		Deletions:    review.Deletions,
		StartedAt:    now,
		ContextKind:  contextKind,
		ContextSHA:   review.ContextSHA,
	}, nil
}

func (r *ReviewRepo) ListReviews(ctx context.Context, userID int64, limit int64) ([]ReviewSession, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, repo_root, user_id, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, completed_at,
			context_kind, context_sha
		FROM review_sessions
		WHERE user_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, userID, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reviews := []ReviewSession{}

	for rows.Next() {
		review, err := scanReview(rows)

		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

// GetReviewByID returns the session without its events or comments.
func (r *ReviewRepo) GetReviewByID(ctx context.Context, id string) (ReviewSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, repo_root, user_id, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, completed_at,
			context_kind, context_sha
		FROM review_sessions
		WHERE id = ?
	`, id)

	return scanReview(row)
}

func (r *ReviewRepo) ReviewDetail(ctx context.Context, comments *CommentRepo, id string) (ReviewDetail, error) {
	review, err := r.GetReviewByID(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	events, err := r.events.List(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	commentList, err := comments.ListReviewComments(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	return ReviewDetail{Review: review, Events: events, Comments: commentList}, nil
}

func scanReview(row scanner) (ReviewSession, error) {
	var review ReviewSession

	var (
		completedAt sql.NullString
		contextSHA  sql.NullString
	)

	if err := row.Scan(
		&review.ID,
		&review.RepoRoot,
		&review.UserID,
		&review.Branch,
		&review.HeadSHA,
		&review.Status,
		&review.Title,
		&review.Summary,
		&review.FilesChanged,
		&review.Additions,
		&review.Deletions,
		&review.StartedAt,
		&completedAt,
		&review.ContextKind,
		&contextSHA,
	); err != nil {
		return ReviewSession{}, err
	}

	review.CompletedAt = fromNull(completedAt)
	review.ContextSHA = fromNull(contextSHA)

	return review, nil
}
