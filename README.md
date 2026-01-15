<!-- markdownlint-disable -->
# RCNbuild

**One-click deployment platform for web applications.** Connect your GitHub repo, push code, and get a live HTTPS URL in seconds.

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-14+-black?logo=next.js)](https://nextjs.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

[![GitHub](https://img.shields.io/badge/GitHub-Sys--Redux-181717?logo=github)](https://github.com/Sys-Redux)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-t--edge-0A66C2?logo=linkedin)](https://www.linkedin.com/in/t-edge/)
[![Website](https://img.shields.io/badge/Website-sysredux.xyz-FF5722?logo=googlechrome&logoColor=white)](https://www.sysredux.xyz)
[![X](https://img.shields.io/badge/X-sys__redux-000000?logo=x)](https://x.com/sys_redux)
[![Discord](https://img.shields.io/badge/Discord-Join-5865F2?logo=discord&logoColor=white)](https://discord.gg/KdfApwrBuW)
[![Upwork](https://img.shields.io/badge/Upwork-Hire%20Me-6FDA44?logo=upwork&logoColor=white)](https://www.upwork.com/freelancers/~011b4cf7ebf1503859?mp_source=share)
[![Freelancer](https://img.shields.io/badge/Freelancer-Hire%20Me-29B2FE?logo=freelancer&logoColor=white)](https://www.freelancer.com/u/trevoredge?frm=trevoredge&sb=t)

---

## 🚀 What is RCNbuild?

**The self-hosted Vercel alternative** — built for indie developers who want control over their infrastructure without sacrificing developer experience.

*Not serverless — servers are good, actually.*

RCNbuild is an opinionated **Vercel/Render-inspired Platform-as-a-Service (PaaS)** that enables developers to deploy web applications with zero configuration. Simply connect your GitHub repository, and RCNbuild handles the rest:

```
Push to GitHub → Webhook triggers build → Docker container created → Live at myapp.rcnbuild.dev
```

### Why RCNbuild?

| | Vercel/Render | RCNbuild |
|---|---------------|----------|
| **Infrastructure** | Their cloud only | Any server you control |
| **Config visibility** | Hidden | "Show Config" on everything |
| **Minimum cost** | $20/mo per seat | $5 VPS + free software |
| **Lock-in** | High | Zero — export standard Docker |
| **Target user** | Teams with budget | Indie devs who want control |

### Core Principles

- 🏠 **Your Infrastructure** — Deploy to any server: $5 VPS, home lab, or enterprise cloud
- 🔓 **Zero Lock-in** — Everything is standard Docker + Traefik. Export and leave anytime
- 👁️ **Transparent by Default** — See every Dockerfile, every config, every decision
- 🎯 **Simple First** — One VPS runs the whole platform. Scale is opt-in
- 💻 **Developer-Owned** — Open source core. Fork it, modify it, own it

### Features

- 🔐 **GitHub OAuth** — Sign in with GitHub, access your repositories
- 🔍 **Auto-detect runtime** — Node.js, Python, static sites, or custom Dockerfile
- 🏗️ **Containerized builds** — Isolated Docker builds for every deployment
- 🌐 **Automatic HTTPS** — Every app gets a unique `*.rcnbuild.dev` subdomain with TLS
- 📊 **Real-time logs** — Stream build and deployment logs via WebSocket
- ⏪ **Instant rollback** — One-click rollback to any previous deployment
- 🔄 **Git-triggered deploys** — Push to deploy, automatic on every commit

---

## 📋 Prerequisites

- **Go 1.24+** — [Install Go](https://golang.org/dl/)
- **Docker & Docker Compose** — [Install Docker](https://docs.docker.com/get-docker/)
- **Node.js 20+** (for dashboard) — [Install Node.js](https://nodejs.org/)
- **golang-migrate** — `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Optional but recommended

- **air** (hot reload) — `go install github.com/air-verse/air@latest`
- **golangci-lint** — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

---

## 🛠️ Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/Sys-Redux/rcnbuild-paas.git
cd rcnbuild-paas
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and fill in the required values:

| Variable | Description | Where to get it |
|----------|-------------|-----------------|
| `POSTGRES_PASSWORD` | Database password | Choose a secure password |
| `GITHUB_CLIENT_ID` | OAuth App client ID | [GitHub Developer Settings](https://github.com/settings/developers) |
| `GITHUB_CLIENT_SECRET` | OAuth App secret | Same as above |
| `NGROK_AUTHTOKEN` | ngrok tunnel token | [ngrok Dashboard](https://dashboard.ngrok.com/get-started/your-authtoken) |
| `JWT_SECRET` | JWT signing key | Generate: `openssl rand -hex 32` |

### 3. Start infrastructure

```bash
make dev
```

This starts:
- **PostgreSQL** (port 5437) — Database
- **Redis** (port 6379) — Cache & job queue
- **Traefik** (port 80/443, dashboard 8080) — Reverse proxy
- **Registry** (port 5000) — Local Docker registry
- **ngrok** (port 4040) — HTTPS tunnel for GitHub webhooks

### 4. Get your ngrok URL

```bash
make ngrok-url
```

Copy the HTTPS URL and update your GitHub OAuth App callback URL to:
```
https://xxxx-xx-xx.ngrok-free.app/api/auth/github/callback
```

### 5. Run database migrations

```bash
make migrate-up
```

### 6. Start the API server

```bash
make api
```

The API will be available at `http://localhost:8080`

---

## 📁 Project Structure

```
rcnbuild-paas/
├── cmd/
│   ├── api/              # API server entry point
│   └── worker/           # Background job worker
├── internal/             # Private application code
│   ├── auth/             # GitHub OAuth, JWT handling
│   ├── github/           # GitHub API client, webhooks
│   ├── projects/         # Project CRUD operations
│   ├── builds/           # Build orchestration
│   ├── deploys/          # Deployment logic
│   ├── containers/       # Docker SDK interactions
│   ├── queue/            # Asynq job definitions
│   └── database/         # PostgreSQL queries
├── pkg/                  # Shared utilities
├── web/                  # Next.js dashboard
├── migrations/           # SQL migration files
├── docker-compose.yml    # Local dev infrastructure
├── Makefile              # Development commands
└── .env.example          # Environment template
```

---

## 🔧 Available Commands

Run `make help` for a full list:

```bash
# Infrastructure
make dev              # Start all services
make down             # Stop all services
make logs             # Tail container logs
make ps               # Show container status
make ngrok-url        # Get GitHub callback URL

# Application
make api              # Run API server (hot reload)
make worker           # Run background worker
make build            # Build production binaries

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback last migration
make migrate-create name=xyz   # Create new migration
make db-shell         # Open psql shell

# Development
make deps             # Download dependencies
make test             # Run tests
make lint             # Run linter
make clean            # Remove build artifacts
```

---

## 🌐 API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/auth/github` | Redirect to GitHub OAuth |
| `GET` | `/api/auth/github/callback` | OAuth callback handler |
| `POST` | `/api/auth/logout` | Clear session |
| `GET` | `/api/auth/me` | Get current user |

### Projects
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/projects` | List user's projects |
| `POST` | `/api/projects` | Create new project |
| `GET` | `/api/projects/:id` | Get project details |
| `PATCH` | `/api/projects/:id` | Update project |
| `DELETE` | `/api/projects/:id` | Delete project |

### Deployments
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/projects/:id/deployments` | List deployments |
| `POST` | `/api/projects/:id/deployments` | Trigger deploy |
| `GET` | `/api/deployments/:id` | Get deployment details |
| `GET` | `/api/deployments/:id/logs` | Stream logs (WebSocket) |
| `POST` | `/api/deployments/:id/rollback` | Rollback to this version |

### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/webhooks/github` | GitHub push events |

---

## 🏗️ Tech Stack

### Backend
- **Go** with **Gin** — HTTP framework
- **PostgreSQL** — Primary database
- **Redis** — Sessions, caching, job queue
- **Asynq** — Background job processing
- **Docker SDK** — Container management
- **zerolog** — Structured logging

### Frontend
- **Next.js 14+** — App Router
- **Tailwind CSS** — Styling
- **TanStack Query** — Server state
- **WebSocket** — Real-time logs

### Infrastructure
- **Docker Compose** — Local orchestration
- **Traefik** — Dynamic reverse proxy with auto-TLS
- **Local Registry** — Container image storage
- **ngrok** — Development tunnels

---

## 🔐 Security

- All sensitive environment variables encrypted at rest
- GitHub webhook signatures validated
- JWT tokens for API authentication
- Containers run as non-root users
- Resource limits enforced on user containers
- HTTPS everywhere via Traefik + Let's Encrypt

---

## 📊 Deployment States

```
PENDING → BUILDING → DEPLOYING → LIVE
              ↓           ↓
           FAILED      FAILED

CANCELLED (any stage before LIVE)
```

---

## 🗺️ Roadmap

### Phase 1: MVP *(In Progress — 40%)*
- [x] Project scaffolding
- [x] Infrastructure setup (Docker Compose)
- [x] GitHub OAuth login
- [x] Database schema (users, projects, deployments, env_vars)
- [x] AES-256-GCM encryption for secrets at rest
- [ ] List GitHub repositories
- [ ] Project API endpoints
- [ ] Auto-detect runtime
- [ ] Build in Docker container
- [ ] Deploy with Traefik routing
- [ ] Stream build logs

### Phase 2: Production Ready
- [ ] Custom domains
- [ ] Preview deployments (per PR)
- [ ] Team collaboration
- [ ] Usage metrics & analytics

### Phase 3: Enterprise
- [ ] Kubernetes orchestration
- [ ] Multi-region deployments
- [ ] SSO / SAML authentication
- [ ] Audit logs

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- **Project Board**: [GitHub Project](https://github.com/users/Sys-Redux/projects/3)
- **Issues**: [GitHub Issues](https://github.com/Sys-Redux/rcnbuild-paas/issues)

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/Sys-Redux">Sys-Redux</a>
</p>
