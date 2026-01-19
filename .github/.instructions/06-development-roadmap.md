<!-- markdownlint-disable -->
# Development Roadmap

> **Last Updated:** January 18, 2026

## Progress Summary

| Phase | Status | Completion |
|-------|--------|------------|
| Phase 0: Project Setup | ✅ Complete | 100% |
| Phase 1: Core MVP | 🚧 In Progress | 60% |
| Phase 2: Enhanced Features | ⏳ Not Started | 0% |
| Phase 3: Advanced Features | ⏳ Not Started | 0% |

---

## Phase 0: Project Setup (Week 1) ✅ COMPLETED

### Goals
- Initialize project structure
- Set up development environment
- Create basic infrastructure with Docker Compose

### Tasks
- [x] Initialize Go module for API (`github.com/Sys-Redux/rcnbuild-paas`)
- [ ] Initialize Next.js project for dashboard *(deferred to Phase 1)*
- [x] Create Docker Compose for local development
  - [x] PostgreSQL (port 5437)
  - [x] Redis (port 6379)
  - [x] Traefik (ports 80, 443, 8080)
  - [x] Docker Registry (port 5000)
  - [x] ngrok tunnel for GitHub OAuth callbacks
- [x] Set up database migrations (golang-migrate)
- [x] Create Makefile for common tasks
- [ ] Set up basic CI (GitHub Actions) *(planned)*

### Deliverables ✅
```
/rcnbuild
  /cmd
    /api              # ✅ API server entrypoint (implemented)
    /worker           # ⏳ Build worker entrypoint (placeholder)
  /internal
    /auth             # ✅ JWT + middleware
    /database         # ✅ PostgreSQL connection + all models
    /github           # ✅ GitHub API client (repos, webhooks)
    /projects         # ✅ Project + EnvVar HTTP handlers
    /builds           # ✅ Runtime detection
  /pkg
    /crypto           # ✅ AES-256-GCM encryption
  /web                # ⏳ Next.js dashboard (not started)
  /migrations         # ✅ SQL migrations (users, projects, deployments, env_vars)
  /docker-compose.yml # ✅ Full dev infrastructure
  /Makefile           # ✅ Comprehensive dev commands
```

### Infrastructure Details

**Docker Compose Services:**
| Service | Image | Port | Status |
|---------|-------|------|--------|
| PostgreSQL | postgres:16-alpine | 5437 | ✅ Running |
| Redis | redis:7-alpine | 6379 | ✅ Running |
| Traefik | traefik:v3.0 | 80, 443, 8080 | ✅ Running |
| Registry | registry:2 | 5000 | ✅ Running |
| ngrok | ngrok/ngrok:latest | 4040 | ✅ Running |

**Go Dependencies Installed:**
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment loading
- `github.com/rs/zerolog` - Structured logging

---

## Phase 1: Core MVP (Weeks 2-4) 🚧 IN PROGRESS

### Week 2: Authentication & Projects

**Auth System** ✅ COMPLETED
- [x] GitHub OAuth flow (login redirect + callback)
- [x] JWT token generation (7-day expiry, HS256)
- [x] Protected API routes (`auth.AuthRequired()` middleware)
- [x] User model and storage (PostgreSQL with pgx)
- [x] Cookie-based session management (HTTP-only, SameSite=Lax)
- [x] Logout functionality

**API Endpoints Implemented:**
```
GET  /api/auth/github           ✅ Redirect to GitHub OAuth
GET  /api/auth/github/callback  ✅ Handle OAuth callback
POST /api/auth/logout           ✅ Clear session
GET  /api/auth/me               ✅ Get current user (protected)
GET  /api/repos                 ✅ List user's GitHub repos
GET  /api/projects              ✅ List user's projects
POST /api/projects              ✅ Create project from repo
GET  /api/projects/:id          ✅ Get project details
PATCH /api/projects/:id         ✅ Update project settings
DELETE /api/projects/:id        ✅ Delete project
GET  /api/projects/:id/env      ✅ List env vars (masked)
POST /api/projects/:id/env      ✅ Create/update env var
DELETE /api/projects/:id/env/:key ✅ Delete env var
POST /api/webhooks/github       🚧 Webhook receiver (skeleton)
GET  /health                    ✅ Health check
```

