DROP INDEX IF EXISTS idx_pending_comments_scope;
DROP TABLE IF EXISTS pending_comments;

DROP TABLE IF EXISTS walkthroughs;

DROP INDEX IF EXISTS idx_review_comments_review_id_file;
DROP INDEX IF EXISTS idx_review_events_review_id_created_at;
DROP INDEX IF EXISTS idx_review_sessions_user;
DROP INDEX IF EXISTS idx_review_sessions_started_at;

DROP TABLE IF EXISTS review_comments;
DROP TABLE IF EXISTS review_events;
DROP TABLE IF EXISTS review_sessions;

DROP INDEX IF EXISTS idx_repository_branches_repo;
DROP TABLE IF EXISTS repository_branches;

DROP INDEX IF EXISTS idx_repository_users_user;
DROP TABLE IF EXISTS repository_users;

DROP INDEX IF EXISTS idx_repositories_owner;
DROP INDEX IF EXISTS idx_repositories_last_opened;
DROP TABLE IF EXISTS repositories;

DROP INDEX IF EXISTS idx_user_preferences_user;
DROP TABLE IF EXISTS user_preferences;

DROP INDEX IF EXISTS idx_user_sessions_user;
DROP TABLE IF EXISTS user_sessions;

DROP TABLE IF EXISTS users;
