package storage

// Row types map 1:1 to the tables defined in migrations/0001_baseline.up.sql.
// They are the *only* layer that knows about column names and SQL nulls; every
// repo translates between these and the public domain types.
//
// Associations are declared with `gorm:"foreignKey/references"` tags so the
// Go layer mirrors the FK constraints in the migration. They exist for
// documentation and for explicit `.Joins()` chains in repos — never use
// `Preload()` or `.Association()`; every join must stay explicit so list
// endpoints remain O(1) queries. The `repository_users` join table is modeled
// directly (not via `many2many:`) because it carries its own columns.

type UserRow struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	OSUsername   string `gorm:"column:os_username;uniqueIndex"`
	DisplayName  string `gorm:"column:display_name"`
	PasswordHash string `gorm:"column:password_hash"`
	LastLoginAt  string `gorm:"column:last_login_at"`
	CreatedAt    string `gorm:"column:created_at"`
	UpdatedAt    string `gorm:"column:updated_at"`

	Sessions          []UserSessionRow    `gorm:"foreignKey:UserID;references:ID"`
	Preferences       []UserPreferenceRow `gorm:"foreignKey:UserID;references:ID"`
	OwnedRepositories []RepositoryRow     `gorm:"foreignKey:OwnerID;references:ID"`
}

type UserSessionRow struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Token      string `gorm:"column:token;uniqueIndex"`
	UserID     int64  `gorm:"column:user_id"`
	ExpiresAt  string `gorm:"column:expires_at"`
	LastUsedAt string `gorm:"column:last_used_at"`
	CreatedAt  string `gorm:"column:created_at"`
	UpdatedAt  string `gorm:"column:updated_at"`

	User *UserRow `gorm:"foreignKey:UserID;references:ID"`
}

type UserPreferenceRow struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64  `gorm:"column:user_id;uniqueIndex:idx_user_preferences_user_key,priority:1"`
	Key       string `gorm:"column:key;uniqueIndex:idx_user_preferences_user_key,priority:2"`
	Value     string `gorm:"column:value"`
	CreatedAt string `gorm:"column:created_at"`
	UpdatedAt string `gorm:"column:updated_at"`

	User *UserRow `gorm:"foreignKey:UserID;references:ID"`
}

type RepositoryRow struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement"`
	Path         string  `gorm:"column:path;uniqueIndex"`
	Name         string  `gorm:"column:name"`
	OwnerID      int64   `gorm:"column:owner_id"`
	AddedAt      string  `gorm:"column:added_at"`
	LastOpenedAt *string `gorm:"column:last_opened_at"`
	CreatedAt    string  `gorm:"column:created_at"`
	UpdatedAt    string  `gorm:"column:updated_at"`

	Owner         *UserRow              `gorm:"foreignKey:OwnerID;references:ID"`
	Collaborators []RepositoryUserRow   `gorm:"foreignKey:RepositoryID;references:ID"`
	Branches      []RepositoryBranchRow `gorm:"foreignKey:RepositoryID;references:ID"`
}

type RepositoryUserRow struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	RepositoryID int64  `gorm:"column:repository_id;uniqueIndex:idx_repository_users_repo_user,priority:1"`
	UserID       int64  `gorm:"column:user_id;uniqueIndex:idx_repository_users_repo_user,priority:2"`
	Role         string `gorm:"column:role"`
	GrantedAt    string `gorm:"column:granted_at"`
	CreatedAt    string `gorm:"column:created_at"`
	UpdatedAt    string `gorm:"column:updated_at"`

	Repository *RepositoryRow `gorm:"foreignKey:RepositoryID;references:ID"`
	User       *UserRow       `gorm:"foreignKey:UserID;references:ID"`
}

type RepositoryBranchRow struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement"`
	RepositoryID int64   `gorm:"column:repository_id;uniqueIndex:idx_repository_branches_repo_name,priority:1"`
	Name         string  `gorm:"column:name;uniqueIndex:idx_repository_branches_repo_name,priority:2"`
	Locked       int     `gorm:"column:locked"`
	LockedBy     *int64  `gorm:"column:locked_by"`
	LockedAt     *string `gorm:"column:locked_at"`
	LastSeenAt   string  `gorm:"column:last_seen_at"`
	CreatedAt    string  `gorm:"column:created_at"`
	UpdatedAt    string  `gorm:"column:updated_at"`

	Repository *RepositoryRow `gorm:"foreignKey:RepositoryID;references:ID"`
	Locker     *UserRow       `gorm:"foreignKey:LockedBy;references:ID"`
}

