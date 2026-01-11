-- Projects table: stores user projects linked to GitHub repos
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,  -- Used for subdomain: slug.rcnbuild.dev
    repo_full_name VARCHAR(255) NOT NULL,  -- owner/repo format
    repo_url TEXT NOT NULL,
    branch VARCHAR(255) DEFAULT 'main',
    root_directory VARCHAR(255) DEFAULT '.',
    build_command TEXT,
    start_command TEXT,
    runtime VARCHAR(50),  -- nodejs, python, static, docker
    port INTEGER DEFAULT 3000,
    webhook_id BIGINT,  -- GitHub webhook ID
    webhook_secret TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_projects_repo ON projects(repo_full_name);