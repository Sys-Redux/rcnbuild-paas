<!-- markdownlint-disable -->
# Technologies & Frameworks

> **Last Updated:** January 11, 2026

## Implementation Status

| Component | Technology | Status |
|-----------|------------|--------|
| API Server | Go + Gin | ✅ Implemented |
| Database | PostgreSQL + pgx | ✅ Implemented |
| Cache/Queue | Redis | ✅ Running (not yet used in code) |
| Job Queue | Asynq | ⏳ Planned |
| JWT Auth | golang-jwt/jwt | ✅ Implemented |
| Logging | zerolog | ✅ Implemented |
| Container Runtime | Docker | ✅ Running |
| Reverse Proxy | Traefik v3.0 | ✅ Running |
| Container Registry | Docker Registry | ✅ Running |
| Dashboard | Next.js | ⏳ Not Started |

---

## Recommended Tech Stack

### Backend / Control Plane

| Component | Recommended Technology | Alternatives | Status |
|-----------|----------------------|--------------|--------|
| **API Server** | Go (Golang) | Node.js, Rust | ✅ Implemented |
| **Web Framework** | Gin or Fiber | Echo, Chi | ✅ Using Gin |
| **Database** | PostgreSQL | - | ✅ Implemented (pgx) |
| **Cache/Queue** | Redis | - | ✅ Running |
| **Job Queue** | Redis + Asynq | Bull (Node.js) | ⏳ Planned |

### Container & Orchestration

| Component | Recommended Technology | Alternatives | Status |
|-----------|----------------------|--------------|--------|
| **Container Runtime** | Docker | Podman | ✅ Running |
| **Orchestration (MVP)** | Docker Compose + Docker Swarm | - | ✅ Docker Compose |
| **Orchestration (Production)** | Kubernetes (K8s) | Nomad | ⏳ Future |
| **Container Registry** | Harbor | Docker Registry | ✅ Docker Registry |

### Build System

| Component | Recommended Technology | Alternatives | Rationale |
|-----------|----------------------|--------------|-----------|
| **Build Isolation** | Docker containers | Firecracker | Isolated build environments |
| **Language Detection** | Buildpacks (Paketo) | Custom detection | Auto-detect runtime, proven by Heroku/Render |
| **Build Artifacts** | S3-compatible storage (MinIO) | Local filesystem | Scalable object storage |

### Networking & Ingress

| Component | Recommended Technology | Alternatives | Status |
|-----------|----------------------|--------------|--------|
| **Reverse Proxy** | Traefik | Nginx, Caddy | ✅ Running (v3.0) |
| **TLS Certificates** | Let's Encrypt (via Traefik) | Caddy | ⏳ Configured |
| **DNS** | Cloudflare API | PowerDNS | ⏳ Not Started |

### Frontend / Dashboard

| Component | Recommended Technology | Alternatives | Status |
|-----------|----------------------|--------------|--------|
| **Framework** | Next.js (React) | SvelteKit, Nuxt | ⏳ Not Started |
| **Styling** | Tailwind CSS | - | ⏳ Not Started |
| **State Management** | TanStack Query | SWR | ⏳ Not Started |
| **Real-time** | WebSockets / SSE | - | ⏳ Not Started |

### Authentication & Authorization

| Component | Recommended Technology | Alternatives | Status |
|-----------|----------------------|--------------|--------|
| **Auth Provider** | GitHub OAuth | - | ✅ Implemented |
| **Session/Token** | JWT + Refresh Tokens | - | ✅ Implemented (JWT) |
| **RBAC** | Casbin | Custom | ⏳ Not Started |

### Observability

| Component | Recommended Technology | Alternatives | Rationale |
|-----------|----------------------|--------------|-----------|
| **Logging** | Loki + Grafana | ELK Stack | Lightweight, integrates well |
| **Metrics** | Prometheus + Grafana | - | Industry standard |
| **Tracing** | Jaeger | Tempo | Distributed tracing |

