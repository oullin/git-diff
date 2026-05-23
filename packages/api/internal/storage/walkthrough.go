package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalkthroughGroupFile mirrors walkthrough.FileEntry; the storage layer
// can't import the walkthrough package without an import cycle.
type WalkthroughGroupFile struct {
	Path   string `json:"path"`
	Note   string `json:"note"`
	Action string `json:"action"`
	Impact string `json:"impact"`
}

type WalkthroughGroup struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Rationale string                 `json:"rationale"`
	Files     []WalkthroughGroupFile `json:"files"`
}

// WalkthroughRecord caches one LLM walkthrough per (repo, context) pair.
// A drifted Fingerprint marks the row stale and forces regeneration.
type WalkthroughRecord struct {
	RepoRoot    string             `json:"repoRoot"`
	ContextKind string             `json:"contextKind"`
	ContextSHA  string             `json:"contextSha,omitempty"`
	Fingerprint string             `json:"fingerprint"`
	ProviderID  string             `json:"providerId"`
	ModelID     string             `json:"modelId"`
	Groups      []WalkthroughGroup `json:"groups"`
	Summary     string             `json:"summary"`
	GeneratedAt string             `json:"generatedAt"`
}

type WalkthroughRepo struct {
	db  *gorm.DB
	clk *clock
}

func newWalkthroughRepo(db *gorm.DB, clk *clock) *WalkthroughRepo {
	return &WalkthroughRepo{db: db, clk: clk}
}

func (r *WalkthroughRepo) GetWalkthrough(ctx context.Context, repoRoot, contextKind, contextSHA string) (WalkthroughRecord, bool, error) {
	var row WalkthroughRow

	err := r.db.WithContext(ctx).
		Where("repo_root = ? AND context_kind = ? AND context_sha = ?", repoRoot, contextKind, contextSHA).
		Take(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return WalkthroughRecord{}, false, nil
	}

	if err != nil {
		return WalkthroughRecord{}, false, err
	}

	record := WalkthroughRecord{
		RepoRoot:    row.RepoRoot,
		ContextKind: row.ContextKind,
		ContextSHA:  row.ContextSHA,
		Fingerprint: row.Fingerprint,
		ProviderID:  row.ProviderID,
		ModelID:     row.ModelID,
		Summary:     row.Summary,
		GeneratedAt: row.GeneratedAt,
	}

	if err := json.Unmarshal([]byte(row.GroupsJSON), &record.Groups); err != nil {
		return WalkthroughRecord{}, false, fmt.Errorf("decode walkthrough groups: %w", err)
	}

	return record, true, nil
}

func (r *WalkthroughRepo) UpsertWalkthrough(ctx context.Context, record WalkthroughRecord) error {
	groupsJSON, err := json.Marshal(record.Groups)

	if err != nil {
		return fmt.Errorf("encode walkthrough groups: %w", err)
	}

	row := WalkthroughRow{
		RepoRoot:    record.RepoRoot,
		ContextKind: record.ContextKind,
		ContextSHA:  record.ContextSHA,
		Fingerprint: record.Fingerprint,
		ProviderID:  record.ProviderID,
		ModelID:     record.ModelID,
		GroupsJSON:  string(groupsJSON),
		Summary:     record.Summary,
		GeneratedAt: record.GeneratedAt,
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "repo_root"}, {Name: "context_kind"}, {Name: "context_sha"}},
			DoUpdates: clause.AssignmentColumns([]string{"fingerprint", "provider_id", "model_id", "groups_json", "summary", "generated_at"}),
		}).
		Create(&row).Error
}
