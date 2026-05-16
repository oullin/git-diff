CREATE TABLE IF NOT EXISTS workflow_runs (
  id TEXT PRIMARY KEY,
  workflow_id TEXT NOT NULL,
  workflow_name TEXT NOT NULL,
  confirmation_option_id TEXT NOT NULL,
  confirmation_option_label TEXT NOT NULL,
  mode TEXT NOT NULL,
  status TEXT NOT NULL,
  started_at TEXT NOT NULL,
  completed_at TEXT,
  error_message TEXT
);

CREATE TABLE IF NOT EXISTS workflow_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
  seq INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  phase_id TEXT,
  phase_name TEXT,
  status TEXT,
  message TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(run_id, seq)
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_started_at ON workflow_runs(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_workflow_events_run_id_seq ON workflow_events(run_id, seq);

CREATE TABLE IF NOT EXISTS user_preferences (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  theme TEXT NOT NULL DEFAULT 'light',
  diff_view_mode TEXT NOT NULL DEFAULT 'split',
  hide_whitespace INTEGER NOT NULL DEFAULT 0,
  last_repo_root TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_sessions (
  id TEXT PRIMARY KEY,
  repo_root TEXT NOT NULL,
  branch TEXT NOT NULL,
  head_sha TEXT NOT NULL,
  status TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  files_changed INTEGER NOT NULL DEFAULT 0,
  additions INTEGER NOT NULL DEFAULT 0,
  deletions INTEGER NOT NULL DEFAULT 0,
  started_at TEXT NOT NULL,
  completed_at TEXT
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
CREATE INDEX IF NOT EXISTS idx_review_events_review_id_created_at ON review_events(review_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_review_comments_review_id_file ON review_comments(review_id, file_path, line_number);
