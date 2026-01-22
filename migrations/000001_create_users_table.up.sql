-- Users table: stores GitHub authenticated users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id BIGINT UNIQUE NOT NULL,
    github_username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    avatar_url TEXT,
    access_token_encrypted TEXT, -- GitHub access token (encrypted at rest)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for fast lookups by GitHub ID (for Oauth callback)
CREATE INDEX idx_users_github_id ON users(github_id);