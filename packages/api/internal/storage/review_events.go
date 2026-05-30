package storage

import (
	"context"
	"time"

	"github.com/oullin/git-diff/internal/db"
)

type ReviewEventInput struct {
	Type     string `json:"type"`
	FilePath string `json:"filePath"`
	Message  string `json:"message"`
	Metadata string `json:"metadata"`
}

type ReviewEvent struct {
	ID        int64  `json:"id"`
	ReviewID  int64  `json:"reviewId"`
	Type      string `json:"type"`
	FilePath  string `json:"filePath,omitempty"`
	Message   string `json:"message,omitempty"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ReviewEventRepo struct{}

func newReviewEventRepo() *ReviewEventRepo {
	return &ReviewEventRepo{}
}

func (r *ReviewEventRepo) Add(ctx context.Context, reviewID int64, input ReviewEventInput) (ReviewEvent, error) {
	now := db.Now().UTC().Format(time.RFC3339Nano)
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
		UpdatedAt: now,
	}

	if err := db.Conn().WithContext(ctx).Create(&row).Error; err != nil {
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
		UpdatedAt: now,
	}, nil
}

func (r *ReviewEventRepo) List(ctx context.Context, reviewID int64) ([]ReviewEvent, error) {
	var rows []ReviewEventRow

	if err := db.Conn().WithContext(ctx).
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
			UpdatedAt: row.UpdatedAt,
		})
	}

	return events, nil
}
