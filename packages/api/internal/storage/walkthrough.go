package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gocanto/git-diff/internal/storage/db"
)

// WalkthroughRecord is the cached output of an LLM walkthrough for a specific
// (repo, context) pair. Fingerprint is the deterministic state hash; if it
// drifts from what the latest RepositoryState produces, the cached row is
// considered stale and should be re-generated.
type WalkthroughRecord struct {
	RepoRoot    string            `json:"repoRoot"`
	ContextKind string            `json:"contextKind"`
	ContextSHA  string            `json:"contextSha,omitempty"`
	Fingerprint string            `json:"fingerprint"`
	ModelID     string            `json:"modelId"`
	Order       []string          `json:"order"`
	Notes       map[string]string `json:"notes"`
	Summary     string            `json:"summary"`
	GeneratedAt string            `json:"generatedAt"`
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
		SELECT repo_root, context_kind, context_sha, fingerprint, model_id, order_json, notes_json, summary, generated_at
		FROM walkthroughs
		WHERE repo_root = ? AND context_kind = ? AND context_sha = ?
	`, repoRoot, contextKind, contextSHA)

	var (
		record    WalkthroughRecord
		orderJSON string
		notesJSON string
	)

	if err := row.Scan(
		&record.RepoRoot,
		&record.ContextKind,
		&record.ContextSHA,
		&record.Fingerprint,
		&record.ModelID,
		&orderJSON,
		&notesJSON,
		&record.Summary,
		&record.GeneratedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WalkthroughRecord{}, false, nil
		}

		return WalkthroughRecord{}, false, err
	}

	if err := json.Unmarshal([]byte(orderJSON), &record.Order); err != nil {
		return WalkthroughRecord{}, false, fmt.Errorf("decode walkthrough order: %w", err)
	}

	if err := json.Unmarshal([]byte(notesJSON), &record.Notes); err != nil {
		return WalkthroughRecord{}, false, fmt.Errorf("decode walkthrough notes: %w", err)
	}

	return record, true, nil
}

func (r *WalkthroughRepo) UpsertWalkthrough(ctx context.Context, record WalkthroughRecord) error {
	orderJSON, err := json.Marshal(record.Order)

	if err != nil {
		return fmt.Errorf("encode walkthrough order: %w", err)
	}

	notesJSON, err := json.Marshal(record.Notes)

	if err != nil {
		return fmt.Errorf("encode walkthrough notes: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO walkthroughs (
			repo_root, context_kind, context_sha, fingerprint, model_id,
			order_json, notes_json, summary, generated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_root, context_kind, context_sha) DO UPDATE SET
			fingerprint  = excluded.fingerprint,
			model_id     = excluded.model_id,
			order_json   = excluded.order_json,
			notes_json   = excluded.notes_json,
			summary      = excluded.summary,
			generated_at = excluded.generated_at
	`, record.RepoRoot, record.ContextKind, record.ContextSHA, record.Fingerprint, record.ModelID,
		string(orderJSON), string(notesJSON), record.Summary, record.GeneratedAt)

	return err
}
