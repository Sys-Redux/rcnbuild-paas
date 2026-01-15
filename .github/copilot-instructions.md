# RCNbuild - Platform-as-a-Service Development Instructions

## Project Overview

**RCNbuild** is a Vercel/Render-inspired Platform-as-a-Service (PaaS) that enables one-click deployment of web applications. Users connect their GitHub repositories and RCNbuild automatically builds, deploys, and hosts their applications with unique HTTPS URLs.

## Core Concept

```
User pushes code to GitHub → RCNbuild receives webhook → Builds in container →
Deploys to Docker → Routes via Traefik → App live at app-name.rcnbuild.dev
```

---

## Design Philosophy: Ownership Through Transparency

RCNbuild is **the self-hosted Vercel alternative** — built for indie developers who want control over their infrastructure without sacrificing developer experience.

### Core Principles

**1. YOUR INFRASTRUCTURE**
Deploy to any server you control. We don't want your hardware, just your trust. Users can run RCNbuild on a $5 VPS, a home lab, or enterprise infrastructure.

**2. ZERO LOCK-IN**
Everything we generate is standard Docker + Traefik. Users can export and leave anytime. No proprietary formats, no hidden configurations.

**3. TRANSPARENT BY DEFAULT**
Show every Dockerfile, every config, every decision. No black boxes. Users should be able to learn infrastructure by using the platform. Consider adding "Show Config" buttons throughout the UI.

**4. SIMPLE UNTIL YOU NEED COMPLEX**
One VPS runs the whole platform. Complexity (clustering, multi-region, etc.) is opt-in, not required. Optimize for single-server deployments first.

**5. DEVELOPER-OWNED**
Open source core with Open Core + Managed Cloud business model. Fork it, modify it, own it. The deployment pipeline is the user's code.

### What We're NOT
- ❌ Not trying to replace Kubernetes
- ❌ Not targeting Fortune 500
- ❌ Not competing on enterprise features
- ❌ Not "serverless" — servers are good, actually

### What We ARE
- ✅ The best way to deploy to YOUR server
- ✅ A learning tool disguised as a platform
- ✅ The answer to "I want Vercel but cheaper and mine"
- ✅ The last deployment tool indie developers need

### How This Shapes Development Decisions

| Decision | Vercel/Render Way | RCNbuild Way |
|----------|-------------------|--------------|
| Infrastructure | Their cloud only | Any SSH-accessible server |
| Config visibility | Hidden | Exposed with "Show Config" |
| Default scale | Multi-region | Single server |
| Minimum viable setup | Managed service | One docker-compose.yml |
| Target user | Teams with budget | Indie devs who want control |

When making implementation decisions, always ask:
1. Does this respect the user's ownership of their infrastructure?
2. Is the configuration visible and exportable?
3. Does this work on a single $5 VPS?
4. Can the user understand what's happening?

---

## Tech Stack (Strictly Follow)

### Backend
| Component | Technology | Notes |
|-----------|------------|-------|
| API Server | **Go** with **Gin** framework | All backend logic |
| Database | **PostgreSQL** | User data, projects, deployments |
| Cache/Queue | **Redis** | Sessions, pub/sub, job queue |
| Job Queue | **Asynq** (Go library) | Build and deploy jobs |

### Infrastructure
| Component | Technology | Notes |
|-----------|------------|-------|
| Containers | **Docker** | Build isolation and app runtime |
| Orchestration | **Docker Compose** (MVP) | Later: Kubernetes |
| Reverse Proxy | **Traefik** | Dynamic routing, auto TLS |
| Registry | **Harbor** or Docker Registry | Store built images |
| Object Storage | **MinIO** (S3-compatible) | Build artifacts, logs |

### Frontend
| Component | Technology | Notes |
|-----------|------------|-------|
| Framework | **Next.js 14+** (App Router) | Dashboard UI |
| Styling | **Tailwind CSS** | Utility-first CSS |
| State | **TanStack Query** | Server state management |
| Real-time | **WebSockets** | Log streaming |

---

## Project Structure

