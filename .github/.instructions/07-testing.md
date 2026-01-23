<!--markdownlint-disable-->
# RCNbuild - Testing & Database Migration Guide

This document provides step-by-step instructions for setting up the database, running migrations, and thoroughly testing all implemented features of RCNbuild.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Database Migrations](#database-migrations)
4. [Testing the API](#testing-the-api)
5. [Testing Authentication](#testing-authentication)
6. [Testing Projects](#testing-projects)
7. [Testing Environment Variables](#testing-environment-variables)
8. [Testing Webhooks](#testing-webhooks)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before testing, ensure you have:

- **Go 1.24+** installed
- **Docker & Docker Compose** installed
- **golang-migrate CLI** installed:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **curl** or **httpie** for API testing (or use Postman/Insomnia)
- A **GitHub OAuth App** or **GitHub App** configured (for auth testing)

---

## Initial Setup

### 1. Configure Environment Variables

Copy the example environment file and fill in your values:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```dotenv
# ===========================================
# Required Configuration
# ===========================================

# Database (use port 5437 since that's what docker-compose exposes)
DATABASE_URL=postgres://rcnbuild:YOUR_SECURE_PASSWORD@localhost:5437/rcnbuild?sslmode=disable
POSTGRES_USER=rcnbuild
POSTGRES_PASSWORD=YOUR_SECURE_PASSWORD  # REQUIRED - set a strong password
POSTGRES_DB=rcnbuild

# Redis
REDIS_URL=localhost:6379

# GitHub OAuth (from https://github.com/settings/developers)
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URI=http://localhost:8080/api/auth/github/callback

# Security Keys (generate with: openssl rand -hex 32)
JWT_SECRET=your_32_byte_hex_secret_here
ENCRYPTION_KEY=your_32_byte_hex_key_here

# Server
API_PORT=8081  # Use any port except 8080 (Traefik dashboard)
API_HOST=0.0.0.0

# URLs
DASHBOARD_URL=http://localhost:3000
API_URL=http://localhost:8081
```

> **Important**: `POSTGRES_PASSWORD` is required by docker-compose and must match the password in `DATABASE_URL`.

### 2. Start Infrastructure Services

```bash
make dev
```

This starts:
- PostgreSQL on `localhost:5437`
- Redis on `localhost:6379`
- Traefik on `localhost:8080` (dashboard)
- Docker Registry on `localhost:5000`
- ngrok tunnel (if configured)

Verify services are running:

```bash
make ps
```

Expected output:
```
NAME                  STATUS    PORTS
rcnbuild-postgres     running   0.0.0.0:5437->5432/tcp
rcnbuild-redis        running   0.0.0.0:6379->6379/tcp
rcnbuild-traefik      running   0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:8080->8080/tcp
rcnbuild-registry     running   0.0.0.0:5000->5000/tcp
```

---

## Database Migrations

### Understanding Migration Files

RCNbuild uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations. The migrations are located in `./migrations/`:

| File | Purpose |
|------|---------|
| `000001_create_users_table.up.sql` | Creates users table for GitHub OAuth |
| `000002_create_projects_table.up.sql` | Creates projects table for deployed apps |
| `000003_create_env_vars_table.up.sql` | Creates encrypted environment variables |
| `000004_create_deployments_table.up.sql` | Creates deployment history tracking |

### Running All Migrations

Apply all pending migrations:

```bash
make migrate-up
```

Expected output:
```
Applying migrations...
1/u create_users_table (XXms)
2/u create_projects_table (XXms)
3/u create_deployments_table (XXms)
4/u create_env_vars_table (XXms)
✓ Migrations applied
```

### Check Migration Status

```bash
make migrate-status
```

This shows the current version (should be `4` after running all migrations).

### Verify Tables Were Created

Connect to the database:

```bash
make db-shell
```

Then run:
```sql
-- List all tables
\dt

-- Expected output:
--              List of relations
--  Schema |       Name       | Type  |  Owner
-- --------+------------------+-------+----------
--  public | deployments      | table | rcnbuild
--  public | env_vars         | table | rcnbuild
--  public | projects         | table | rcnbuild
--  public | schema_migrations| table | rcnbuild
--  public | users            | table | rcnbuild

-- Describe users table
\d users

-- Describe projects table
\d projects

-- Describe env_vars table
\d env_vars

-- Describe deployments table
\d deployments

-- Exit psql
\q
```

### Rollback Migrations (If Needed)

Roll back the last migration:
```bash
make migrate-down
```

> **Note**: Migration down files (`*.down.sql`) are not included by default. You may need to create them for rollback functionality.

### Reset Database (Development Only)

To completely reset the database (drops all data):

```bash
make down-clean  # Stops containers and removes volumes
make dev         # Restarts containers
make migrate-up  # Re-applies all migrations
```

---

## Testing the API

### Start the API Server

In a new terminal:

```bash
make api
```

Expected output:
```
Starting RCNbuild API server
{"level":"info","addr":"0.0.0.0:${API_PORT}","message":"Starting RCNbuild API server"}
```

> **Note:** In all examples below, replace `${API_PORT}` with the port you configured in `.env` (default: `8081`).

### Health Check

```bash
curl http://localhost:${API_PORT}/health
```

Expected response:
```json
{"service":"rcnbuild-api","status":"ok"}
```

---

## Testing Authentication

### 1. GitHub OAuth Login Flow

Open in browser:
```
http://localhost:${API_PORT}/api/auth/github
```

This redirects to GitHub for authorization. After authorizing, you'll be redirected back to your `DASHBOARD_URL`.

**For testing without a frontend**, you can:
1. Capture the callback URL from the browser's network tab
2. The callback sets an `auth_token` cookie

### 2. Get Current User

After authenticating, test the `/me` endpoint:

```bash
# Replace AUTH_TOKEN with your actual JWT from the cookie
curl -H "Cookie: auth_token=YOUR_JWT_TOKEN" \
  http://localhost:${API_PORT}/api/auth/me
```

Expected response:
```json
{
  "id": "uuid-here",
  "github_id": 12345678,
  "github_username": "your-username",
  "email": "your@email.com",
  "avatar_url": "https://avatars.githubusercontent.com/u/12345678",
  "created_at": "2026-01-22T10:00:00Z",
  "updated_at": "2026-01-22T10:00:00Z"
}
```

### 3. Test Logout

```bash
curl -X POST -H "Cookie: auth_token=YOUR_JWT_TOKEN" \
  http://localhost:${API_PORT}/api/auth/logout
```

Expected response:
```json
{"message":"Logged out successfully"}
```

### 4. Test Unauthorized Access

```bash
curl http://localhost:${API_PORT}/api/auth/me
```

Expected response (401):
```json
{"error":"authorization header required"}
```

---

## Testing Projects

All project endpoints require authentication. Set up a variable for convenience:

```bash
export AUTH="Cookie: auth_token=YOUR_JWT_TOKEN"
```

### 1. List GitHub Repositories

```bash
curl -H "$AUTH" "http://localhost:${API_PORT}/api/repos?page=1&page_size=10"
```

Expected response:
```json
{
  "repos": [
    {
      "id": 123456,
      "name": "my-repo",
      "full_name": "username/my-repo",
      "description": "Description here",
      "private": false,
      "html_url": "https://github.com/username/my-repo",
      "default_branch": "main",
      "language": "JavaScript"
    }
  ],
  "page": 1
}
```

### 2. Create a Project

```bash
curl -X POST -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "repo_full_name": "username/my-repo",
    "name": "My Awesome App",
    "branch": "main"
  }' \
  http://localhost:${API_PORT}/api/projects
```

Expected response (201):
```json
{
  "project": {
    "id": "uuid-here",
    "user_id": "user-uuid",
    "name": "My Awesome App",
    "slug": "my-awesome-app",
    "repo_full_name": "username/my-repo",
    "repo_url": "https://github.com/username/my-repo",
    "branch": "main",
    "root_directory": ".",
    "runtime": "nodejs",
    "port": 3000,
    "created_at": "2026-01-22T10:00:00Z"
  },
  "runtime_info": {
    "runtime": "nodejs",
    "build_command": "npm install && npm run build",
    "start_command": "npm start",
    "port": 3000
  }
}
```

### 3. List User's Projects

```bash
curl -H "$AUTH" http://localhost:${API_PORT}/api/projects
```

Expected response:
```json
{
  "projects": [
    {
      "id": "uuid-here",
      "name": "My Awesome App",
      "slug": "my-awesome-app",
      "repo_full_name": "username/my-repo",
      "runtime": "nodejs",
      "port": 3000
    }
  ]
}
```

### 4. Get Project Details

```bash
curl -H "$AUTH" http://localhost:${API_PORT}/api/projects/PROJECT_UUID
```

Expected response:
```json
{
  "project": { ... },
  "latest_deployment": null
}
```

### 5. Update Project Settings

```bash
curl -X PATCH -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "name": "Updated App Name",
    "port": 8000,
    "build_command": "npm install && npm run build:prod"
  }' \
  http://localhost:${API_PORT}/api/projects/PROJECT_UUID
```

### 6. Delete a Project

```bash
curl -X DELETE -H "$AUTH" http://localhost:${API_PORT}/api/projects/PROJECT_UUID
```

Expected response:
```json
{"message":"Project deleted successfully"}
```

### 7. Test Authorization (Access Denied)

Try accessing another user's project:

```bash
curl -H "$AUTH" http://localhost:${API_PORT}/api/projects/OTHER_USER_PROJECT_UUID
```

Expected response (403):
```json
{"error":"Access denied"}
```

---

## Testing Environment Variables

### 1. Create/Update Environment Variable

```bash
curl -X POST -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "key": "DATABASE_URL",
    "value": "postgres://user:pass@localhost:5432/db"
  }' \
  http://localhost:${API_PORT}/api/projects/PROJECT_UUID/env
```

Expected response (201):
```json
{
  "id": "env-var-uuid",
  "key": "DATABASE_URL",
  "value": "••••••••",
  "created_at": "2026-01-22T10:00:00Z"
}
```

> **Note**: Values are always masked in API responses for security.

### 2. List Environment Variables

```bash
curl -H "$AUTH" http://localhost:${API_PORT}/api/projects/PROJECT_UUID/env
```

Expected response:
```json
{
  "env_vars": [
    {
      "id": "uuid",
      "key": "DATABASE_URL",
      "value": "••••••••",
      "created_at": "2026-01-22T10:00:00Z"
    }
  ]
}
```

### 3. Update an Existing Variable

Same endpoint as create - uses upsert:

```bash
curl -X POST -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "key": "DATABASE_URL",
    "value": "postgres://user:newpass@localhost:5432/db"
  }' \
  http://localhost:${API_PORT}/api/projects/PROJECT_UUID/env
```

### 4. Delete Environment Variable

```bash
curl -X DELETE -H "$AUTH" \
  http://localhost:${API_PORT}/api/projects/PROJECT_UUID/env/DATABASE_URL
```

Expected response:
```json
{"message":"env var deleted"}
```

### 5. Test Invalid Key Format

```bash
curl -X POST -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "key": "123_INVALID",
    "value": "test"
  }' \
  http://localhost:${API_PORT}/api/projects/PROJECT_UUID/env
```

Expected response (400):
```json
{"error":"invalid key format"}
```

> Keys must start with a letter and contain only letters, numbers, and underscores.

### 6. Verify Encryption in Database

```bash
make db-shell
```

```sql
SELECT key, value_encrypted FROM env_vars LIMIT 5;
```

You should see encrypted (base64-encoded) values, NOT plaintext.

---

## Testing Webhooks

### 1. Webhook Endpoint Exists

```bash
curl -X POST http://localhost:${API_PORT}/api/webhooks/github
```

Expected response (400 - no body):
```json
{"error":"Invalid request"}
```

### 2. Test with Mock Push Event

Create a test payload file `push_event.json`:

```json
{
  "ref": "refs/heads/main",
  "before": "0000000000000000000000000000000000000000",
  "after": "abc123def456789",
  "repository": {
    "full_name": "username/my-repo",
    "html_url": "https://github.com/username/my-repo"
  },
  "head_commit": {
    "id": "abc123def456789",
    "message": "Test commit",
    "author": {
      "name": "Test User"
    }
  },
  "pusher": {
    "name": "testuser"
  }
}
```

Test (without signature - will fail signature validation if project exists):

```bash
curl -X POST -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-GitHub-Delivery: test-123" \
  -d @push_event.json \
  http://localhost:${API_PORT}/api/webhooks/github
```

Expected responses:
- If no matching project: `{"message":"No associated project found"}`
- If project exists but invalid signature: `{"error":"Unauthorized"}`

### 3. Test with ngrok (Real GitHub Webhooks)

1. Get your ngrok URL:
   ```bash
   make ngrok-url
   ```

2. Update your GitHub App/OAuth callback and webhook URLs
3. Configure `API_URL` in `.env` to your ngrok URL
4. Push to your repo and monitor logs:
   ```bash
   make logs
   ```

---

## Comprehensive Test Checklist

Use this checklist to verify all features work correctly:

### Infrastructure
- [ ] `make dev` starts all services
- [ ] `make ps` shows all containers healthy
- [ ] PostgreSQL accessible on port 5437
- [ ] Redis accessible on port 6379

### Database
- [ ] `make migrate-up` completes without errors
- [ ] All 4 tables created (users, projects, env_vars, deployments)
- [ ] `make db-shell` connects successfully
- [ ] Indexes created correctly

### Health & Startup
- [ ] `make api` starts without errors
- [ ] `/health` returns `{"status":"ok"}`
- [ ] Database connection logged
- [ ] Redis connection logged

### Authentication
- [ ] `/api/auth/github` redirects to GitHub
- [ ] Callback creates/updates user in database
- [ ] JWT cookie is set after login
- [ ] `/api/auth/me` returns user info with valid token
- [ ] `/api/auth/me` returns 401 without token
- [ ] `/api/auth/logout` clears cookie

### Projects
- [ ] `/api/repos` lists GitHub repositories
- [ ] Creating project detects runtime correctly
- [ ] Creating project generates unique slug
- [ ] Creating project creates GitHub webhook
- [ ] Listing projects returns only user's projects
- [ ] Getting project returns project + latest deployment
- [ ] Updating project modifies fields correctly
- [ ] Deleting project removes from database
- [ ] Deleting project removes GitHub webhook
- [ ] Cannot access other users' projects (403)

### Environment Variables
- [ ] Creating env var encrypts value
- [ ] Listing env vars masks values
- [ ] Updating existing key uses upsert
- [ ] Deleting env var removes from database
- [ ] Invalid key format rejected
- [ ] Cannot access other project's env vars

### Webhooks
- [ ] Webhook endpoint receives POST
- [ ] Non-push events are ignored
- [ ] Unknown repos return gracefully
- [ ] Invalid signatures rejected (401)
- [ ] Wrong branch pushes are skipped
- [ ] Valid pushes create deployment records
- [ ] Build jobs are enqueued to Redis

---

## Troubleshooting

### Database Connection Failed

```
Failed to connect to database
```

**Solutions:**
1. Check if PostgreSQL is running: `make ps`
2. Verify `DATABASE_URL` uses port `5437` (not 5432)
3. Ensure password matches `POSTGRES_PASSWORD`

### Migration Errors

```
error: pq: relation "users" already exists
```

**Solution:** Database already has tables. Reset with:
```bash
make down-clean
make dev
make migrate-up
```

### "ENCRYPTION_KEY must be at least 32 bytes"

**Solution:** Generate a proper key:
```bash
openssl rand -hex 32
```
Add to `.env` as `ENCRYPTION_KEY`.

### GitHub OAuth Errors

```
Failed to exchange code for token
```

**Solutions:**
1. Verify `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`
2. Check `GITHUB_REDIRECT_URI` matches GitHub App settings
3. For local dev with ngrok, update callback URL

### Port Conflicts

```
listen tcp :8080: bind: address already in use
```

**Solution:** Traefik uses 8080. Set `API_PORT=8081` (or any other port) in `.env`.

### Redis Connection Failed

```
Failed to connect to Redis queue
```

**Solutions:**
1. Check Redis is running: `docker logs rcnbuild-redis`
2. Verify `REDIS_URL=localhost:6379` (no `redis://` prefix for Asynq)

### Webhook Signature Validation Failed

```
Invalid webhook signature
```

**Solutions:**
1. Ensure project has webhook secret stored
2. Verify GitHub is using the correct secret
3. Check the webhook was created successfully when project was created

---

## Next Steps

After verifying all tests pass:

1. **Build Worker Implementation** - Implement the build worker that processes queue jobs
2. **Container Deployment** - Implement Docker container creation and management
3. **Log Streaming** - Add WebSocket support for real-time build logs
4. **Frontend Dashboard** - Build the Next.js dashboard UI

See `06-development-roadmap.md` for the complete development timeline.
