<!-- markdownlint-disable -->
# Core Features to Implement

## Phase 1: Minimum Viable PaaS (Proof of Concept)

### 1. Git Integration & Webhooks
**Priority: Critical**

- [ ] GitHub OAuth integration
- [ ] Repository listing and selection
- [ ] Webhook receiver for push events
- [ ] Branch selection for deployment

**Flow:**
```
User connects GitHub → Selects repo → Configures branch → Webhook created
Push to branch → Webhook triggered → Deployment initiated
```

### 2. Build System
**Priority: Critical**

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
**Priority: Critical**

- [ ] Deployment queue management
- [ ] Build → Deploy workflow
- [ ] Deployment status tracking
- [ ] Deployment history
- [ ] Basic rollback capability

### 4. Container Orchestration
**Priority: Critical**

- [ ] Container creation from build artifacts
- [ ] Container lifecycle management
- [ ] Port binding and exposure
- [ ] Environment variable injection
- [ ] Basic health checks

### 5. Domain & Routing
**Priority: High**

- [ ] Unique subdomain per deployment (e.g., `app-name.rcnbuild.dev`)
- [ ] Reverse proxy / ingress routing
- [ ] HTTPS/TLS termination (Let's Encrypt)
- [ ] Custom domain support (future)

### 6. User Dashboard
**Priority: High**

- [ ] Project listing
- [ ] Deployment status and logs
- [ ] Environment variable management
- [ ] Basic settings (name, branch, commands)

---

## Phase 2: Enhanced Features

### 7. Preview Deployments
**Priority: Medium**

- [ ] Unique URL per commit/PR
- [ ] Automatic cleanup of old previews
- [ ] PR comments with preview links

### 8. Zero-Downtime Deploys
**Priority: Medium**

- [ ] Blue-green deployment strategy
- [ ] Health check verification before traffic switch
- [ ] Graceful shutdown handling (SIGTERM → SIGKILL)

### 9. Database Services
**Priority: Medium**

- [ ] Managed PostgreSQL instances
- [ ] Connection string injection
- [ ] Basic backup/restore

### 10. Infrastructure as Code
**Priority: Medium**

- [ ] YAML configuration file (`rcnbuild.yaml`)
- [ ] Multi-service definitions
- [ ] Environment groups
- [ ] Auto-sync on config changes

---

## Phase 3: Advanced Features

### 11. Serverless Functions
**Priority: Low (Phase 3)**

- [ ] Function deployment from API routes
- [ ] Auto-scaling to zero
- [ ] Cold start optimization

### 12. Edge Functions
**Priority: Low (Phase 3)**

- [ ] Middleware support
- [ ] Edge caching
- [ ] Geographic distribution

### 13. Background Workers & Cron Jobs
**Priority: Low (Phase 3)**

- [ ] Long-running process support
- [ ] Cron job scheduling
- [ ] Job queue integration

### 14. Scaling
**Priority: Low (Phase 3)**

- [ ] Manual horizontal scaling
- [ ] Auto-scaling based on metrics
- [ ] Load balancer integration

---

## Feature Comparison Matrix

| Feature | Vercel | Render | Our PaaS (Phase) |
|---------|--------|--------|------------------|
| Git integration | ✅ | ✅ | Phase 1 |
| Auto-deploy on push | ✅ | ✅ | Phase 1 |
| Build system | ✅ | ✅ | Phase 1 |
| Unique URLs | ✅ | ✅ | Phase 1 |
| Environment variables | ✅ | ✅ | Phase 1 |
| Preview deployments | ✅ | ✅ | Phase 2 |
| Zero-downtime deploy | ✅ | ✅ | Phase 2 |
| Managed databases | ❌ | ✅ | Phase 2 |
| IaC (YAML config) | ❌ | ✅ | Phase 2 |
| Serverless functions | ✅ | ❌ | Phase 3 |
| Edge functions | ✅ | ❌ | Phase 3 |
| Background workers | ❌ | ✅ | Phase 3 |
| Cron jobs | ✅ | ✅ | Phase 3 |
| Auto-scaling | ✅ | ✅ | Phase 3 |

---

## MVP Success Criteria

The proof-of-concept is successful when a user can:

1. ✅ Connect their GitHub account
2. ✅ Select a Node.js or Python repository
3. ✅ Configure build and start commands
4. ✅ Trigger an automatic deployment on git push
5. ✅ Access their app via a unique URL with HTTPS
6. ✅ View build/deployment logs
7. ✅ Set environment variables
8. ✅ Roll back to a previous deployment