**Database Schema Created:**
- [x] `users` table with GitHub OAuth fields
- [x] `projects` table with full CRUD operations
- [x] `deployments` table with status management
- [x] `env_vars` table with encryption support

**Database Functions Implemented:**

*Users (`internal/database/users.go`):*
- [x] `CreateOrUpdateUser()` - Upsert with encrypted access token
- [x] `GetUserByID()` - Retrieve by UUID
- [x] `GetUserByGitHubID()` - Retrieve by GitHub ID
- [x] `DeleteUser()` - Remove user
- [x] `UpdateUserEmail()` - Update email
- [x] `GetUserAccessToken()` - Decrypt and return GitHub token

*Projects (`internal/database/projects.go`):*
- [x] `CreateProject()` - Insert new project
- [x] `GetProjectByID()` - Retrieve by UUID
- [x] `GetProjectBySlug()` - Retrieve by slug
- [x] `GetProjectByRepoFullName()` - Retrieve by repo name
- [x] `GetProjectsByUserID()` - List user's projects
- [x] `UpdateProject()` - Update project settings
- [x] `SetProjectWebhook()` - Store webhook credentials
- [x] `DeleteProject()` - Remove project
- [x] `SlugExists()` - Check slug availability

*Deployments (`internal/database/deployments.go`):*
- [x] `CreateDeployment()` - Create with pending status
- [x] `GetDeploymentByID()` - Retrieve by UUID
- [x] `GetDeploymentsByProjectID()` - List project deployments
- [x] `GetLiveDeployment()` - Get current live deployment
- [x] `UpdateDeploymentStatus()` - Update status with error
- [x] `StartDeploymentBuild()` - Mark as building
- [x] `SetDeploymentBuilt()` - Mark as built with image tag
- [x] `SetDeploymentLive()` - Mark as live with URL
- [x] `SupersededOldDeployments()` - Mark old deploys as superseded
- [x] `SetDeploymentFailed()` - Mark as failed with error
- [x] `CancelDeployment()` - Cancel in-progress deployment
- [x] `DeleteDeployment()` - Remove deployment
- [x] `DeleteDeploymentsByProjectID()` - Remove all for project

*Environment Variables (`internal/database/env_vars.go`):*
- [x] `CreateOrUpdateEnvVar()` - Upsert with encrypted value
- [x] `GetEnvVarsByProjectID()` - List project env vars
- [x] `DeleteEnvVar()` - Remove single env var
- [x] `DeleteAllEnvVar()` - Remove all for project
- [x] `GetEnvVarsAsMap()` - Decrypt for container injection
- [x] `ToDisplay()` / `ToDisplayList()` - Safe API response (masked)

*Crypto (`pkg/crypto/crypto.go`):*
- [x] `Encrypt()` - AES-256-GCM encryption with base64 output
- [x] `Decrypt()` - Decrypt base64 ciphertext

**GitHub API Client** ✅ COMPLETED (`internal/github/client.go`)
- [x] `NewClient()` - Create authenticated GitHub client
- [x] `ListUserRepos()` - Fetch repos with push access
- [x] `GetRepo()` - Fetch specific repository
- [x] `GetRepoContents()` - List directory contents
- [x] `FileExists()` - Check if file exists in repo
- [x] `CreateWebhook()` - Register webhook on repo
- [x] `DeleteWebhook()` - Remove webhook from repo
- [x] `GenerateWebhookSecret()` - Crypto-secure secret generation
- [x] `ParseRepoFullName()` - Parse "owner/repo" format

**Runtime Detection** ✅ COMPLETED (`internal/builds/runtime.go`)
- [x] `DetectRuntime()` - Analyze repo to detect runtime
- [x] Node.js detection (package.json, lock files)
- [x] Python detection (requirements.txt, pyproject.toml, Pipfile)
- [x] Go detection (go.mod)
- [x] Static site detection (index.html)
- [x] Docker detection (Dockerfile)
- [x] Package manager detection (npm, yarn, pnpm, bun)
- [x] Framework detection (Next.js, Vite)
- [x] `GetDockerfileForRuntime()` - Generate Dockerfile for runtime

