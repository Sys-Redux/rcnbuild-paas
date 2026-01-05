<!-- markdownlint-disable -->
# Render Architecture & Features Research

## Overview

Render is a full-stack cloud platform designed for deploying web services, databases, background workers, and more. It emphasizes simplicity with Git-based workflows, zero-config SSL, and infrastructure-as-code via Blueprints.

## Service Types

### 1. Web Services
- Host apps in any language/framework (Node.js, Python, Django, FastAPI, Ruby, Go, Rust, Elixir, PHP)
- Automatic builds and deploys on Git push
- Unique `onrender.com` subdomain per service
- Custom domains supported
- Port binding required on `0.0.0.0` (default port: 10000)

### 2. Private Services
- Internal services not exposed to public internet
- Communicate via private network

### 3. Background Workers
- Long-running processes
- No HTTP interface

### 4. Cron Jobs
- Scheduled tasks
- Supports standard cron expressions

### 5. Databases
- **PostgreSQL** - Fully managed
- **Key Value (Redis)** - Caching and sessions

### 6. Static Sites
- CDN-backed delivery
- No compute instances needed

## Deployment Workflow

### Automatic Deploys
1. Link GitHub/GitLab/Bitbucket repo branch
2. Push/merge to linked branch triggers auto-deploy
3. Optional: Wait for CI checks to pass before deploying

**Auto-Deploy Options:**
| Option | Behavior |
|--------|----------|
| On Commit | Deploy immediately on push/merge |
| After CI Checks Pass | Wait for all CI checks to succeed |
| Off | Manual deploys only |

**Supported CI Integrations:**
- GitHub Actions
- GitHub Checks API (CircleCI, etc.)
- GitLab CI/CD pipelines
- Bitbucket Pipelines

### Manual Deploys

**Methods:**
1. **Dashboard** - Deploy latest, specific commit, or clear cache & deploy
2. **CLI** - `render deploys create`
3. **Deploy Hooks** - HTTP GET/POST to unique URL
4. **API** - POST to Render API deploy endpoint

### Deploy Steps

```
Deploy initiated → Build Command → Pre-deploy Command (optional) → Start Command → Deploy complete
```

**Command Timeouts:**
| Command | Timeout |
|---------|---------|
| Build | 120 minutes |
| Pre-deploy | 30 minutes |
| Start | 15 minutes |

**Example Commands by Runtime:**

| Runtime | Build Command | Start Command |
|---------|--------------|---------------|
| Node.js | `npm install` | `npm start` |
| Python | `pip install -r requirements.txt` | `gunicorn app.wsgi` |
| Ruby | `bundle install` | `bundle exec puma` |
| Go | `go build -o app` | `./app` |
| Rust | `cargo build --release` | `cargo run --release` |
| Docker | N/A (uses Dockerfile) | CMD from Dockerfile |

### Pre-deploy Command
- Runs after build, before deploy
- Use for database migrations, CDN uploads
- Runs on separate instance (no filesystem access)
- Consumes pipeline minutes

## Zero-Downtime Deploys

**Sequence:**
1. Build new code version
2. Spin up NEW instance while original handles traffic
3. Health check new instance
4. Redirect traffic to new instance
5. Send `SIGTERM` to old instance (after 60 seconds)
6. Send `SIGKILL` if not terminated within shutdown delay (default 30s, max 300s)

**Key Points:**
- Original instance continues serving during new instance startup
- Failed deploys keep original instance running
- Multi-instance services: rolling update one instance at a time
- Persistent disks DISABLE zero-downtime deploys

## Infrastructure as Code: Blueprints

**Definition:** YAML configuration (`render.yaml`) that defines and manages multiple resources.

**Example Blueprint:**
```yaml
services:
  - type: web
    plan: free
    name: django-app
    runtime: python
    repo: https://github.com/render-examples/django.git
    buildCommand: './build.sh'
    startCommand: 'python -m gunicorn mysite.asgi:application -k uvicorn.workers.UvicornWorker'
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: django-app-db
          property: connectionString

databases:
  - name: django-app-db
    plan: free
```

**Blueprint Features:**
- Single source of truth for interconnected services
- Automatic redeploy on `render.yaml` changes
- Generate blueprints from existing services
- Replicate blueprints for multiple environments
- Auto-sync (enabled by default) or manual sync
- Deleting resource requires: remove from blueprint THEN delete in dashboard

**Supported Resources:**
- Web Services
- Private Services
- Background Workers
- Cron Jobs
- Databases (PostgreSQL, Key Value)
- Environment Groups

## Service Creation Flow (One-Click)

### From Git Repository:
1. Click **New > Web Service**
2. Select "Build and deploy from Git repository"
3. Connect GitHub/GitLab/Bitbucket account
4. Select repository → Connect
5. Configure:
   - Name (also used for subdomain)
   - Region
   - Branch
   - Language/Runtime
   - Build Command
   - Start Command
6. Select instance type (Free/Paid)
7. Optional: Environment variables, persistent disk, health checks
8. Click **Create Web Service**

### From Container Registry:
1. Click **New > Web Service**
2. Select "Deploy existing image from registry"
3. Enter image path (e.g., `docker.io/library/nginx:latest`)
4. Configure name and region
5. Select instance type
6. Click **Create Web Service**

## Key Features

| Feature | Description |
|---------|-------------|
| Zero-downtime deploys | Rolling updates with traffic switching |
| TLS Certificates | Free, fully-managed SSL |
| Custom Domains | Including wildcards |
| Auto/Manual Scaling | Horizontal instance scaling |
| Persistent Disks | Attached storage (disables zero-downtime) |
| Edge Caching | Static asset caching |
| WebSocket Support | Real-time connections |
| Service Previews | Preview environments |
| Instant Rollbacks | Revert to previous deploys |
| DDoS Protection | Built-in security |
| Blueprints (IaC) | Infrastructure as code |
| Private Network | Internal service communication |

## Ephemeral vs Persistent Filesystem

**Ephemeral (Default):**
- Filesystem changes lost on each deploy
- Standard for web services

**Persistence Options:**
- Render PostgreSQL / Key Value
- Custom datastores (MySQL, MongoDB)
- Persistent disks (with limitations)

## Deploy Management

**Overlapping Deploy Policies:**
| Policy | Behavior |
|--------|----------|
| Wait | Complete current deploy, then run latest triggered |
| Override | Cancel current deploy, start new one |

**Actions:**
- Cancel in-progress deploys
- Restart service (same commit, same env vars)
- Rollback to previous deploy
- Skip auto-deploy with commit message: `[skip render]`, `[render skip]`

## Build Filters (Monorepo Support)
- Only deploy when specific files change
- Useful for monorepo setups