type ReviewSessionRow struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement"`
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
	CreatedAt    string  `gorm:"column:created_at"`
	UpdatedAt    string  `gorm:"column:updated_at"`

	User     *UserRow           `gorm:"foreignKey:UserID;references:ID"`
	Events   []ReviewEventRow   `gorm:"foreignKey:ReviewID;references:ID"`
	Comments []ReviewCommentRow `gorm:"foreignKey:ReviewID;references:ID"`
}

type ReviewEventRow struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ReviewID  int64  `gorm:"column:review_id"`
	EventType string `gorm:"column:event_type"`
	FilePath  string `gorm:"column:file_path"`
	Message   string `gorm:"column:message"`
	Metadata  string `gorm:"column:metadata"`
	CreatedAt string `gorm:"column:created_at"`
	UpdatedAt string `gorm:"column:updated_at"`

	Session *ReviewSessionRow `gorm:"foreignKey:ReviewID;references:ID"`
}

type ReviewCommentRow struct {
	ID              int64   `gorm:"column:id;primaryKey;autoIncrement"`
	ReviewID        int64   `gorm:"column:review_id"`
	FilePath        string  `gorm:"column:file_path"`
	DiffSection     string  `gorm:"column:diff_section"`
	Side            string  `gorm:"column:side"`
	LineNumber      int64   `gorm:"column:line_number"`
	StartLineNumber *int64  `gorm:"column:start_line_number"`
	StartSide       *string `gorm:"column:start_side"`
	AuthorLabel     string  `gorm:"column:author_label"`
	BodyHTML        string  `gorm:"column:body_html"`
	// String null, *not* gorm.DeletedAt: preserves the existing on-disk format
	// (RFC3339Nano string) so pre-GORM databases keep working.
	DeletedAt *string `gorm:"column:deleted_at"`
	CreatedAt string  `gorm:"column:created_at"`
	UpdatedAt string  `gorm:"column:updated_at"`

	Session *ReviewSessionRow `gorm:"foreignKey:ReviewID;references:ID"`
}

type WalkthroughRow struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	RepoRoot    string `gorm:"column:repo_root;uniqueIndex:idx_walkthroughs_scope,priority:1"`
	ContextKind string `gorm:"column:context_kind;uniqueIndex:idx_walkthroughs_scope,priority:2"`
	ContextSHA  string `gorm:"column:context_sha;uniqueIndex:idx_walkthroughs_scope,priority:3"`
	Fingerprint string `gorm:"column:fingerprint"`
	ProviderID  string `gorm:"column:provider_id"`
	ModelID     string `gorm:"column:model_id"`
	GroupsJSON  string `gorm:"column:groups_json"`
	Summary     string `gorm:"column:summary"`
	GeneratedAt string `gorm:"column:generated_at"`
	CreatedAt   string `gorm:"column:created_at"`
	UpdatedAt   string `gorm:"column:updated_at"`
}

type PendingCommentRow struct {
	ID              int64   `gorm:"column:id;primaryKey;autoIncrement"`
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

	User *UserRow `gorm:"foreignKey:UserID;references:ID"`
}

func (UserRow) TableName() string { return "users" }

func (UserSessionRow) TableName() string { return "user_sessions" }

func (UserPreferenceRow) TableName() string { return "user_preferences" }

func (RepositoryRow) TableName() string { return "repositories" }

func (RepositoryUserRow) TableName() string { return "repository_users" }

func (RepositoryBranchRow) TableName() string { return "repository_branches" }

func (ReviewSessionRow) TableName() string { return "review_sessions" }

func (ReviewEventRow) TableName() string { return "review_events" }

func (ReviewCommentRow) TableName() string { return "review_comments" }

func (WalkthroughRow) TableName() string { return "walkthroughs" }

func (PendingCommentRow) TableName() string { return "pending_comments" }
