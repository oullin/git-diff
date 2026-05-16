package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocanto/git-diff/internal/domain"
	"github.com/gocanto/git-diff/internal/storage/db"
	_ "modernc.org/sqlite"
)

const DefaultTheme = "light"
const DefaultDiffViewMode = "split"

//go:embed schema.sql
var schemaFS embed.FS

type Store struct {
	db      *sql.DB
	queries *db.Queries
	now     func() time.Time
}

type RunStart struct {
	ID                      string
	WorkflowID              string
	WorkflowName            string
	ConfirmationOptionID    string
	ConfirmationOptionLabel string
	Mode                    domain.RunMode
	Status                  domain.RunStatus
}

type RunSummary struct {
	ID                      string `json:"id"`
	WorkflowID              string `json:"workflowId"`
	WorkflowName            string `json:"workflowName"`
	ConfirmationOptionID    string `json:"confirmationOptionId"`
	ConfirmationOptionLabel string `json:"confirmationOptionLabel"`
	Mode                    string `json:"mode"`
	Status                  string `json:"status"`
	StartedAt               string `json:"startedAt"`
	CompletedAt             string `json:"completedAt,omitempty"`
	ErrorMessage            string `json:"errorMessage,omitempty"`
}

type EventRecord struct {
	ID        int64  `json:"id"`
	RunID     string `json:"runId"`
	Seq       int64  `json:"seq"`
	Type      string `json:"type"`
	PhaseID   string `json:"phaseId,omitempty"`
	PhaseName string `json:"phaseName,omitempty"`
	Status    string `json:"status,omitempty"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type RunLog struct {
	Run    RunSummary    `json:"run"`
	Events []EventRecord `json:"events"`
}

type UserPreferences struct {
	Theme          string `json:"theme"`
	DiffViewMode   string `json:"diffViewMode"`
	HideWhitespace bool   `json:"hideWhitespace"`
	LastRepoRoot   string `json:"lastRepoRoot"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

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
}

type ReviewSession struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
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

type ReviewCommentInput struct {
	FilePath    string `json:"filePath"`
	DiffSection string `json:"diffSection"`
	Side        string `json:"side"`
	LineNumber  int64  `json:"lineNumber"`
	AuthorLabel string `json:"authorLabel"`
	BodyHTML    string `json:"bodyHtml"`
}