---

## Detailed Component Selection

### 1. API Server: Go with Gin

**Why Go:**
- Excellent performance and low memory footprint
- Built-in concurrency (goroutines)
- Single binary deployment
- Strong typing catches errors at compile time
- Used by Docker, Kubernetes, Terraform

**Example Structure:**
```
/cmd
  /api          # Main API server
  /worker       # Background job worker
  /builder      # Build runner
/internal
  /auth         # Authentication
  /github       # GitHub integration
  /builder      # Build logic
  /deployer     # Deployment logic
  /container    # Docker interactions
/pkg            # Shared libraries
```

### 2. Database: PostgreSQL

**Schema Concepts:**
```sql
-- Core entities
users, teams, projects, deployments, builds

-- Example: projects table
CREATE TABLE projects (
  id UUID PRIMARY KEY,
  name VARCHAR(255),
  repo_url TEXT,
  branch VARCHAR(255),
  build_command TEXT,
  start_command TEXT,
  env_vars JSONB,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

### 3. Build System: Docker + Buildpacks

**Buildpacks Approach:**
- Auto-detect language from source files
- Layer caching for fast rebuilds
- Consistent, reproducible builds
- Used by Heroku, Google Cloud Run, Render

**Alternative: Custom Dockerfile Generation:**
```dockerfile
# Generated for Node.js projects
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
CMD ["npm", "start"]
```

### 4. Reverse Proxy: Traefik

**Why Traefik:**
- Dynamic configuration via labels/API
- Automatic Let's Encrypt certificates
- Docker and Kubernetes native
- Dashboard for monitoring
- Middleware support (rate limiting, auth)

**Example Docker label configuration:**
```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.myapp.rule=Host(`myapp.rcnbuild.dev`)"
  - "traefik.http.routers.myapp.tls.certresolver=letsencrypt"
```

### 5. Job Queue: Redis + Asynq

**Why Asynq:**
- Go-native job queue
- Redis-backed (simple infrastructure)
- Supports delayed/scheduled jobs
- Retries with exponential backoff
- Web UI for monitoring

**Job Types:**
```go
const (
    TypeBuild    = "build"
    TypeDeploy   = "deploy"
    TypeCleanup  = "cleanup"
)
```

---

## Infrastructure Options

### Development / PoC Environment

```yaml
# docker-compose.yml approach
services:
  api:          # Go API server
  worker:       # Background job worker
  postgres:     # Database
  redis:        # Cache + Queue
  traefik:      # Reverse proxy
  registry:     # Private container registry (optional)
```

### Production Considerations

| Aspect | PoC Approach | Production Approach |
|--------|--------------|---------------------|
| Orchestration | Docker Compose | Kubernetes |
| Database | Single Postgres | Managed Postgres (RDS, Cloud SQL) |
| Storage | MinIO | S3/GCS |
| Secrets | Environment variables | Vault / Sealed Secrets |
| Networking | Single host | Multi-node with service mesh |

---

## Language/Runtime Support (MVP)

### Node.js Detection
```
package.json exists → Node.js project
  "engines.node" → version
  "scripts.build" → build command
  "scripts.start" → start command
```

### Python Detection
```
requirements.txt OR Pipfile OR pyproject.toml → Python project
  runtime.txt → version
  Procfile → start command
```

### Static Site Detection
```
index.html in root → Static site
  OR output of build (dist/, build/, public/)
```

---

## Security Considerations

1. **Build Isolation**
   - Run builds in ephemeral containers
   - No network access during build (optional)
   - Resource limits (CPU, memory, time)

2. **Runtime Isolation**
   - Separate networks per customer/project
   - Read-only filesystems where possible
   - Non-root container users

3. **Secrets Management**
   - Encrypt environment variables at rest
   - Never log secrets
   - Separate secret storage from config

4. **Network Security**
   - TLS everywhere
   - Rate limiting
   - DDoS protection (Cloudflare)
