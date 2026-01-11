<!-- markdownlint-disable -->
# Core Features to Implement

> **Last Updated:** January 11, 2026

## Implementation Status Overview

| Feature Category | Status | Progress |
|-----------------|--------|----------|
| Git Integration & Webhooks | 🚧 Partial | 40% |
| Build System | ⏳ Not Started | 0% |
| Deployment Pipeline | ⏳ Not Started | 0% |
| Container Orchestration | ⏳ Not Started | 0% |
| Domain & Routing | ⏳ Not Started | 0% |
| User Dashboard | ⏳ Not Started | 0% |

---

## Phase 1: Minimum Viable PaaS (Proof of Concept)

### 1. Git Integration & Webhooks
**Priority: Critical** | **Status: 🚧 Partial (40%)**

- [x] GitHub OAuth integration
- [ ] Repository listing and selection
- [x] Webhook receiver for push events (skeleton)
- [ ] Webhook signature validation
- [ ] Branch selection for deployment

**Flow:**
```
User connects GitHub → Selects repo → Configures branch → Webhook created
Push to branch → Webhook triggered → Deployment initiated
```

**Implemented:**
- GitHub OAuth flow with token exchange
- User authentication and session management
- Webhook endpoint structure (`POST /api/webhooks/github`)

**Remaining:**
- List user's repositories via GitHub API
- Create/manage webhooks on repositories
- Validate webhook signatures (HMAC-SHA256)

### 2. Build System
**Priority: Critical** | **Status: ⏳ Not Started**

- [ ] Container-based build environment
- [ ] Language/runtime detection
- [ ] Build command execution
- [ ] Build artifact storage
- [ ] Build logs streaming

**Supported Runtimes (Initial):**
- Node.js (npm, yarn, pnpm)
- Python (pip)
- Static sites (HTML/CSS/JS)

### 3. Deployment Pipeline
**Priority: Critical** | **Status: ⏳ Not Started**

- [ ] Deployment queue management (Asynq)
- [ ] Build → Deploy workflow
- [ ] Deployment status tracking
- [ ] Deployment history
- [ ] Basic rollback capability

### 4. Container Orchestration
**Priority: Critical** | **Status: ⏳ Not Started**

- [ ] Container creation from build artifacts
- [ ] Container lifecycle management
- [ ] Port binding and exposure
- [ ] Environment variable injection
- [ ] Basic health checks

**Infrastructure Ready:**
- Docker Compose environment configured
- Local Docker Registry running (port 5000)
- Traefik reverse proxy configured

### 5. Domain & Routing
**Priority: High** | **Status: ⏳ Not Started (Infrastructure Ready)**