type ReviewComment struct {
	ID          string `json:"id"`
	ReviewID    string `json:"reviewId"`
	FilePath    string `json:"filePath"`
	DiffSection string `json:"diffSection"`
	Side        string `json:"side"`
	LineNumber  int64  `json:"lineNumber"`
	AuthorLabel string `json:"authorLabel"`
	BodyHTML    string `json:"bodyHtml"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	DeletedAt   string `json:"deletedAt,omitempty"`
}

type ReviewDetail struct {
	Review   ReviewSession   `json:"review"`
	Events   []ReviewEvent   `json:"events"`
	Comments []ReviewComment `json:"comments"`
}

type Repository struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	AddedAt      string `json:"addedAt"`
	LastOpenedAt string `json:"lastOpenedAt,omitempty"`
}

type Recorder struct {
	store *Store
	runID string
	mu    sync.Mutex
	seq   int64
	also  func(domain.Event) error
}

type scanner interface {
	Scan(dest ...any) error
}

const envDBPath = "GIT_DIFF_WORKFLOW_DB_PATH"

func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	store := &Store{db: conn, queries: db.New(conn), now: time.Now}

	if err := store.Init(ctx); err != nil {
		_ = conn.Close()

		return nil, err
	}

	return store, nil
}

func DefaultPath(home string) string {
	if override := os.Getenv(envDBPath); override != "" {
		return override
	}

	return filepath.Join(home, "Library", "Application Support", "git-diff", "reviews.sqlite3")
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	schema, err := schemaFS.ReadFile("schema.sql")

	if err != nil {
		return fmt.Errorf("read embedded sqlite schema: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("initialize sqlite schema: %w", err)
	}

	if err := s.migratePreferences(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) migratePreferences(ctx context.Context) error {
	migrations := []string{
		"ALTER TABLE user_preferences ADD COLUMN diff_view_mode TEXT NOT NULL DEFAULT 'split'",
		"ALTER TABLE user_preferences ADD COLUMN hide_whitespace INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE user_preferences ADD COLUMN last_repo_root TEXT NOT NULL DEFAULT ''",
	}

	for _, migration := range migrations {
		if _, err := s.db.ExecContext(ctx, migration); err != nil && !isDuplicateColumn(err) {
			return fmt.Errorf("migrate user preferences: %w", err)
		}
	}

	return nil
}

func (s *Store) CreateRun(ctx context.Context, run RunStart) error {
	return s.queries.CreateRun(ctx, db.CreateRunParams{
		ID:                      run.ID,
		WorkflowID:              run.WorkflowID,
		WorkflowName:            run.WorkflowName,
		ConfirmationOptionID:    run.ConfirmationOptionID,
		ConfirmationOptionLabel: run.ConfirmationOptionLabel,
		Mode:                    string(run.Mode),
		Status:                  string(run.Status),
		StartedAt:               s.now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Store) CompleteRun(ctx context.Context, id string, status domain.RunStatus, message string) error {
	return s.queries.CompleteRun(ctx, db.CompleteRunParams{
		ID:           id,
		Status:       string(status),
		CompletedAt:  nullString(s.now().UTC().Format(time.RFC3339Nano)),
		ErrorMessage: nullString(message),
	})
}

func (s *Store) InsertEvent(ctx context.Context, event domain.Event) error {
	return s.queries.InsertEvent(ctx, db.InsertEventParams{
		RunID:     event.RunID,
		Seq:       event.Seq,
		EventType: event.Type,
		PhaseID:   nullString(event.PhaseID),
		PhaseName: nullString(event.PhaseName),
		Status:    nullString(event.Status),
		Message:   nullString(event.Message),
		CreatedAt: s.now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Store) ListRuns(ctx context.Context, limit int64) ([]RunSummary, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.queries.ListRuns(ctx, limit)

	if err != nil {
		return nil, err
	}

	runs := make([]RunSummary, 0, len(rows))

	for _, row := range rows {
		runs = append(runs, runSummary(row))
	}

	return runs, nil
}

func (s *Store) RunLog(ctx context.Context, runID string) (RunLog, error) {
	run, err := s.queries.GetRun(ctx, runID)

	if err != nil {
		return RunLog{}, err
	}

	rows, err := s.queries.ListRunEvents(ctx, runID)

	if err != nil {
		return RunLog{}, err
	}

	events := make([]EventRecord, 0, len(rows))

	for _, row := range rows {
		events = append(events, eventRecord(row))
	}

	return RunLog{Run: runSummary(run), Events: events}, nil
}

func (s *Store) GetUserPreferences(ctx context.Context) (UserPreferences, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT theme, diff_view_mode, hide_whitespace, last_repo_root, updated_at
		FROM user_preferences
		WHERE id = 1
	`)

	var prefs UserPreferences

	var hideWhitespace int
	err := row.Scan(&prefs.Theme, &prefs.DiffViewMode, &hideWhitespace, &prefs.LastRepoRoot, &prefs.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return defaultUserPreferences(), nil
	}

	if err != nil {
		return UserPreferences{}, err
	}

	prefs.HideWhitespace = hideWhitespace != 0

	if prefs.Theme == "" {
		prefs.Theme = DefaultTheme
	}

	if prefs.DiffViewMode == "" {
		prefs.DiffViewMode = DefaultDiffViewMode
	}

	return prefs, nil
}

func (s *Store) SaveUserPreferences(ctx context.Context, prefs UserPreferences) (UserPreferences, error) {
	theme := prefs.Theme

	if theme == "" {
		theme = DefaultTheme
	}

	diffViewMode := prefs.DiffViewMode

	if diffViewMode == "" {
		diffViewMode = DefaultDiffViewMode
	}

	updatedAt := s.now().UTC().Format(time.RFC3339Nano)
	hideWhitespace := 0

	if prefs.HideWhitespace {
		hideWhitespace = 1
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_preferences (id, theme, diff_view_mode, hide_whitespace, last_repo_root, updated_at)
		VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			theme = excluded.theme,
			diff_view_mode = excluded.diff_view_mode,
			hide_whitespace = excluded.hide_whitespace,
			last_repo_root = excluded.last_repo_root,
			updated_at = excluded.updated_at
	`, theme, diffViewMode, hideWhitespace, prefs.LastRepoRoot, updatedAt)

	if err != nil {
		return UserPreferences{}, err
	}

	return UserPreferences{
		Theme:          theme,
		DiffViewMode:   diffViewMode,
		HideWhitespace: prefs.HideWhitespace,
		LastRepoRoot:   prefs.LastRepoRoot,
		UpdatedAt:      updatedAt,
	}, nil
}

func defaultUserPreferences() UserPreferences {
	return UserPreferences{Theme: DefaultTheme, DiffViewMode: DefaultDiffViewMode}
}

func (s *Store) CreateReview(ctx context.Context, review ReviewSessionStart) (ReviewSession, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)
	title := review.Title

	if title == "" {
		title = "Local review"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO review_sessions (
			id, repo_root, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at
		) VALUES (?, ?, ?, ?, 'open', ?, ?, ?, ?, ?, ?)
	`, review.ID, review.RepoRoot, review.Branch, review.HeadSHA, title, review.Summary, review.FilesChanged, review.Additions, review.Deletions, now)

	if err != nil {
		return ReviewSession{}, err
	}

	if _, err := s.AddReviewEvent(ctx, review.ID, ReviewEventInput{
		Type:    "review_started",
		Message: title,
	}); err != nil {
		return ReviewSession{}, err
	}

	return ReviewSession{
		ID:           review.ID,
		RepoRoot:     review.RepoRoot,
		Branch:       review.Branch,
		HeadSHA:      review.HeadSHA,
		Status:       "open",
		Title:        title,
		Summary:      review.Summary,
		FilesChanged: review.FilesChanged,
		Additions:    review.Additions,
		Deletions:    review.Deletions,
		StartedAt:    now,
	}, nil
}

