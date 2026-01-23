-- Rollback: Drop deployments table and indexes
DROP INDEX IF EXISTS idx_deployments_created_at;
DROP INDEX IF EXISTS idx_deployments_status;
DROP INDEX IF EXISTS idx_deployments_project_id;
DROP TABLE IF EXISTS deployments;
