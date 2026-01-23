-- Rollback: Drop env_vars table and indexes
DROP INDEX IF EXISTS idx_env_vars_project_id;
DROP TABLE IF EXISTS env_vars;