func (s *Store) ListReviews(ctx context.Context, limit int64) ([]ReviewSession, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, repo_root, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, completed_at
		FROM review_sessions
		ORDER BY started_at DESC
		LIMIT ?
	`, limit)

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

func (s *Store) ReviewDetail(ctx context.Context, id string) (ReviewDetail, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, repo_root, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, completed_at
		FROM review_sessions
		WHERE id = ?
	`, id)
	review, err := scanReview(row)

	if err != nil {
		return ReviewDetail{}, err
	}

	events, err := s.ListReviewEvents(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	comments, err := s.ListReviewComments(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	return ReviewDetail{Review: review, Events: events, Comments: comments}, nil
}

func (s *Store) AddReviewEvent(ctx context.Context, reviewID string, input ReviewEventInput) (ReviewEvent, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)
	metadata := input.Metadata

	if metadata == "" {
		metadata = "{}"
	}

	result, err := s.db.ExecContext(ctx, `
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

func (s *Store) ListReviewEvents(ctx context.Context, reviewID string) ([]ReviewEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
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

func (s *Store) CreateReviewComment(ctx context.Context, reviewID string, commentID string, input ReviewCommentInput) (ReviewComment, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)

	if input.AuthorLabel == "" {
		input.AuthorLabel = "You"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO review_comments (
			id, review_id, file_path, diff_section, side, line_number,
			author_label, body_html, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, commentID, reviewID, input.FilePath, input.DiffSection, input.Side, input.LineNumber, input.AuthorLabel, input.BodyHTML, now, now)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = s.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_added", FilePath: input.FilePath, Message: fmt.Sprintf("Commented on line %d", input.LineNumber)})

	return ReviewComment{
		ID: commentID, ReviewID: reviewID, FilePath: input.FilePath, DiffSection: input.DiffSection,
		Side: input.Side, LineNumber: input.LineNumber, AuthorLabel: input.AuthorLabel,
		BodyHTML: input.BodyHTML, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Store) UpdateReviewComment(ctx context.Context, reviewID string, commentID string, bodyHTML string) (ReviewComment, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		UPDATE review_comments
		SET body_html = ?, updated_at = ?, deleted_at = NULL
		WHERE id = ? AND review_id = ?
	`, bodyHTML, now, commentID, reviewID)

	if err != nil {
		return ReviewComment{}, err
	}

	comment, err := s.GetReviewComment(ctx, reviewID, commentID)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = s.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_edited", FilePath: comment.FilePath, Message: fmt.Sprintf("Edited comment on line %d", comment.LineNumber)})

	return comment, nil
}

func (s *Store) DeleteReviewComment(ctx context.Context, reviewID string, commentID string) error {
	now := s.now().UTC().Format(time.RFC3339Nano)
	comment, _ := s.GetReviewComment(ctx, reviewID, commentID)
	_, err := s.db.ExecContext(ctx, `
		UPDATE review_comments
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND review_id = ?
	`, now, now, commentID, reviewID)

	if err != nil {
		return err
	}

	if comment.ID != "" {
		_, _ = s.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_deleted", FilePath: comment.FilePath, Message: fmt.Sprintf("Deleted comment on line %d", comment.LineNumber)})
	}

	return nil
}

