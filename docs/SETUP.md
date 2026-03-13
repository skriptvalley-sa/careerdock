# Local Development Setup

Step-by-step guide to run CareerDock locally.

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) (LTS recommended) |
| Docker & Docker Compose | Latest | [docker.com](https://www.docker.com/products/docker-desktop/) |

## 1. Install Dev Tools

### Air (Go hot-reload)

```bash
go install github.com/air-verse/air@latest
```

### golangci-lint (Go linter)

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Verify `GOPATH/bin` is in your PATH

Both `air` and `golangci-lint` are installed to `$(go env GOPATH)/bin`. Ensure that directory is in your `PATH`:

```bash
# Check if it's already there
which air

# If not found, add to your shell profile (~/.zshrc or ~/.bashrc):
export PATH="$PATH:$(go env GOPATH)/bin"

# Then reload
source ~/.zshrc   # or source ~/.bashrc
```

### Frontend dependencies

```bash
cd frontend && npm install
```

## 2. Environment Setup

```bash
# Copy the example env file (must be at the project root)
cp .env.example .env

# The defaults work out of the box for local development.
# Only change values if you need custom ports or want to enable
# AI/payment integrations (Claude API key, Razorpay, etc.)
```

> **Note:** The `.env` file lives at the project root, not inside `backend/` or `frontend/`. The backend config automatically searches both the current directory and the parent directory.

## 3. Start Infrastructure

This starts PostgreSQL, Redis, MinIO (S3-compatible storage), and Mailhog (email testing):

```bash
make dev-infra
```

Verify services are healthy:

```bash
docker compose ps
```

| Service | Port | Purpose |
|---------|------|---------|
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Caching & job queues |
| MinIO API | 9000 | S3-compatible file storage |
| MinIO Console | 9001 | MinIO web UI (minioadmin/minioadmin) |
| Mailhog SMTP | 1025 | Email capture |
| Mailhog UI | 8025 | View captured emails |

## 4. Run Migrations & Seed Data

```bash
# Apply all database migrations (creates tables, indexes, etc.)
make migrate

# Seed the company directory (60 Indian tech companies)
make seed
```

## 5. Start Application Services

Run each in a **separate terminal** (or use a terminal multiplexer like tmux):

```bash
# Terminal 1 — Go API server (hot-reload on file changes)
make dev-api

# Terminal 2 — Asynq background worker (hot-reload)
make dev-worker

# Terminal 3 — Next.js frontend dev server
make dev-frontend
```

Once running:

| Service | URL |
|---------|-----|
| API | http://localhost:8080 |
| Health check | http://localhost:8080/api/health |
| Frontend | http://localhost:3000 |
| Mailhog UI | http://localhost:8025 |
| MinIO Console | http://localhost:9001 |

## Quick Start (TL;DR)

```bash
# One-time setup
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
cp .env.example .env
cd frontend && npm install && cd ..

# Start everything
make dev-infra
make migrate
make seed

# Then in separate terminals:
make dev-api
make dev-worker
make dev-frontend
```

## Makefile Reference

| Command | Description |
|---------|-------------|
| `make dev-infra` | Start Docker services (Postgres, Redis, MinIO, Mailhog) |
| `make dev-api` | Start Go API with hot-reload |
| `make dev-worker` | Start Asynq worker with hot-reload |
| `make dev-frontend` | Start Next.js dev server |
| `make migrate` | Run pending database migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-new NAME=xxx` | Create a new migration pair |
| `make seed` | Seed company data |
| `make test` | Run all tests (backend + frontend) |
| `make lint` | Run all linters |
| `make build` | Production build (backend + frontend) |
| `make clean` | Stop Docker services and remove volumes |

## Troubleshooting

### `air: command not found`

`air` is not in your PATH. Run:
```bash
go install github.com/air-verse/air@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### `golangci-lint: command not found`

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Docker containers won't start

Check if ports are already in use:
```bash
lsof -i :5432  # Postgres
lsof -i :6379  # Redis
lsof -i :9000  # MinIO
```

### Migration fails with connection error

Ensure Postgres is healthy before migrating:
```bash
docker compose ps   # postgres should show "healthy"
make migrate
```

### Frontend can't reach API

Check that `NEXT_PUBLIC_API_URL=http://localhost:8080` is set in `.env` and the API is running.
