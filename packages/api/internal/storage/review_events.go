package storage

import (
	"context"
	"time"

	"gorm.io/gorm"
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
	db  *gorm.DB
	clk *clock
}

func newReviewEventRepo(db *gorm.DB, clk *clock) *ReviewEventRepo {
	return &ReviewEventRepo{db: db, clk: clk}
}

func (r *ReviewEventRepo) Add(ctx context.Context, reviewID string, input ReviewEventInput) (ReviewEvent, error) {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	metadata := input.Metadata

	if metadata == "" {
		metadata = "{}"
	}

	row := ReviewEventRow{
		ReviewID:  reviewID,
		EventType: input.Type,
		FilePath:  input.FilePath,
		Message:   input.Message,
		Metadata:  metadata,
		CreatedAt: now,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return ReviewEvent{}, err
	}

	return ReviewEvent{
		ID:        row.ID,
		ReviewID:  reviewID,
		Type:      input.Type,
		FilePath:  input.FilePath,
		Message:   input.Message,
		Metadata:  metadata,
		CreatedAt: now,
	}, nil
}

func (r *ReviewEventRepo) List(ctx context.Context, reviewID string) ([]ReviewEvent, error) {
	var rows []ReviewEventRow

	if err := r.db.WithContext(ctx).
		Where("review_id = ?", reviewID).
		Order("created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	events := make([]ReviewEvent, 0, len(rows))

	for _, row := range rows {
		events = append(events, ReviewEvent{
			ID:        row.ID,
			ReviewID:  row.ReviewID,
			Type:      row.EventType,
			FilePath:  row.FilePath,
			Message:   row.Message,
			Metadata:  row.Metadata,
			CreatedAt: row.CreatedAt,
		})
	}

	return events, nil
}
