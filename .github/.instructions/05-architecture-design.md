<!-- markdownlint-disable -->
# Architecture Design

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              USER INTERACTION                                │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │  Dashboard  │    │  CLI Tool   │    │  REST API   │    │  Webhooks   │  │
│  │  (Next.js)  │    │   (Go CLI)  │    │    (Go)     │    │  (GitHub)   │  │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘  │
└─────────┼──────────────────┼──────────────────┼──────────────────┼─────────┘
          │                  │                  │                  │
          └──────────────────┴────────┬─────────┴──────────────────┘
                                      │
┌─────────────────────────────────────┼───────────────────────────────────────┐
│                              CONTROL PLANE                                   │
├─────────────────────────────────────┼───────────────────────────────────────┤
│                                     ▼                                        │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         API Gateway (Traefik)                         │   │
│  │                    - Authentication / Rate Limiting                   │   │
│  └───────────────────────────────┬──────────────────────────────────────┘   │
│                                  │                                           │
│  ┌───────────────────────────────┼──────────────────────────────────────┐   │
│  │                         API Server (Go)                               │   │
│  │  ┌─────────────┬─────────────┬─────────────┬─────────────────────┐   │   │
│  │  │   Auth      │   Projects  │   Deploys   │   GitHub Integration│   │   │
│  │  │   Service   │   Service   │   Service   │   Service           │   │   │
│  │  └─────────────┴─────────────┴─────────────┴─────────────────────┘   │   │
│  └───────────────────────────────┬──────────────────────────────────────┘   │
│                                  │                                           │
│  ┌──────────────┬────────────────┼────────────────┬─────────────────────┐   │
│  │              │                │                │                      │   │
│  ▼              ▼                ▼                ▼                      │   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────────────┐    │   │
│  │PostgreSQL│ │  Redis   │ │  MinIO   │ │     Job Queue (Asynq)    │    │   │
│  │(metadata)│ │ (cache)  │ │(artifacts)│ │ - Build Jobs             │    │   │
│  └──────────┘ └──────────┘ └──────────┘ │ - Deploy Jobs            │    │   │
│                                          │ - Cleanup Jobs           │    │   │
│                                          └────────────┬─────────────┘    │   │
└───────────────────────────────────────────────────────┼─────────────────────┘
                                                        │
