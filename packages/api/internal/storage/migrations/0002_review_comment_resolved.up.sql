ALTER TABLE review_comments ADD COLUMN resolved INTEGER NOT NULL DEFAULT 0;
ALTER TABLE review_comments ADD COLUMN resolved_at TEXT;
