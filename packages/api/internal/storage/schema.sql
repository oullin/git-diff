
CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  os_username   TEXT NOT NULL UNIQUE,
  display_name  TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL DEFAULT '',
  created_at    TEXT NOT NULL,
  last_login_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS user_sessions (
  token        TEXT PRIMARY KEY,
  user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at   TEXT NOT NULL,
  expires_at   TEXT NOT NULL,
  last_used_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions(user_id);

CREATE TABLE IF NOT EXISTS ui_preferences (
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  key        TEXT NOT NULL,
  value      TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (user_id, key)
);

CREATE INDEX IF NOT EXISTS idx_ui_preferences_user ON ui_preferences(user_id);

CREATE TABLE IF NOT EXISTS review_sessions (
  id TEXT PRIMARY KEY,
  repo_root TEXT NOT NULL,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  branch TEXT NOT NULL,
  head_sha TEXT NOT NULL,
  status TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  files_changed INTEGER NOT NULL DEFAULT 0,
  additions INTEGER NOT NULL DEFAULT 0,
  deletions INTEGER NOT NULL DEFAULT 0,
  started_at TEXT NOT NULL,
  completed_at TEXT,
  context_kind TEXT NOT NULL DEFAULT 'working',
  context_sha TEXT
);

CREATE TABLE IF NOT EXISTS review_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  review_id TEXT NOT NULL REFERENCES review_sessions(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  file_path TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '',
  metadata TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_comments (
  id TEXT PRIMARY KEY,
  review_id TEXT NOT NULL REFERENCES review_sessions(id) ON DELETE CASCADE,
  file_path TEXT NOT NULL,
  diff_section TEXT NOT NULL,
  side TEXT NOT NULL,
  line_number INTEGER NOT NULL,
  author_label TEXT NOT NULL,
  body_html TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_review_sessions_started_at ON review_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_review_sessions_user ON review_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_review_events_review_id_created_at ON review_events(review_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_review_comments_review_id_file ON review_comments(review_id, file_path, line_number);

CREATE TABLE IF NOT EXISTS repositories (
  path TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  added_at TEXT NOT NULL,
  last_opened_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_repositories_last_opened ON repositories(last_opened_at DESC);
CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories(owner_id);

CREATE TABLE IF NOT EXISTS repository_users (
  repo_path  TEXT NOT NULL REFERENCES repositories(path) ON DELETE CASCADE,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role       TEXT NOT NULL CHECK (role IN ('write', 'read')),
  granted_at TEXT NOT NULL,
  PRIMARY KEY (repo_path, user_id)
);

CREATE INDEX IF NOT EXISTS idx_repository_users_user ON repository_users(user_id);

CREATE TABLE IF NOT EXISTS branches (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  repo_path    TEXT NOT NULL REFERENCES repositories(path) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  locked       INTEGER NOT NULL DEFAULT 0,
  locked_by    INTEGER REFERENCES users(id) ON DELETE SET NULL,
  locked_at    TEXT,
  last_seen_at TEXT NOT NULL,
  UNIQUE (repo_path, name)
);

CREATE INDEX IF NOT EXISTS idx_branches_repo ON branches(repo_path);
