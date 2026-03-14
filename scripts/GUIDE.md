# Scripts Guide

This directory contains the supported shell entrypoints for local development and database migration scaffolding.

## Quick Start

From the repository root:

```bash
./scripts/dev.sh setup
./scripts/dev.sh start
./scripts/dev.sh status
./scripts/dev.sh logs api
./scripts/dev.sh stop
```

If you prefer the older split-terminal workflow, the underlying `Makefile` targets still work:

```bash
make dev-infra
make dev-api
make dev-worker
make dev-frontend
```

`dev.sh` is the preferred workflow because it manages the full stack, writes logs to `.dev/`, and includes a watchdog that can restart managed processes after crashes.

## Scripts

### `dev.sh`

Primary development environment manager.

Commands:

```bash
./scripts/dev.sh setup
./scripts/dev.sh start
./scripts/dev.sh stop
./scripts/dev.sh restart
./scripts/dev.sh restart api
./scripts/dev.sh restart worker
./scripts/dev.sh restart frontend
./scripts/dev.sh status
./scripts/dev.sh logs api
./scripts/dev.sh logs all
./scripts/dev.sh help
```

What each command does:

- `setup`: checks local prerequisites, installs Air and `golangci-lint` v2 if missing, copies `.env.example` to `.env`, and installs frontend dependencies.
- `start`: starts Docker services, waits for health checks, runs database migrations, then launches API, worker, frontend, and the watchdog.
- `stop`: gracefully stops managed application processes and then stops Docker services without removing volumes.
- `restart [component]`: restarts all managed services or one of `api`, `worker`, `frontend`.
- `status`: prints a combined table for Docker services and app processes.
- `logs [component]`: tails one log or all logs from `.dev/logs/`.

Services started by `dev.sh`:

- PostgreSQL at `localhost:5432`
- Redis at `localhost:6379`
- MinIO API at `localhost:9000`
- MinIO console at `http://localhost:9001`
- Mailhog SMTP at `localhost:1025`
- Mailhog UI at `http://localhost:8025`
- API at `http://localhost:8080`
- Frontend at `http://localhost:3000`

Runtime files created under `.dev/`:

- `.dev/pids/`: PID files for `api`, `worker`, `frontend`, and `watchdog`
- `.dev/logs/`: logs for each managed process
- `.dev/state/`: markers that tell the watchdog which components it is responsible for

Watchdog behavior:

- It checks managed processes every 30 seconds.
- It will restart `api`, `worker`, and `frontend` if they crash.
- It will try up to 3 restarts within a 5-minute window.
- It only restarts components that `dev.sh` successfully started or adopted. If `start` skips a service because a port is already in use, the watchdog leaves it alone.

Operational notes:

- `start` expects `.env` to exist. Run `./scripts/dev.sh setup` first if it does not.
- `start` requires Docker to be running.
- `stop` preserves database and object storage data. Use `make clean` only when you want to remove volumes and backend build artifacts.
- `logs` only works after the environment has created log files.

### `setup.sh`

Compatibility wrapper.

```bash
./scripts/setup.sh
```

This simply forwards to:

```bash
./scripts/dev.sh setup
```

Use `dev.sh setup` directly for new workflows. `setup.sh` is kept so older notes and muscle memory still work.

### `gen-migration.sh`

Creates the next sequential migration pair in `backend/migrations/`.

Usage:

```bash
./scripts/gen-migration.sh add_interview_rounds_table
make migrate-new NAME=add_interview_rounds_table
```

Behavior:

- Scans existing `*.up.sql` files in `backend/migrations/`
- Finds the highest six-digit prefix
- Creates the next pair:

```text
backend/migrations/000020_add_interview_rounds_table.up.sql
backend/migrations/000020_add_interview_rounds_table.down.sql
```

Naming rules:

- Use lowercase names
- Use underscores between words
- Spaces and hyphens are normalized to underscores

The script only creates blank files. It does not write SQL or run the migration.

## Recommended Workflows

### Full local stack

```bash
./scripts/dev.sh setup
./scripts/dev.sh start
./scripts/dev.sh status
./scripts/dev.sh logs all
```

### Restart only one process after config or dependency changes

```bash
./scripts/dev.sh restart api
./scripts/dev.sh restart worker
./scripts/dev.sh restart frontend
```

### Create and run a migration

```bash
make migrate-new NAME=add_resume_metadata
make migrate
make migrate-down
```

## Troubleshooting

### `dev.sh start` says Docker is not running

Start Docker Desktop or your local Docker daemon, then rerun `./scripts/dev.sh start`.

### API or frontend does not come up

Inspect logs:

```bash
./scripts/dev.sh logs api
./scripts/dev.sh logs frontend
```

Also check whether ports `8080` or `3000` are already in use by another process.

### `status` shows `starting` for API

The process is alive but the health endpoint at `/api/health` is not returning success yet. Check the API log.

### `gen-migration.sh` rejects the name

Use lowercase letters, numbers, and underscores. If you pass spaces or hyphens, they are converted to underscores automatically.

## Relationship To The Makefile

The `Makefile` is still the low-level command surface for linting, tests, builds, migrations, and seeding:

- `make lint`
- `make test`
- `make build`
- `make migrate`
- `make migrate-down`
- `make seed`

Use `dev.sh` for environment lifecycle. Use `make` for focused development tasks.