```
/rcnbuild
├── /cmd
│   ├── /api              # API server main.go
│   └── /worker           # Build worker main.go
├── /internal
│   ├── /auth             # GitHub OAuth, JWT
│   ├── /github           # GitHub API, webhooks
│   ├── /projects         # Project CRUD
│   ├── /builds           # Build logic
│   ├── /deploys          # Deployment logic
│   ├── /containers       # Docker SDK interactions
│   ├── /queue            # Asynq job definitions
│   └── /database         # PostgreSQL models, queries
├── /pkg
│   └── /...              # Shared utilities
├── /web                  # Next.js dashboard
│   ├── /app              # App router pages
│   ├── /components       # React components
│   └── /lib              # API client, utils
├── /migrations           # SQL migrations
├── docker-compose.yml    # Local development
├── Makefile              # Build/run commands
└── rcnbuild.yaml         # Example IaC config
```

---

## Database Schema

### Core Tables

```sql
-- Users (linked to GitHub)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id BIGINT UNIQUE NOT NULL,
    github_username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    access_token_encrypted TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,  -- Used for subdomain
    repo_full_name VARCHAR(255) NOT NULL,  -- owner/repo
    repo_url TEXT NOT NULL,
    branch VARCHAR(255) DEFAULT 'main',
    root_directory VARCHAR(255) DEFAULT '.',
    build_command TEXT,
    start_command TEXT,
    runtime VARCHAR(50),  -- nodejs, python, static, docker
    port INTEGER DEFAULT 3000,
    webhook_id BIGINT,  -- GitHub webhook ID
    webhook_secret TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Environment Variables
CREATE TABLE env_vars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value_encrypted TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, key)
);

-- Deployments
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    commit_sha VARCHAR(40) NOT NULL,
    commit_message TEXT,
    commit_author VARCHAR(255),
    branch VARCHAR(255),
    status VARCHAR(50) NOT NULL,  -- pending, building, deploying, live, failed, cancelled
    image_tag VARCHAR(255),
    container_id VARCHAR(255),
    url VARCHAR(255),
    build_logs_url TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_deployments_project_id ON deployments(project_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_created_at ON deployments(created_at DESC);
```

---

## API Endpoints

### Authentication
```
GET  /api/auth/github           # Redirect to GitHub OAuth
GET  /api/auth/github/callback  # OAuth callback
POST /api/auth/logout           # Logout, clear session
GET  /api/auth/me               # Get current user
```

### Projects
```
GET    /api/projects             # List user's projects
POST   /api/projects             # Create new project
GET    /api/projects/:id         # Get project details
PATCH  /api/projects/:id         # Update project settings
DELETE /api/projects/:id         # Delete project and deployments
```

### Deployments
```
GET    /api/projects/:id/deployments       # List deployments
POST   /api/projects/:id/deployments       # Trigger manual deploy
GET    /api/deployments/:id                # Get deployment details
POST   /api/deployments/:id/cancel         # Cancel in-progress deploy
POST   /api/deployments/:id/rollback       # Rollback to this deployment
GET    /api/deployments/:id/logs           # Stream logs (WebSocket upgrade)
```

### Environment Variables
```
GET    /api/projects/:id/env               # List env vars (values hidden)
POST   /api/projects/:id/env               # Create/update env var
DELETE /api/projects/:id/env/:key          # Delete env var
```

### Webhooks
```
POST   /api/webhooks/github                # GitHub push/PR webhook
```

---

## Key Workflows

### 1. GitHub OAuth Flow
```
1. User clicks "Sign in with GitHub"
2. Redirect to GitHub OAuth authorize URL
3. GitHub redirects back with code
4. Exchange code for access token
5. Fetch user info from GitHub API
6. Create/update user in database
7. Generate JWT, set cookie
8. Redirect to dashboard
```

### 2. Create Project Flow
```
1. User selects GitHub repo from list
2. POST /api/projects with repo details
3. Detect runtime from repo files (package.json, requirements.txt, etc.)
4. Suggest build/start commands
5. Create webhook on GitHub repo
6. Store project in database
7. Optionally trigger first deployment
```

### 3. Deployment Flow (One-Click)
```
1. GitHub sends webhook on push
2. Validate webhook signature
3. Find project by repo URL
4. Create deployment record (status: pending)
5. Enqueue build job

BUILD WORKER:
6. Pull job from queue
7. Clone repo at commit SHA
8. Detect/confirm runtime
9. Run build command in Docker container
10. Build final Docker image
11. Push to registry
12. Update deployment (status: deploying)
13. Enqueue deploy job

DEPLOY WORKER:
14. Pull job from queue
15. Stop old container (if exists)
16. Start new container from image
17. Configure Traefik routing
18. Wait for health check
19. Update deployment (status: live)
20. Notify via WebSocket
```

