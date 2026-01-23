-- Rollback: Drop users table and indexes
DROP INDEX IF EXISTS idx_users_github_id;
DROP TABLE IF EXISTS users;
