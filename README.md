# CareerDock

Career intelligence platform for tech job seekers in India.

Browse 200+ Indian tech companies, track applications, get AI-powered ATS scores, and land your dream tech job.

## Features

- **Company Directory** — Browse and filter Indian tech companies by size, tech stack, compensation, and hiring status
- **Application Tracker** — Track job applications with status, dates, and notes (free tier)
- **AI Resume Analysis** — Upload resumes for AI-powered skill extraction and improvement suggestions
- **ATS Scoring** — Get general, company-specific, and job-specific ATS compatibility scores
- **AI Company Matching** — Receive curated company recommendations based on your resume profile
- **Offline Support** — Browse the company directory even without an internet connection

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Chi, PostgreSQL, Redis, Asynq |
| Frontend | Next.js 15 (App Router), React 19, TailwindCSS, TanStack Query |
| Storage | AWS S3 / MinIO |
| AI | Claude API (primary), OpenAI (fallback) |
| Payments | Razorpay |
| Infra | Docker, GitHub Actions CI |

## Quick Start

```bash
# Prerequisites: Go 1.25+, Node.js 20+, Docker

# Install dev tools
go install github.com/air-verse/air@latest
cp .env.example .env
cd frontend && npm install && cd ..

# Start infrastructure
make dev-infra
make migrate
make seed

# Run services (separate terminals)
make dev-api        # API on :8080
make dev-worker     # Background jobs
make dev-frontend   # Frontend on :3000
```

See [docs/SETUP.md](docs/SETUP.md) for the full setup guide with troubleshooting.

## Project Structure

```
careerdock/
├── backend/
│   ├── cmd/              # Entry points (api, worker, migrate, seed)
│   ├── internal/
│   │   ├── domain/       # Entities, interfaces, errors
│   │   ├── handler/      # HTTP handlers + routes
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Database access
│   │   ├── storage/      # S3/MinIO file storage
│   │   ├── middleware/    # Auth, CORS, logging
│   │   └── config/       # Configuration
│   ├── migrations/       # SQL migration files
│   └── seeds/            # Seed data (companies.json)
├── frontend/
│   └── src/
│       ├── app/          # Next.js App Router pages
│       ├── components/   # React components
│       ├── hooks/        # Custom hooks
│       ├── lib/          # API client, utilities
│       ├── store/        # Zustand state
│       └── types/        # TypeScript types
├── docs/                 # Design docs, setup guide
├── docker-compose.yml    # Local dev infrastructure
└── Makefile              # Dev commands
```

## Make Commands

| Command | Description |
|---------|-------------|
| `make dev-infra` | Start Postgres, Redis, MinIO, Mailhog |
| `make dev-api` | Go API server with hot-reload |
| `make dev-worker` | Background worker with hot-reload |
| `make dev-frontend` | Next.js dev server |
| `make migrate` | Run database migrations |
| `make seed` | Seed company data |
| `make test` | Run all tests |
| `make lint` | Run all linters |
| `make build` | Production build |
| `make clean` | Stop Docker, remove volumes |

## Documentation

| Document | Description |
|----------|-------------|
| [SETUP.md](docs/SETUP.md) | Local development setup guide |
| [PRD.md](docs/PRD.md) | Product requirements |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | System architecture |
| [BUILD-PLAN.md](docs/BUILD-PLAN.md) | Sprint breakdown and task list |
| [STATUS.md](docs/STATUS.md) | Current build progress |
| [LLD/](docs/LLD/) | Low-level design (database, API, frontend, AI, payments) |
| [SECURITY.md](docs/SECURITY.md) | Security design |
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | Deployment strategy |

## License

Private — all rights reserved.