func (s *Store) GetReviewComment(ctx context.Context, reviewID string, commentID string) (ReviewComment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, review_id, file_path, diff_section, side, line_number, author_label, body_html, created_at, updated_at, deleted_at
		FROM review_comments
		WHERE id = ? AND review_id = ?
	`, commentID, reviewID)

	return scanComment(row)
}

func (s *Store) ListReviewComments(ctx context.Context, reviewID string) ([]ReviewComment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, review_id, file_path, diff_section, side, line_number, author_label, body_html, created_at, updated_at, deleted_at
		FROM review_comments
		WHERE review_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, reviewID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []ReviewComment{}

	for rows.Next() {
		comment, err := scanComment(rows)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func (s *Store) ListRepositories(ctx context.Context) ([]Repository, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT path, name, added_at, last_opened_at
		FROM repositories
		ORDER BY COALESCE(last_opened_at, added_at) DESC
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repos := []Repository{}

	for rows.Next() {
		var repo Repository

		var lastOpenedAt sql.NullString

		if err := rows.Scan(&repo.Path, &repo.Name, &repo.AddedAt, &lastOpenedAt); err != nil {
			return nil, err
		}

		repo.LastOpenedAt = fromNull(lastOpenedAt)
		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

func (s *Store) UpsertRepository(ctx context.Context, path string, name string) (Repository, error) {
	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	if name == "" {
		name = filepath.Base(path)
	}

	now := s.now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO repositories (path, name, added_at, last_opened_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			last_opened_at = excluded.last_opened_at
	`, path, name, now, now)

	if err != nil {
		return Repository{}, err
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT path, name, added_at, last_opened_at
		FROM repositories
		WHERE path = ?
	`, path)

	var repo Repository

	var lastOpenedAt sql.NullString

	if err := row.Scan(&repo.Path, &repo.Name, &repo.AddedAt, &lastOpenedAt); err != nil {
		return Repository{}, err
	}

	repo.LastOpenedAt = fromNull(lastOpenedAt)

	return repo, nil
}

func (s *Store) RemoveRepository(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("repository path is required")
	}

	_, err := s.db.ExecContext(ctx, `DELETE FROM repositories WHERE path = ?`, path)

	return err
}

func NewRecorder(store *Store, runID string, also func(domain.Event) error) *Recorder {
	return &Recorder{store: store, runID: runID, also: also}
}

func (r *Recorder) Emit(ctx context.Context, event domain.Event) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	r.seq++
	event.RunID = r.runID
	event.Seq = r.seq

	if err := r.store.InsertEvent(ctx, event); err != nil {
		return err
	}

	if r.also != nil {
		return r.also(event)
	}

	return nil
}

func runSummary(row db.WorkflowRun) RunSummary {
	return RunSummary{
		ID:                      row.ID,
		WorkflowID:              row.WorkflowID,
		WorkflowName:            row.WorkflowName,
		ConfirmationOptionID:    row.ConfirmationOptionID,
		ConfirmationOptionLabel: row.ConfirmationOptionLabel,
		Mode:                    row.Mode,
		Status:                  row.Status,
		StartedAt:               row.StartedAt,
		CompletedAt:             fromNull(row.CompletedAt),
		ErrorMessage:            fromNull(row.ErrorMessage),
	}
}

func eventRecord(row db.WorkflowEvent) EventRecord {
	return EventRecord{
		ID:        row.ID,
		RunID:     row.RunID,
		Seq:       row.Seq,
		Type:      row.EventType,
		PhaseID:   fromNull(row.PhaseID),
		PhaseName: fromNull(row.PhaseName),
		Status:    fromNull(row.Status),
		Message:   fromNull(row.Message),
		CreatedAt: row.CreatedAt,
	}
}

func scanReview(row scanner) (ReviewSession, error) {
	var review ReviewSession

	var completedAt sql.NullString

	if err := row.Scan(
		&review.ID,
		&review.RepoRoot,
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
	); err != nil {
		return ReviewSession{}, err
	}

	review.CompletedAt = fromNull(completedAt)

	return review, nil
}

func scanComment(row scanner) (ReviewComment, error) {
	var comment ReviewComment

	var deletedAt sql.NullString

	if err := row.Scan(
		&comment.ID,
		&comment.ReviewID,
		&comment.FilePath,
		&comment.DiffSection,
		&comment.Side,
		&comment.LineNumber,
		&comment.AuthorLabel,
		&comment.BodyHTML,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&deletedAt,
	); err != nil {
		return ReviewComment{}, err
	}

	comment.DeletedAt = fromNull(deletedAt)

	return comment, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func fromNull(value sql.NullString) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func isDuplicateColumn(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate column")
}
