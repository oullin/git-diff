package storage

// Row types map 1:1 to the tables defined in migrations/0001_baseline.up.sql.
// They are the *only* layer that knows about column names and SQL nulls; every
// repo translates between these and the public domain types.
//
// No `gorm:"foreignKey/references/many2many"` tags on purpose: joins are
// written explicitly via Joins/Select to avoid the implicit per-row Preload
// queries that turn list endpoints into N+1.

type UserRow struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	OSUsername   string `gorm:"column:os_username;uniqueIndex"`
	DisplayName  string `gorm:"column:display_name"`
	PasswordHash string `gorm:"column:password_hash"`
	CreatedAt    string `gorm:"column:created_at"`
	LastLoginAt  string `gorm:"column:last_login_at"`
}

type UserSessionRow struct {
	Token      string `gorm:"column:token;primaryKey"`
	UserID     int64  `gorm:"column:user_id"`
	CreatedAt  string `gorm:"column:created_at"`
	ExpiresAt  string `gorm:"column:expires_at"`
	LastUsedAt string `gorm:"column:last_used_at"`
}

type UIPreferenceRow struct {
	UserID    int64  `gorm:"column:user_id;primaryKey"`
	Key       string `gorm:"column:key;primaryKey"`
	Value     string `gorm:"column:value"`
	UpdatedAt string `gorm:"column:updated_at"`
}

type ReviewSessionRow struct {
	ID           string  `gorm:"column:id;primaryKey"`
	RepoRoot     string  `gorm:"column:repo_root"`
	UserID       int64   `gorm:"column:user_id"`
	Branch       string  `gorm:"column:branch"`
	HeadSHA      string  `gorm:"column:head_sha"`
	Status       string  `gorm:"column:status"`
	Title        string  `gorm:"column:title"`
	Summary      string  `gorm:"column:summary"`
	FilesChanged int     `gorm:"column:files_changed"`
	Additions    int     `gorm:"column:additions"`
	Deletions    int     `gorm:"column:deletions"`
	StartedAt    string  `gorm:"column:started_at"`
	CompletedAt  *string `gorm:"column:completed_at"`
	ContextKind  string  `gorm:"column:context_kind"`
	ContextSHA   *string `gorm:"column:context_sha"`
}

type ReviewEventRow struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ReviewID  string `gorm:"column:review_id"`
	EventType string `gorm:"column:event_type"`
	FilePath  string `gorm:"column:file_path"`
	Message   string `gorm:"column:message"`
	Metadata  string `gorm:"column:metadata"`
	CreatedAt string `gorm:"column:created_at"`
}

type ReviewCommentRow struct {
	ID              string  `gorm:"column:id;primaryKey"`
	ReviewID        string  `gorm:"column:review_id"`
	FilePath        string  `gorm:"column:file_path"`
	DiffSection     string  `gorm:"column:diff_section"`
	Side            string  `gorm:"column:side"`
	LineNumber      int64   `gorm:"column:line_number"`
	StartLineNumber *int64  `gorm:"column:start_line_number"`
	StartSide       *string `gorm:"column:start_side"`
	AuthorLabel     string  `gorm:"column:author_label"`
	BodyHTML        string  `gorm:"column:body_html"`
	CreatedAt       string  `gorm:"column:created_at"`
	UpdatedAt       string  `gorm:"column:updated_at"`
	// String null, *not* gorm.DeletedAt: preserves the existing on-disk format
	// (RFC3339Nano string) so pre-GORM databases keep working.
	DeletedAt *string `gorm:"column:deleted_at"`
}

type WalkthroughRow struct {
	RepoRoot    string `gorm:"column:repo_root;primaryKey"`
	ContextKind string `gorm:"column:context_kind;primaryKey"`
	ContextSHA  string `gorm:"column:context_sha;primaryKey"`
	Fingerprint string `gorm:"column:fingerprint"`
	ProviderID  string `gorm:"column:provider_id"`
	ModelID     string `gorm:"column:model_id"`
	GroupsJSON  string `gorm:"column:groups_json"`
	Summary     string `gorm:"column:summary"`
	GeneratedAt string `gorm:"column:generated_at"`
}

type PendingCommentRow struct {
	ID              string  `gorm:"column:id;primaryKey"`
	UserID          int64   `gorm:"column:user_id"`
	RepoRoot        string  `gorm:"column:repo_root"`
	ContextKind     string  `gorm:"column:context_kind"`
	ContextSHA      string  `gorm:"column:context_sha"`
	FilePath        string  `gorm:"column:file_path"`
	DiffSection     string  `gorm:"column:diff_section"`
	Side            string  `gorm:"column:side"`
	LineNumber      int64   `gorm:"column:line_number"`
	StartLineNumber *int64  `gorm:"column:start_line_number"`
	StartSide       *string `gorm:"column:start_side"`
	AuthorLabel     string  `gorm:"column:author_label"`
	BodyHTML        string  `gorm:"column:body_html"`
	CreatedAt       string  `gorm:"column:created_at"`
	UpdatedAt       string  `gorm:"column:updated_at"`
}

type RepositoryRow struct {
	Path         string  `gorm:"column:path;primaryKey"`
	Name         string  `gorm:"column:name"`
	OwnerID      int64   `gorm:"column:owner_id"`
	AddedAt      string  `gorm:"column:added_at"`
	LastOpenedAt *string `gorm:"column:last_opened_at"`
}

type RepositoryUserRow struct {
	RepoPath  string `gorm:"column:repo_path;primaryKey"`
	UserID    int64  `gorm:"column:user_id;primaryKey"`
	Role      string `gorm:"column:role"`
	GrantedAt string `gorm:"column:granted_at"`
}

type BranchRow struct {
	ID         int64   `gorm:"column:id;primaryKey;autoIncrement"`
	RepoPath   string  `gorm:"column:repo_path"`
	Name       string  `gorm:"column:name"`
	Locked     int     `gorm:"column:locked"`
	LockedBy   *int64  `gorm:"column:locked_by"`
	LockedAt   *string `gorm:"column:locked_at"`
	LastSeenAt string  `gorm:"column:last_seen_at"`
}

func (UserRow) TableName() string { return "users" }

func (UserSessionRow) TableName() string { return "user_sessions" }

func (UIPreferenceRow) TableName() string { return "ui_preferences" }

func (ReviewSessionRow) TableName() string { return "review_sessions" }

func (ReviewEventRow) TableName() string { return "review_events" }

func (ReviewCommentRow) TableName() string { return "review_comments" }

func (WalkthroughRow) TableName() string { return "walkthroughs" }

func (PendingCommentRow) TableName() string { return "pending_comments" }

func (RepositoryRow) TableName() string { return "repositories" }

func (RepositoryUserRow) TableName() string { return "repository_users" }

func (BranchRow) TableName() string { return "branches" }
