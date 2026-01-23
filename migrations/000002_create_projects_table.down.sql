-- Rollback: Drop projects table and indexes
DROP INDEX IF EXISTS idx_projects_repo;
DROP INDEX IF EXISTS idx_projects_slug;
DROP INDEX IF EXISTS idx_projects_user_id;
DROP TABLE IF EXISTS projects;