**Project Management** ✅ COMPLETED
- [x] List GitHub repos via API (`GET /api/repos`)
- [x] Project database model and queries (complete)
- [x] Project API endpoints (all wired up)
- [x] Environment variable storage (encrypted)
- [x] Environment variable API endpoints
- [x] Webhook creation on GitHub repo
- [x] Runtime auto-detection on project creation
- [x] Slug generation and uniqueness validation

**Project Handlers** ✅ COMPLETED (`internal/projects/handlers.go`)
- [x] `HandleListRepos()` - List deployable GitHub repos
- [x] `HandleListProjects()` - List user's projects
- [x] `HandleCreateProject()` - Create project from repo
- [x] `HandleGetProject()` - Get project details + live deployment
- [x] `HandleUpdateProject()` - Update project settings
- [x] `HandleDeleteProject()` - Delete project + cleanup webhook

**Environment Variable Handlers** ✅ COMPLETED (`internal/projects/env_handlers.go`)
- [x] `HandleListEnvVars()` - List env vars (masked values)
- [x] `HandleCreateEnvVar()` - Create/update env var
- [x] `HandleDeleteEnvVar()` - Delete env var

**Dashboard** ⏳ NOT STARTED
- [ ] Initialize Next.js 14+ with App Router
- [ ] Set up Tailwind CSS
- [ ] Login with GitHub button
- [ ] New project wizard
- [ ] Project list view
- [ ] Basic project settings page

### Week 3: Build System

**Build Pipeline** ⏳ NOT STARTED
- [ ] Job queue setup (Asynq + Redis)
- [ ] Build worker process (`cmd/worker`)
- [ ] Git clone functionality
- [x] Runtime detection (Node.js, Python, Static, Docker) ✅ Implemented in `internal/builds/runtime.go`
- [x] Dockerfile generation ✅ Implemented in `internal/builds/runtime.go`
- [ ] Docker image building
- [ ] Push to local registry
- [ ] Build log streaming (Redis pub/sub → WebSocket)

**Supported Runtimes**
- [ ] Node.js (npm/yarn/pnpm)
- [ ] Python (pip/pipenv/poetry)
- [ ] Static HTML
- [ ] Custom Dockerfile

**Dashboard**
- [ ] Build logs viewer (WebSocket streaming)
- [ ] Build status indicators
- [ ] Build history

### Week 4: Deployment & Routing

**Container Management** ⏳ NOT STARTED
- [ ] Docker SDK for Go integration
- [ ] Start containers from built images
- [ ] Environment variable injection
- [ ] Port binding
- [ ] Container lifecycle (start/stop/restart)
- [ ] Health checks

