package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

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

type ReviewEventRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

func newReviewEventRepo(conn *sql.DB, queries *db.Queries, clk *clock) *ReviewEventRepo {
	return &ReviewEventRepo{db: conn, queries: queries, clk: clk}
}

func (r *ReviewEventRepo) Add(ctx context.Context, reviewID string, input ReviewEventInput) (ReviewEvent, error) {
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

func (r *ReviewEventRepo) List(ctx context.Context, reviewID string) ([]ReviewEvent, error) {
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