- [ ] Unique subdomain per deployment (e.g., `app-name.rcnbuild.dev`)
- [ ] Reverse proxy / ingress routing
- [ ] HTTPS/TLS termination (Let's Encrypt)
- [ ] Custom domain support (future)

**Infrastructure Ready:**
- Traefik v3.0 configured with Docker provider
- HTTPS entrypoints configured
- Dynamic routing via Docker labels supported

### 6. User Dashboard
**Priority: High** | **Status: ⏳ Not Started**

- [ ] Project listing
- [ ] Deployment status and logs
- [ ] Environment variable management
- [ ] Basic settings (name, branch, commands)

---

## Phase 2: Enhanced Features

### 7. Preview Deployments
**Priority: Medium** | **Status: ⏳ Not Started**

- [ ] Unique URL per commit/PR
- [ ] Automatic cleanup of old previews
- [ ] PR comments with preview links

### 8. Zero-Downtime Deploys
**Priority: Medium** | **Status: ⏳ Not Started**

- [ ] Blue-green deployment strategy
- [ ] Health check verification before traffic switch
- [ ] Graceful shutdown handling (SIGTERM → SIGKILL)

### 9. Database Services
**Priority: Medium** | **Status: ⏳ Not Started**

- [ ] Managed PostgreSQL instances
- [ ] Connection string injection
- [ ] Basic backup/restore

### 10. Infrastructure as Code
**Priority: Medium** | **Status: ⏳ Not Started**

- [ ] YAML configuration file (`rcnbuild.yaml`)
- [ ] Multi-service definitions
- [ ] Environment groups
- [ ] Auto-sync on config changes

---

## Phase 3: Advanced Features

### 11. Serverless Functions
**Priority: Low (Phase 3)** | **Status: ⏳ Not Started**

- [ ] Function deployment from API routes
- [ ] Auto-scaling to zero
- [ ] Cold start optimization

### 12. Edge Functions
**Priority: Low (Phase 3)** | **Status: ⏳ Not Started**

- [ ] Middleware support
- [ ] Edge caching
- [ ] Geographic distribution

### 13. Background Workers & Cron Jobs
**Priority: Low (Phase 3)** | **Status: ⏳ Not Started**

- [ ] Long-running process support
- [ ] Cron job scheduling
- [ ] Job queue integration

### 14. Scaling
**Priority: Low (Phase 3)** | **Status: ⏳ Not Started**

- [ ] Manual horizontal scaling
- [ ] Auto-scaling based on metrics
- [ ] Load balancer integration

---

## Feature Comparison Matrix

| Feature | Vercel | Render | RCNbuild (Phase) | Status |
|---------|--------|--------|------------------|--------|
| Git integration | ✅ | ✅ | Phase 1 | 🚧 40% |
| Auto-deploy on push | ✅ | ✅ | Phase 1 | ⏳ |
| Build system | ✅ | ✅ | Phase 1 | ⏳ |
| Unique URLs | ✅ | ✅ | Phase 1 | ⏳ |
| Environment variables | ✅ | ✅ | Phase 1 | ⏳ |
| Preview deployments | ✅ | ✅ | Phase 2 | ⏳ |
| Zero-downtime deploy | ✅ | ✅ | Phase 2 | ⏳ |
| Managed databases | ❌ | ✅ | Phase 2 | ⏳ |
| IaC (YAML config) | ❌ | ✅ | Phase 2 | ⏳ |
| Serverless functions | ✅ | ❌ | Phase 3 | ⏳ |
| Edge functions | ✅ | ❌ | Phase 3 | ⏳ |
| Background workers | ❌ | ✅ | Phase 3 | ⏳ |
| Cron jobs | ✅ | ✅ | Phase 3 | ⏳ |
| Auto-scaling | ✅ | ✅ | Phase 3 | ⏳ |

---

## MVP Success Criteria

The proof-of-concept is successful when a user can:

| # | Requirement | Status |
|---|-------------|--------|
| 1 | Connect their GitHub account | ✅ Complete |
| 2 | Select a Node.js or Python repository | ⏳ Not Started |
| 3 | Configure build and start commands | ⏳ Not Started |
| 4 | Trigger an automatic deployment on git push | ⏳ Not Started |
| 5 | Access their app via a unique URL with HTTPS | ⏳ Not Started |
| 6 | View build/deployment logs | ⏳ Not Started |
| 7 | Set environment variables | ⏳ Not Started |
| 8 | Roll back to a previous deployment | ⏳ Not Started |

---

## What's Been Built So Far

### Authentication System ✅
- GitHub OAuth integration (login/callback)
- JWT token generation and validation
- Cookie-based session management
- Auth middleware for protected routes
- User CRUD operations in PostgreSQL

### Infrastructure ✅
- Docker Compose with all services
- PostgreSQL database with migrations
- Redis for cache/queue
- Traefik reverse proxy
- Local Docker registry
- ngrok for development OAuth callbacks

### API Endpoints ✅
```
GET  /health                    - Health check
GET  /api/auth/github           - GitHub OAuth redirect
GET  /api/auth/github/callback  - OAuth callback handler
POST /api/auth/logout           - Logout
GET  /api/auth/me               - Get current user (protected)
POST /api/webhooks/github       - Webhook receiver (skeleton)
```

### Database Schema ✅
```sql
-- users table (implemented)
- id (UUID)
- github_id (BIGINT)
- github_username (VARCHAR)
- email (VARCHAR)
- avatar_url (TEXT)
- access_token_encrypted (TEXT)
- created_at, updated_at (TIMESTAMPTZ)
```
