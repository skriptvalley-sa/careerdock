.PHONY: dev dev-infra dev-api dev-worker dev-frontend \
        test test-backend test-frontend lint build \
        migrate migrate-down migrate-new seed \
        docker-build clean

# ─── Development ─────────────────────────────────────────

## Start all infrastructure (Postgres, Redis, MinIO, Mailhog)
dev-infra:
	docker compose up -d

## Start Go API server with hot reload
dev-api:
	cd backend && air -c .air.api.toml

## Start Asynq worker with hot reload
dev-worker:
	cd backend && air -c .air.worker.toml

## Start Next.js dev server
dev-frontend:
	cd frontend && npm run dev

## Start everything (run in separate terminals or use a process manager)
dev: dev-infra
	@echo "Infrastructure started. Run in separate terminals:"
	@echo "  make dev-api"
	@echo "  make dev-worker"
	@echo "  make dev-frontend"

# ─── Testing ─────────────────────────────────────────────

## Run all backend tests
test-backend:
	cd backend && go test -race -count=1 ./...

## Run frontend tests
test-frontend:
	cd frontend && npm test

## Run all tests
test: test-backend test-frontend

# ─── Linting ─────────────────────────────────────────────

## Run Go linter
lint-backend:
	cd backend && golangci-lint run ./...

## Run frontend linter
lint-frontend:
	cd frontend && npm run lint

## Run all linters
lint: lint-backend lint-frontend

# ─── Database ────────────────────────────────────────────

## Run pending migrations
migrate:
	cd backend && go run ./cmd/migrate/ up

## Rollback last migration
migrate-down:
	cd backend && go run ./cmd/migrate/ down 1

## Create new migration pair (usage: make migrate-new NAME=add_xyz_table)
migrate-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-new NAME=description"; exit 1; fi
	./scripts/gen-migration.sh $(NAME)

## Seed database with initial data
seed:
	cd backend && go run ./cmd/seed/

# ─── Build ───────────────────────────────────────────────

## Build Go binaries
build-backend:
	cd backend && CGO_ENABLED=0 go build -o bin/api ./cmd/api/
	cd backend && CGO_ENABLED=0 go build -o bin/worker ./cmd/worker/

## Build frontend
build-frontend:
	cd frontend && npm run build

## Build everything
build: build-backend build-frontend

## Build Docker images
docker-build:
	docker build -f infra/docker/Dockerfile.api -t careerdock-api:latest ./backend
	docker build -f infra/docker/Dockerfile.worker -t careerdock-worker:latest ./backend

# ─── Cleanup ─────────────────────────────────────────────

## Stop all Docker services and remove volumes
clean:
	docker compose down -v
	cd backend && rm -rf bin/ tmp/
