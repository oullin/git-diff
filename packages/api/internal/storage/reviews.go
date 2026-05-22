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

type ReviewEventInput struct {
	Type     string `json:"type"`
	FilePath string `json:"filePath"`
	Message  string `json:"message"`
	Metadata string `json:"metadata"`
}

type ReviewEvent struct {
	ID        int64  `json:"id"`
	ReviewID  string `json:"reviewId"`
	Type      string `json:"type"`
	FilePath  string `json:"filePath,omitempty"`
	Message   string `json:"message,omitempty"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"createdAt"`
}

type ReviewDetail struct {
	Review   ReviewSession   `json:"review"`
	Events   []ReviewEvent   `json:"events"`
	Comments []ReviewComment `json:"comments"`
}

// ReviewEventWriter lets other repos log timeline events without depending
// on the concrete ReviewRepo.
type ReviewEventWriter interface {
	AddReviewEvent(ctx context.Context, reviewID string, input ReviewEventInput) (ReviewEvent, error)
}

type ReviewRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

func newReviewRepo(conn *sql.DB, queries *db.Queries, clk *clock) *ReviewRepo {
	return &ReviewRepo{db: conn, queries: queries, clk: clk}
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

	if _, err := r.AddReviewEvent(ctx, review.ID, ReviewEventInput{
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

	events, err := r.ListReviewEvents(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	commentList, err := comments.ListReviewComments(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	return ReviewDetail{Review: review, Events: events, Comments: commentList}, nil
}

func (r *ReviewRepo) AddReviewEvent(ctx context.Context, reviewID string, input ReviewEventInput) (ReviewEvent, error) {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	metadata := input.Metadata

	if metadata == "" {
		metadata = "{}"
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO review_events (review_id, event_type, file_path, message, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reviewID, input.Type, input.FilePath, input.Message, metadata, now)

	if err != nil {
		return ReviewEvent{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return ReviewEvent{}, err
	}

	return ReviewEvent{ID: id, ReviewID: reviewID, Type: input.Type, FilePath: input.FilePath, Message: input.Message, Metadata: metadata, CreatedAt: now}, nil
}

func (r *ReviewRepo) ListReviewEvents(ctx context.Context, reviewID string) ([]ReviewEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, review_id, event_type, file_path, message, metadata, created_at
		FROM review_events
		WHERE review_id = ?
		ORDER BY created_at ASC, id ASC
	`, reviewID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	events := []ReviewEvent{}

	for rows.Next() {
		var event ReviewEvent

		if err := rows.Scan(&event.ID, &event.ReviewID, &event.Type, &event.FilePath, &event.Message, &event.Metadata, &event.CreatedAt); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, rows.Err()
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