┌───────────────────────────────────────────────────────┼─────────────────────┐
│                              BUILD PLANE                                     │
├───────────────────────────────────────────────────────┼─────────────────────┤
│                                                       ▼                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                       Build Worker (Go + Docker)                       │  │
│  │  ┌────────────────┬────────────────┬────────────────────────────────┐ │  │
│  │  │ Clone Repo     │ Detect Runtime │ Execute Build                  │ │  │
│  │  │ (git)          │ (Buildpacks)   │ (Docker container)             │ │  │
│  │  └────────────────┴────────────────┴────────────────────────────────┘ │  │
│  │                                    │                                   │  │
│  │                                    ▼                                   │  │
│  │  ┌────────────────────────────────────────────────────────────────┐   │  │
│  │  │                    Build Container (ephemeral)                  │   │  │
│  │  │  - npm install / pip install                                    │   │  │
│  │  │  - npm run build / python setup                                 │   │  │
│  │  │  - Create Docker image                                          │   │  │
│  │  └────────────────────────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
│                                      │                                       │
│                                      ▼                                       │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                       Container Registry (Harbor)                       │ │
│  │                       - Store built images                              │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              RUNTIME PLANE                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                    Ingress Controller (Traefik)                         │ │
│  │  - Dynamic routing based on hostname                                    │ │
│  │  - TLS termination (Let's Encrypt)                                      │ │
│  │  - Load balancing                                                       │ │
│  └───────────────────────────────────┬────────────────────────────────────┘ │
│                                      │                                       │
│  ┌───────────────────────────────────┼────────────────────────────────────┐ │
│  │                    Container Runtime (Docker/Swarm/K8s)                 │ │
│  │                                   │                                     │ │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │ │
│  │  │ app-1.rcnbuild.dev│  │ app-2.rcnbuild.dev│  │ app-3.rcnbuild.dev│  ...   │ │
│  │  │  ┌───────────┐  │  │  ┌───────────┐  │  │  ┌───────────┐  │         │ │
│  │  │  │ Container │  │  │  │ Container │  │  │  │ Container │  │         │ │
│  │  │  │  :10000   │  │  │  │  :10000   │  │  │  │  :10000   │  │         │ │
│  │  │  └───────────┘  │  │  └───────────┘  │  │  └───────────┘  │         │ │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘         │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Descriptions

### 1. User Interaction Layer

**Dashboard (Next.js)**
- Web-based UI for project management
- Real-time build/deploy logs (WebSocket)
- Environment variable configuration
- Deployment history and rollback

**CLI Tool (Go)**
- `rcnbuild login` - Authenticate
- `rcnbuild deploy` - Manual deployment
- `rcnbuild logs` - Stream logs
- `rcnbuild env` - Manage environment variables

**REST API**
- All operations exposed via REST endpoints
- JWT authentication
- OpenAPI/Swagger documentation

**Webhooks**
- GitHub webhook receiver
- Validates signatures
- Triggers builds on push/PR events

### 2. Control Plane

**API Server (Go)**

Core services:
- **Auth Service**: GitHub OAuth, JWT tokens, sessions
- **Projects Service**: CRUD for projects, configuration
- **Deploys Service**: Deployment management, rollbacks
- **GitHub Service**: Repo access, webhook management

**PostgreSQL**
- Project metadata
- User accounts and teams
- Deployment history
- Build logs (references)

**Redis**
- Session cache
- Real-time pub/sub for logs
- Job queue backend (Asynq)

**MinIO (S3-compatible)**
- Build artifacts
- Log storage
- Source code archives

### 3. Build Plane

**Build Worker**
- Consumes jobs from queue
- Manages ephemeral build containers
- Streams logs to Redis pub/sub
- Pushes images to registry

**Build Process:**
```
1. Receive build job
2. Clone repository
3. Detect language/runtime
4. Generate Dockerfile (or use Buildpacks)
5. Build Docker image
6. Push to registry
7. Update deployment status
8. Trigger deploy job
```

### 4. Runtime Plane

**Ingress Controller (Traefik)**
- Routes `*.rcnbuild.dev` to containers
- Automatic TLS certificates
- Health checking
- Load balancing (if scaled)

**Container Runtime**

MVP: Docker Compose / Docker Swarm
- Simple deployment
- Good for single-node
- Easy to understand

Production: Kubernetes
- Multi-node scaling
- Advanced scheduling
- Rolling updates
- Self-healing

---

## Data Flow: One-Click Deployment

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ GitHub  │────▶│ Webhook │────▶│  Queue  │────▶│  Build  │────▶│ Deploy  │
│  Push   │     │ Handler │     │   Job   │     │ Worker  │     │ Worker  │
└─────────┘     └─────────┘     └─────────┘     └─────────┘     └─────────┘
                                                     │               │
                     ┌───────────────────────────────┘               │
                     ▼                                               ▼
              ┌─────────────┐                               ┌─────────────┐
              │  Container  │                               │   Traefik   │
              │  Registry   │                               │   Config    │
              └─────────────┘                               └─────────────┘
```

**Detailed Flow:**

1. **Push Event** → GitHub sends webhook to `/api/webhooks/github`
2. **Webhook Handler**:
   - Validates signature
   - Finds project by repo URL
   - Creates deployment record (status: pending)
   - Enqueues build job
3. **Build Job**:
   - Clones repo at specific commit
   - Runs build in isolated container
   - Creates Docker image with tag `registry/project:commit-sha`
   - Pushes image to registry
   - Updates deployment status → built
   - Enqueues deploy job
4. **Deploy Job**:
   - Pulls image from registry
   - Creates/updates container with new image
   - Configures Traefik routing
   - Performs health check
   - Updates deployment status → live
   - (If blue-green) Routes traffic to new container
   - (If blue-green) Terminates old container

---

## Database Schema (Core Tables)

```sql
-- Users (linked to GitHub)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    github_id BIGINT UNIQUE,
    github_username VARCHAR(255),
    email VARCHAR(255),
    access_token_encrypted TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR(255),
    slug VARCHAR(255) UNIQUE,  -- Used for subdomain
    repo_full_name VARCHAR(255),  -- owner/repo
    repo_url TEXT,
    branch VARCHAR(255) DEFAULT 'main',
    build_command TEXT,
    start_command TEXT,
    root_directory VARCHAR(255) DEFAULT '/',
    runtime VARCHAR(50),  -- nodejs, python, static
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Environment Variables
CREATE TABLE env_vars (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    key VARCHAR(255),
    value_encrypted TEXT,
    is_secret BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Deployments
CREATE TABLE deployments (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    commit_sha VARCHAR(40),
    commit_message TEXT,
    branch VARCHAR(255),
    status VARCHAR(50),  -- pending, building, deploying, live, failed, cancelled
    image_tag VARCHAR(255),
    url VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    deployed_at TIMESTAMP,
    logs_url TEXT
);

-- Indexes
CREATE INDEX idx_projects_user ON projects(user_id);
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_deployments_project ON deployments(project_id);
CREATE INDEX idx_deployments_status ON deployments(status);
```

---

## API Endpoints (Core)

```
# Authentication
POST   /api/auth/github          # GitHub OAuth callback
POST   /api/auth/logout          # Logout
GET    /api/auth/me              # Current user

# Projects
GET    /api/projects             # List projects
POST   /api/projects             # Create project
GET    /api/projects/:id         # Get project
PATCH  /api/projects/:id         # Update project
DELETE /api/projects/:id         # Delete project

# Deployments
GET    /api/projects/:id/deployments      # List deployments
POST   /api/projects/:id/deployments      # Trigger deployment
GET    /api/deployments/:id               # Get deployment
POST   /api/deployments/:id/rollback      # Rollback to this deployment
POST   /api/deployments/:id/cancel        # Cancel deployment
GET    /api/deployments/:id/logs          # Stream logs (WebSocket)

# Environment Variables
GET    /api/projects/:id/env              # List env vars
POST   /api/projects/:id/env              # Create env var
PATCH  /api/projects/:id/env/:key         # Update env var
DELETE /api/projects/:id/env/:key         # Delete env var

# Webhooks
POST   /api/webhooks/github               # GitHub webhook receiver
```

---

## Deployment States

```
PENDING ─────▶ BUILDING ─────▶ DEPLOYING ─────▶ LIVE
    │              │               │
    │              │               │
    ▼              ▼               ▼
CANCELLED      FAILED          FAILED
```

| State | Description |
|-------|-------------|
| `pending` | Deployment created, waiting for build worker |
| `building` | Build in progress |
| `deploying` | Build complete, starting container |
| `live` | Container running, receiving traffic |
| `failed` | Error during build or deploy |
| `cancelled` | Manually cancelled |
| `superseded` | Replaced by newer deployment |
