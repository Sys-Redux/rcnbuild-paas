<!-- markdownlint-disable -->
# Development Roadmap

## Phase 0: Project Setup (Week 1)

### Goals
- Initialize project structure
- Set up development environment
- Create basic infrastructure with Docker Compose

### Tasks
- [ ] Initialize Go module for API
- [ ] Initialize Next.js project for dashboard
- [ ] Create Docker Compose for local development
  - PostgreSQL
  - Redis
  - Traefik
  - MinIO (optional for MVP)
- [ ] Set up database migrations (golang-migrate)
- [ ] Create Makefile for common tasks
- [ ] Set up basic CI (GitHub Actions)

### Deliverables
```
/rcnbuild
  /cmd
    /api          # API server entrypoint
    /worker       # Build worker entrypoint
  /internal
    /...          # Business logic
  /web            # Next.js dashboard
  /docker-compose.yml
  /Makefile
```

---

## Phase 1: Core MVP (Weeks 2-4)

### Week 2: Authentication & Projects

**Auth System**
- [ ] GitHub OAuth flow
- [ ] JWT token generation
- [ ] Protected API routes
- [ ] User model and storage

**Project Management**
- [ ] Create project from GitHub repo
- [ ] List user's projects
- [ ] Project settings (name, branch, commands)
- [ ] Environment variable storage (encrypted)

**Dashboard**
- [ ] Login with GitHub
- [ ] New project wizard
- [ ] Project list view
- [ ] Basic project settings page

### Week 3: Build System

**Build Pipeline**
- [ ] Job queue setup (Asynq)
- [ ] Build worker process
- [ ] Git clone functionality
- [ ] Runtime detection (Node.js, Python, Static)
- [ ] Docker image building
- [ ] Build log streaming (Redis pub/sub → WebSocket)

**Supported Runtimes**
- [ ] Node.js (npm/yarn)
- [ ] Python (pip)
- [ ] Static HTML

**Dashboard**
- [ ] Build logs viewer
- [ ] Build status indicators

### Week 4: Deployment & Routing

**Container Management**
- [ ] Start containers from built images
- [ ] Environment variable injection
- [ ] Port binding
- [ ] Container lifecycle (start/stop/restart)

**Routing**
- [ ] Traefik dynamic configuration
- [ ] Subdomain routing (`slug.rcnbuild.dev`)
- [ ] TLS certificate generation

**GitHub Integration**
- [ ] Webhook receiver
- [ ] Auto-deploy on push
- [ ] Webhook setup via GitHub API

**Dashboard**
- [ ] Deployment history
- [ ] Live deployment URL
- [ ] Manual deploy button
- [ ] Basic rollback

### Phase 1 Milestone ✅
User can:
1. Sign in with GitHub
2. Create a project from a repository
3. See automatic deployments on push
4. Access their app via HTTPS URL
5. View build logs
6. Roll back to previous deployment

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
