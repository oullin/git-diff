CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  os_username   TEXT NOT NULL UNIQUE,
  display_name  TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL DEFAULT '',
  last_login_at TEXT NOT NULL DEFAULT '',
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_sessions (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  token        TEXT NOT NULL UNIQUE,
  user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at   TEXT NOT NULL,
  last_used_at TEXT NOT NULL,
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions(user_id);

CREATE TABLE IF NOT EXISTS user_preferences (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  key        TEXT NOT NULL,
  value      TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (user_id, key)
);

CREATE INDEX IF NOT EXISTS idx_user_preferences_user ON user_preferences(user_id);

CREATE TABLE IF NOT EXISTS repositories (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  path           TEXT NOT NULL UNIQUE,
  name           TEXT NOT NULL,
  owner_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  added_at       TEXT NOT NULL,
  last_opened_at TEXT,
  created_at     TEXT NOT NULL,
  updated_at     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_repositories_last_opened ON repositories(last_opened_at DESC);
CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories(owner_id);

CREATE TABLE IF NOT EXISTS repository_users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  repository_id INTEGER NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
  user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role          TEXT NOT NULL CHECK (role IN ('write', 'read')),
  granted_at    TEXT NOT NULL,
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL,
  UNIQUE (repository_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_repository_users_user ON repository_users(user_id);

CREATE TABLE IF NOT EXISTS repository_branches (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  repository_id INTEGER NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  locked        INTEGER NOT NULL DEFAULT 0,
  locked_by     INTEGER REFERENCES users(id) ON DELETE SET NULL,
  locked_at     TEXT,
  last_seen_at  TEXT NOT NULL,
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL,
  UNIQUE (repository_id, name)
);

CREATE INDEX IF NOT EXISTS idx_repository_branches_repo ON repository_branches(repository_id);

CREATE TABLE IF NOT EXISTS review_sessions (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  repo_root     TEXT NOT NULL,
  user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  branch        TEXT NOT NULL,
  head_sha      TEXT NOT NULL,
  status        TEXT NOT NULL,
  title         TEXT NOT NULL,
  summary       TEXT NOT NULL DEFAULT '',
  files_changed INTEGER NOT NULL DEFAULT 0,
  additions     INTEGER NOT NULL DEFAULT 0,
  deletions     INTEGER NOT NULL DEFAULT 0,
  started_at    TEXT NOT NULL,
  completed_at  TEXT,
  context_kind  TEXT NOT NULL DEFAULT 'working',
  context_sha   TEXT,
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_sessions_started_at ON review_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_review_sessions_user ON review_sessions(user_id);

CREATE TABLE IF NOT EXISTS review_events (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  review_id  INTEGER NOT NULL REFERENCES review_sessions(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  file_path  TEXT NOT NULL DEFAULT '',
  message    TEXT NOT NULL DEFAULT '',
  metadata   TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_events_review_id_created_at ON review_events(review_id, created_at ASC);

CREATE TABLE IF NOT EXISTS review_comments (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  review_id         INTEGER NOT NULL REFERENCES review_sessions(id) ON DELETE CASCADE,
  file_path         TEXT NOT NULL,
  diff_section      TEXT NOT NULL,
  side              TEXT NOT NULL,
  line_number       INTEGER NOT NULL,
  start_line_number INTEGER,
  start_side        TEXT,
  author_label      TEXT NOT NULL,
  body_html         TEXT NOT NULL,
  deleted_at        TEXT,
  created_at        TEXT NOT NULL,
  updated_at        TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_comments_review_id_file ON review_comments(review_id, file_path, line_number);

CREATE TABLE IF NOT EXISTS walkthroughs (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  repo_root    TEXT NOT NULL,
  context_kind TEXT NOT NULL,
  context_sha  TEXT NOT NULL DEFAULT '',
  fingerprint  TEXT NOT NULL,
  provider_id  TEXT NOT NULL DEFAULT '',
  model_id     TEXT NOT NULL,
  groups_json  TEXT NOT NULL DEFAULT '[]',
  summary      TEXT NOT NULL DEFAULT '',
  generated_at TEXT NOT NULL,
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL,
  UNIQUE (repo_root, context_kind, context_sha)
);

CREATE TABLE IF NOT EXISTS pending_comments (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id           INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  repo_root         TEXT NOT NULL,
  context_kind      TEXT NOT NULL DEFAULT 'working',
  context_sha       TEXT NOT NULL DEFAULT '',
  file_path         TEXT NOT NULL,
  diff_section      TEXT NOT NULL,
  side              TEXT NOT NULL,
  line_number       INTEGER NOT NULL,
  start_line_number INTEGER,
  start_side        TEXT,
  author_label      TEXT NOT NULL,
  body_html         TEXT NOT NULL,
  created_at        TEXT NOT NULL,
  updated_at        TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pending_comments_scope
  ON pending_comments(user_id, repo_root, context_kind, context_sha);