**Routing** ⏳ NOT STARTED (Traefik infrastructure ready)
- [ ] Traefik dynamic configuration via Docker labels
- [ ] Subdomain routing (`slug.rcnbuild.dev`)
- [ ] TLS certificate generation (Let's Encrypt)

**GitHub Integration** ✅ MOSTLY COMPLETE
- [x] Webhook receiver endpoint (structure only)
- [ ] Webhook signature validation (HMAC-SHA256)
- [ ] Parse push/PR events
- [ ] Auto-deploy on push to configured branch
- [x] Webhook setup via GitHub API ✅ Creates webhook on project creation

**Dashboard**
- [ ] Deployment history view
- [ ] Live deployment URL display
- [ ] Manual deploy button
- [ ] Rollback to previous deployment
- [ ] Cancel in-progress deployment

### Phase 1 Milestone Checklist

| # | Requirement | Status |
|---|-------------|--------|
| 1 | Sign in with GitHub | ✅ Complete |
| 2 | Create a project from a repository | ✅ Complete (API + webhook setup) |
| 3 | See automatic deployments on push | 🚧 Webhook endpoint exists, validation pending |
| 4 | Access app via HTTPS URL | ⏳ Not Started |
| 5 | View build logs | ⏳ Not Started |
| 6 | Roll back to previous deployment | 🚧 Database ready, API pending |

---

## Phase 2: Enhanced Features (Weeks 5-7)

### Week 5: Preview Deployments & Domains

**Preview Environments**
- [ ] Deploy on pull request
- [ ] Unique URLs per PR (`pr-123-slug.rcnbuild.dev`)
- [ ] GitHub status checks
- [ ] Comment on PR with preview URL
- [ ] Auto-cleanup on PR close

**Custom Domains**
- [ ] Add custom domain to project
- [ ] DNS verification
- [ ] TLS certificate for custom domains

### Week 6: Zero-Downtime & Databases

**Zero-Downtime Deploys**
- [ ] Blue-green deployment strategy
- [ ] Health check endpoint detection
- [ ] Graceful traffic switching
- [ ] SIGTERM/SIGKILL handling

**Managed PostgreSQL**
- [ ] Create database for project
- [ ] Connection string injection
- [ ] Basic backup schedule

### Week 7: Infrastructure as Code

**rcnbuild.yaml**
- [ ] Parse YAML configuration
- [ ] Multi-service definitions
- [ ] Environment groups
- [ ] Auto-sync on config changes

**Example:**
```yaml
services:
  - name: web
    type: web
    runtime: nodejs
    buildCommand: npm install
    startCommand: npm start
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: mydb

databases:
  - name: mydb
    plan: starter
```

---

## Phase 3: Advanced Features (Weeks 8-10)

### Week 8: Background Workers & Cron

- [ ] Background worker service type
- [ ] Cron job service type
- [ ] Job scheduling
- [ ] Logs for background processes

### Week 9: Scaling & Metrics

- [ ] Manual horizontal scaling
- [ ] Resource metrics (CPU, memory)
- [ ] Basic auto-scaling rules
- [ ] Dashboard metrics view

### Week 10: Polish & Documentation

- [ ] API documentation (OpenAPI)
- [ ] User documentation
- [ ] CLI tool improvements
- [ ] Error handling improvements
- [ ] Rate limiting
- [ ] Audit logging

---

## Future Considerations (Post-MVP)

### Serverless Functions
- Function extraction from API routes
- Cold start optimization
- Scale to zero

### Edge Functions
- Middleware support
- Global edge deployment
- Edge caching

### Multi-Region
- Geographic deployment options
- Data replication
- Latency-based routing

### Team Features
- Organizations/Teams
- Role-based access control
- Shared environments
- Deployment approvals

### Enterprise
- SSO/SAML
- Audit logs
- Compliance certifications
- SLA guarantees

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Build security (code execution) | Isolated containers, resource limits, no network in build |
| Container escape | Security-focused base images, regular updates |
| Secrets exposure | Encryption at rest, audit logging |
| Resource exhaustion | Per-project limits, quotas |
| Single point of failure | Stateless services, database replication |

---

## Success Metrics

### Phase 1 Success
- [ ] Deploy a Node.js app in under 5 minutes
- [ ] Build time < 3 minutes for simple apps
- [ ] Zero manual infrastructure setup required
- [ ] Works with public GitHub repos

### Phase 2 Success
- [ ] Zero-downtime deploys work reliably
- [ ] Preview URLs work for PRs
- [ ] Database provisioning in < 1 minute

### Phase 3 Success
- [ ] Can scale to 10+ concurrent builds
- [ ] Can host 100+ applications
- [ ] Uptime > 99% for deployed apps

---

## Next Steps (Recommended Order)

1. ~~**Create Projects Database Schema**~~ ✅ Done - All tables and queries implemented
2. ~~**GitHub Repo Listing**~~ ✅ Done - `/api/repos` lists user's deployable repos
3. ~~**Project API Endpoints**~~ ✅ Done - Full CRUD wired up in `main.go`
4. ~~**Runtime Detection**~~ ✅ Done - Auto-detects Node.js, Python, Go, Static, Docker
5. ~~**Environment Variables API**~~ ✅ Done - Full CRUD with encryption
6. **GitHub Webhook Validation** - Implement HMAC-SHA256 signature verification
7. **Initialize Next.js Dashboard** - Set up the web frontend in `/web`
8. **Build Worker** - Implement Asynq job queue and build process
9. **Container Deployment** - Docker SDK integration for running containers
10. **Traefik Routing** - Dynamic subdomain configuration