### 4. Runtime Detection
```go
func DetectRuntime(repoPath string) string {
    if fileExists(repoPath, "package.json") {
        return "nodejs"
    }
    if fileExists(repoPath, "requirements.txt") ||
       fileExists(repoPath, "Pipfile") ||
       fileExists(repoPath, "pyproject.toml") {
        return "python"
    }
    if fileExists(repoPath, "Dockerfile") {
        return "docker"
    }
    if fileExists(repoPath, "index.html") {
        return "static"
    }
    return "unknown"
}
```

---

## Docker Configuration

### Traefik Labels for User Containers
```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.${PROJECT_SLUG}.rule=Host(`${PROJECT_SLUG}.rcnbuild.dev`)"
  - "traefik.http.routers.${PROJECT_SLUG}.tls=true"
  - "traefik.http.routers.${PROJECT_SLUG}.tls.certresolver=letsencrypt"
  - "traefik.http.services.${PROJECT_SLUG}.loadbalancer.server.port=${PORT}"
```

### Generated Dockerfiles

**Node.js:**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE ${PORT:-3000}
CMD ["npm", "start"]
```

**Python:**
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE ${PORT:-8000}
CMD ["gunicorn", "--bind", "0.0.0.0:${PORT:-8000}", "app:app"]
```

**Static:**
```dockerfile
FROM nginx:alpine
COPY . /usr/share/nginx/html
EXPOSE 80
```

---

## Environment Variables

### Platform Configuration
```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/rcnbuild

# Redis
REDIS_URL=redis://localhost:6379

# GitHub OAuth
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
GITHUB_REDIRECT_URI=https://dashboard.rcnbuild.dev/api/auth/github/callback

# JWT
JWT_SECRET=your-secret-key

# Domain
BASE_DOMAIN=rcnbuild.dev
DASHBOARD_URL=https://dashboard.rcnbuild.dev
API_URL=https://api.rcnbuild.dev

# Docker Registry
REGISTRY_URL=registry.rcnbuild.dev
REGISTRY_USERNAME=xxx
REGISTRY_PASSWORD=xxx

# Traefik
TRAEFIK_API_URL=http://traefik:8080
```

---

## Deployment States

```
PENDING → BUILDING → DEPLOYING → LIVE
              ↓           ↓
           FAILED      FAILED

CANCELLED (can happen at any stage before LIVE)
```

---

## Coding Guidelines

### Go Backend
- Use **Gin** for HTTP routing
- Use **sqlx** or **pgx** for database access
- Use **Asynq** for background jobs
- Use **Docker SDK for Go** for container management
- Follow standard Go project layout
- Use context for cancellation
- Return proper HTTP status codes
- Log with structured logging (zerolog or zap)

### Next.js Frontend
- Use **App Router** (not Pages Router)
- Use **Server Components** where possible
- Use **TanStack Query** for data fetching
- Use **Tailwind CSS** for styling
- Implement proper loading and error states
- Use WebSocket for real-time log streaming

### Security
- Encrypt sensitive env vars at rest
- Validate GitHub webhook signatures
- Use HTTPS everywhere
- Implement rate limiting
- Run containers as non-root
- Set resource limits on containers

---

## Development Commands

```bash
# Start all services
make dev

# Run API server only
make api

# Run worker only
make worker

# Run database migrations
make migrate-up

# Generate new migration
make migrate-create name=add_users_table

# Build Docker images
make build

# Run tests
make test
```

---

## Phase 1 MVP Checklist

1. [x] GitHub OAuth login
2. [ ] List GitHub repos
3. [ ] Create project from repo
4. [ ] Detect runtime automatically
5. [ ] Configure build/start commands
6. [x] Set environment variables (database layer + encryption)
7. [ ] Receive GitHub webhooks (endpoint exists, validation pending)
8. [ ] Build in Docker container
9. [ ] Push to container registry
10. [ ] Deploy container with Traefik routing
11. [ ] Generate HTTPS URL (slug.rcnbuild.dev)
12. [ ] Stream build logs via WebSocket
13. [ ] View deployment history
14. [ ] Rollback to previous deployment

---

## Reference Documentation

See `.github/.instructions/` for detailed research:
- `01-vercel-architecture.md` - Vercel infrastructure details
- `02-render-architecture.md` - Render infrastructure details
- `03-core-features-to-implement.md` - Feature prioritization
- `04-technologies-and-frameworks.md` - Tech stack details
- `05-architecture-design.md` - System architecture
- `06-development-roadmap.md` - Week-by-week plan