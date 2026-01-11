<!-- markdownlint-disable -->
# Development Roadmap

> **Last Updated:** January 11, 2026

## Progress Summary

| Phase | Status | Completion |
|-------|--------|------------|
| Phase 0: Project Setup | ✅ Complete | 90% |
| Phase 1: Core MVP | 🚧 In Progress | 25% |
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
    /database         # ✅ PostgreSQL connection + user queries
  /web                # ⏳ Next.js dashboard (not started)
  /migrations         # ✅ SQL migrations
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
POST /api/webhooks/github       🚧 Webhook receiver (skeleton)
GET  /health                    ✅ Health check
```

**Database Schema Created:**
- [x] `users` table with GitHub OAuth fields
- [ ] `projects` table
- [ ] `deployments` table
- [ ] `env_vars` table

**Project Management** ⏳ NOT STARTED
- [ ] List GitHub repos via API
- [ ] Create project from GitHub repo
- [ ] List user's projects
- [ ] Project settings (name, branch, commands)
- [ ] Environment variable storage (encrypted)
- [ ] Webhook creation on GitHub repo

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
- [ ] Runtime detection (Node.js, Python, Static, Docker)
- [ ] Dockerfile generation
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

**GitHub Integration** 🚧 PARTIAL
- [x] Webhook receiver endpoint (structure only)
- [ ] Webhook signature validation (HMAC-SHA256)
- [ ] Parse push/PR events
- [ ] Auto-deploy on push to configured branch
- [ ] Webhook setup via GitHub API

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
| 2 | Create a project from a repository | ⏳ Not Started |
| 3 | See automatic deployments on push | ⏳ Not Started |
| 4 | Access app via HTTPS URL | ⏳ Not Started |
| 5 | View build logs | ⏳ Not Started |
| 6 | Roll back to previous deployment | ⏳ Not Started |

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

1. **Create Projects Database Schema** - Add `projects`, `deployments`, `env_vars` tables
2. **GitHub Repo Listing** - Implement `/api/github/repos` to list user's repositories
3. **Project CRUD** - Implement project creation and management endpoints
4. **Initialize Next.js Dashboard** - Set up the web frontend in `/web`
5. **Build Worker** - Implement Asynq job queue and build process
6. **Container Deployment** - Docker SDK integration for running containers
7. **Traefik Routing** - Dynamic subdomain configuration
