package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gocanto/git-diff/internal/storage/db"
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
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

func newWalkthroughRepo(conn *sql.DB, queries *db.Queries, clk *clock) *WalkthroughRepo {
	return &WalkthroughRepo{db: conn, queries: queries, clk: clk}
}

func (r *WalkthroughRepo) GetWalkthrough(ctx context.Context, repoRoot, contextKind, contextSHA string) (WalkthroughRecord, bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT repo_root, context_kind, context_sha, fingerprint, provider_id, model_id,
		       groups_json, summary, generated_at
		FROM walkthroughs
		WHERE repo_root = ? AND context_kind = ? AND context_sha = ?
	`, repoRoot, contextKind, contextSHA)

	var (
		record     WalkthroughRecord
		groupsJSON string
	)

	if err := row.Scan(
		&record.RepoRoot,
		&record.ContextKind,
		&record.ContextSHA,
		&record.Fingerprint,
		&record.ProviderID,
		&record.ModelID,
		&groupsJSON,
		&record.Summary,
		&record.GeneratedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WalkthroughRecord{}, false, nil
		}

		return WalkthroughRecord{}, false, err
	}

	if err := json.Unmarshal([]byte(groupsJSON), &record.Groups); err != nil {
		return WalkthroughRecord{}, false, fmt.Errorf("decode walkthrough groups: %w", err)
	}

	return record, true, nil
}

func (r *WalkthroughRepo) UpsertWalkthrough(ctx context.Context, record WalkthroughRecord) error {
	groupsJSON, err := json.Marshal(record.Groups)

	if err != nil {
		return fmt.Errorf("encode walkthrough groups: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO walkthroughs (
			repo_root, context_kind, context_sha, fingerprint, provider_id, model_id,
			groups_json, summary, generated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_root, context_kind, context_sha) DO UPDATE SET
			fingerprint  = excluded.fingerprint,
			provider_id  = excluded.provider_id,
			model_id     = excluded.model_id,
			groups_json  = excluded.groups_json,
			summary      = excluded.summary,
			generated_at = excluded.generated_at
	`, record.RepoRoot, record.ContextKind, record.ContextSHA, record.Fingerprint, record.ProviderID, record.ModelID,
		string(groupsJSON), record.Summary, record.GeneratedAt)

	return err
}
