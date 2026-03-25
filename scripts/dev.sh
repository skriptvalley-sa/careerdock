#!/usr/bin/env bash
# CareerDock Development Environment Manager
# Usage: ./scripts/dev.sh <command> [args]
set -euo pipefail

# ─── Global Constants ────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEV_DIR="$PROJECT_ROOT/.dev"
PID_DIR="$DEV_DIR/pids"
LOG_DIR="$DEV_DIR/logs"
STATE_DIR="$DEV_DIR/state"

API_PORT=8080
FRONTEND_PORT=3000

# ─── PATH Bootstrapping ────────────────────────────────────────────────────
# On VPS or non-login shells, Go and Air may not be in PATH.
# Bootstrap common install locations so the script is self-sufficient.
for p in /usr/local/go/bin "$HOME/go/bin" "$HOME/.local/bin"; do
  [[ -d "$p" ]] && [[ ":$PATH:" != *":$p:"* ]] && export PATH="$p:$PATH"
done

WATCHDOG_INTERVAL=30
GRACEFUL_TIMEOUT=10
HEALTH_TIMEOUT=30
API_HEALTH_TIMEOUT=60  # First air build can be slow
MAX_RESTART_ATTEMPTS=3
RESTART_WINDOW=300  # 5 minutes

# ─── Color Output ────────────────────────────────────────────────────────────

if [[ -t 1 ]] && [[ "${TERM:-}" != "dumb" ]]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BLUE='\033[0;34m'
  CYAN='\033[0;36m'
  BOLD='\033[1m'
  DIM='\033[2m'
  NC='\033[0m'
else
  RED='' GREEN='' YELLOW='' BLUE='' CYAN='' BOLD='' DIM='' NC=''
fi

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# ─── Utility Functions ───────────────────────────────────────────────────────

# Portable version comparison: returns 0 if $1 >= $2
version_gte() {
  printf '%s\n%s' "$2" "$1" | sort -t. -k1,1n -k2,2n -k3,3n -C 2>/dev/null
}

# Wait for a condition with timeout
wait_for() {
  local description="$1"
  local cmd="$2"
  local max_seconds="${3:-30}"
  local interval="${4:-2}"
  local elapsed=0

  printf "  Waiting for %s " "$description"
  while (( elapsed < max_seconds )); do
    if eval "$cmd" >/dev/null 2>&1; then
      echo -e " ${GREEN}ready${NC}"
      return 0
    fi
    printf "."
    sleep "$interval"
    elapsed=$((elapsed + interval))
  done
  echo -e " ${RED}timeout${NC}"
  return 1
}

# Check if a component is running by PID
is_running() {
  local component="$1"
  local pidfile="$PID_DIR/${component}.pid"
  if [[ ! -f "$pidfile" ]]; then
    return 1
  fi
  local pid
  pid=$(cat "$pidfile")
  if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
    return 0
  else
    rm -f "$pidfile"
    return 1
  fi
}

# Get PID of a component
get_pid() {
  local component="$1"
  local pidfile="$PID_DIR/${component}.pid"
  if [[ -f "$pidfile" ]]; then
    cat "$pidfile"
  fi
}

# Check if a port is in use
port_in_use() {
  local port="$1"
  lsof -i ":${port}" -t >/dev/null 2>&1
}

# Get PID using a specific port
port_pid() {
  local port="$1"
  lsof -i ":${port}" -t 2>/dev/null | head -1
}

# Gracefully stop a process and its entire process tree
graceful_stop() {
  local component="$1"
  local timeout="${2:-$GRACEFUL_TIMEOUT}"
  local pidfile="$PID_DIR/${component}.pid"

  if [[ ! -f "$pidfile" ]]; then
    return 0
  fi

  local pid
  pid=$(cat "$pidfile")
  if ! kill -0 "$pid" 2>/dev/null; then
    rm -f "$pidfile"
    return 0
  fi

  printf "  Stopping %-12s (PID %s) " "$component" "$pid"

  # Send SIGTERM to the process group (negative PID), then to the process itself
  kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null || true

  # Wait for graceful shutdown
  local elapsed=0
  while (( elapsed < timeout )); do
    if ! kill -0 "$pid" 2>/dev/null; then
      echo -e "${GREEN}stopped${NC}"
      rm -f "$pidfile"
      _cleanup_port "$component"
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done

  # Force kill the entire process tree
  kill -9 -- "-$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
  # Also kill any children that escaped the process group
  pkill -9 -P "$pid" 2>/dev/null || true
  sleep 1
  echo -e "${YELLOW}killed${NC}"
  rm -f "$pidfile"
  _cleanup_port "$component"
}

# Kill any orphan process still holding the component's port
_cleanup_port() {
  local component="$1"
  local port=""
  case "$component" in
    api)      port="$API_PORT" ;;
    frontend) port="$FRONTEND_PORT" ;;
    *)        return 0 ;;
  esac

  if port_in_use "$port"; then
    local orphan_pid
    orphan_pid=$(port_pid "$port")
    if [[ -n "$orphan_pid" ]]; then
      warn "Killing orphan process $orphan_pid on port $port"
      kill -9 "$orphan_pid" 2>/dev/null || true
      sleep 1
    fi
  fi
}

# Ensure .dev directories exist
ensure_dev_dirs() {
  mkdir -p "$PID_DIR" "$LOG_DIR" "$STATE_DIR"
}

state_file() {
  local component="$1"
  echo "$STATE_DIR/${component}.managed"
}

mark_managed() {
  local component="$1"
  touch "$(state_file "$component")"
}

unmark_managed() {
  local component="$1"
  rm -f "$(state_file "$component")"
}

is_managed() {
  local component="$1"
  [[ -f "$(state_file "$component")" ]]
}

launch_component() {
  local component="$1"
  local command

  case "$component" in
    api)
      command="cd backend && exec air -c .air.api.toml"
      ;;
    worker)
      command="cd backend && exec air -c .air.worker.toml"
      ;;
    frontend)
      command="cd frontend && exec npm run dev"
      ;;
    *)
      error "Unknown component: $component"
      return 1
      ;;
  esac

  cd "$PROJECT_ROOT"

  # Start in a new session (setsid) so we can kill the entire process tree.
  # On macOS, setsid may not exist — fall back to plain bash.
  if command -v setsid >/dev/null 2>&1; then
    setsid bash -c "$command" >> "$LOG_DIR/${component}.log" 2>&1 &
  else
    bash -c "$command" >> "$LOG_DIR/${component}.log" 2>&1 &
  fi
  STARTED_PID=$!
  echo "$STARTED_PID" > "$PID_DIR/${component}.pid"
  mark_managed "$component"
}

# Sync NEXT_PUBLIC_* vars from root .env to frontend/.env.local.
# Next.js only reads env files relative to the frontend/ directory, so the
# root .env is invisible to it when launched from frontend/.
sync_frontend_env() {
  local root_env="$PROJECT_ROOT/.env"
  local frontend_env="$PROJECT_ROOT/frontend/.env.local"

  if [[ ! -f "$root_env" ]]; then
    return
  fi

  # Extract all NEXT_PUBLIC_* lines from root .env
  local next_vars
  next_vars=$(grep -E '^NEXT_PUBLIC_' "$root_env" 2>/dev/null || true)

  if [[ -z "$next_vars" ]]; then
    return
  fi

  # Write (or overwrite) frontend/.env.local
  {
    echo "# Auto-generated from root .env by scripts/dev.sh"
    echo "# Do not edit — changes will be overwritten on next start/setup."
    echo "$next_vars"
  } > "$frontend_env"

  success "Synced frontend/.env.local ($(echo "$next_vars" | wc -l | tr -d ' ') vars)"
}

# Validate that go and air are available before starting backend processes.
require_go_and_air() {
  if ! command -v go >/dev/null 2>&1; then
    error "go not found in PATH. Install Go >= 1.25 or check your PATH."
    error "Searched: $PATH"
    exit 1
  fi
  if ! command -v air >/dev/null 2>&1; then
    error "air not found in PATH. Install with: go install github.com/air-verse/air@latest"
    error "Searched: $PATH"
    exit 1
  fi
}

# ─── Setup Command ───────────────────────────────────────────────────────────

cmd_setup() {
  echo -e "${BOLD}${CYAN}CareerDock Development Setup${NC}"
  echo ""

  local os
  os=$(uname -s)
  info "Detected OS: $os"
  echo ""

  local all_ok=true
  local missing=()

  # Check Go
  echo -e "${BOLD}Checking prerequisites...${NC}"
  if command -v go >/dev/null 2>&1; then
    local go_ver
    go_ver=$(go version | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
    if version_gte "$go_ver" "1.25"; then
      success "Go $go_ver"
    else
      warn "Go $go_ver found (need >= 1.25)"
      all_ok=false
      missing+=("Go >= 1.25")
    fi
  else
    error "Go not found"
    all_ok=false
    missing+=("Go")
  fi

  # Check Node.js
  if command -v node >/dev/null 2>&1; then
    local node_ver
    node_ver=$(node --version | tr -d 'v')
    local node_major
    node_major=$(echo "$node_ver" | cut -d. -f1)
    if (( node_major >= 20 )); then
      success "Node.js $node_ver"
    else
      warn "Node.js $node_ver found (need >= 20)"
      all_ok=false
      missing+=("Node.js >= 20")
    fi
  else
    error "Node.js not found"
    all_ok=false
    missing+=("Node.js")
  fi

  # Check Docker
  if command -v docker >/dev/null 2>&1; then
    success "Docker $(docker --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)"
  else
    error "Docker not found"
    all_ok=false
    missing+=("Docker")
  fi

  # Check Docker Compose
  if docker compose version >/dev/null 2>&1; then
    success "Docker Compose $(docker compose version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)"
  else
    error "Docker Compose not found (need docker compose v2)"
    all_ok=false
    missing+=("Docker Compose v2")
  fi

  # Check npm
  if command -v npm >/dev/null 2>&1; then
    success "npm $(npm --version)"
  else
    error "npm not found"
    all_ok=false
    missing+=("npm")
  fi

  echo ""

  # Print install instructions for missing prerequisites
  if [[ ${#missing[@]} -gt 0 ]]; then
    echo -e "${BOLD}Missing prerequisites:${NC}"
    for m in "${missing[@]}"; do
      echo -e "  ${RED}x${NC} $m"
    done
    echo ""
    if [[ "$os" == "Darwin" ]]; then
      info "Install with: brew install go node docker"
    else
      info "Install Go from https://go.dev/dl/, Node from https://nodejs.org/"
    fi
    echo ""
  fi

  # Install dev tools
  echo -e "${BOLD}Installing dev tools...${NC}"

  # Air (Go hot-reload)
  if command -v air >/dev/null 2>&1; then
    success "Air already installed"
  else
    info "Installing Air..."
    if go install github.com/air-verse/air@latest 2>/dev/null; then
      success "Air installed"
    else
      warn "Failed to install Air (run: go install github.com/air-verse/air@latest)"
    fi
  fi

  # golangci-lint v2
  if command -v golangci-lint >/dev/null 2>&1; then
    local lint_ver
    lint_ver=$(golangci-lint --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    local lint_major
    lint_major=$(echo "$lint_ver" | cut -d. -f1)
    if (( lint_major >= 2 )); then
      success "golangci-lint $lint_ver (v2)"
    else
      warn "golangci-lint $lint_ver found (need v2)"
      info "Installing golangci-lint v2..."
      if [[ "$os" == "Darwin" ]]; then
        brew install golangci-lint 2>/dev/null && success "golangci-lint v2 installed" || warn "Failed — run: brew install golangci-lint"
      else
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest 2>/dev/null && success "golangci-lint v2 installed" || warn "Failed — see https://golangci-lint.run/welcome/install/"
      fi
    fi
  else
    info "Installing golangci-lint v2..."
    if [[ "$os" == "Darwin" ]]; then
      brew install golangci-lint 2>/dev/null && success "golangci-lint v2 installed" || warn "Failed — run: brew install golangci-lint"
    else
      go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest 2>/dev/null && success "golangci-lint v2 installed" || warn "Failed — see https://golangci-lint.run/welcome/install/"
    fi
  fi

  echo ""

  # Environment file
  echo -e "${BOLD}Environment setup...${NC}"
  if [[ -f "$PROJECT_ROOT/.env" ]]; then
    success ".env file exists"
  else
    cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
    success ".env created from .env.example"
  fi

  # Sync frontend env
  sync_frontend_env

  # Frontend dependencies
  echo ""
  echo -e "${BOLD}Frontend dependencies...${NC}"
  if [[ -d "$PROJECT_ROOT/frontend/node_modules" ]]; then
    success "node_modules exists (run 'cd frontend && npm install' to update)"
  else
    info "Running npm install..."
    cd "$PROJECT_ROOT/frontend" && npm install --silent 2>/dev/null
    success "Frontend dependencies installed"
  fi

  # Check .gitignore
  if ! grep -q '^\.dev' "$PROJECT_ROOT/.gitignore" 2>/dev/null; then
    warn ".dev/ is not in .gitignore — add it to prevent committing runtime files"
  fi

  echo ""
  if [[ "$all_ok" == true ]]; then
    echo -e "${GREEN}${BOLD}Setup complete! Run './scripts/dev.sh start' to begin.${NC}"
  else
    echo -e "${YELLOW}${BOLD}Setup complete with warnings. Fix missing prerequisites above.${NC}"
  fi
}

# ─── Start Command ───────────────────────────────────────────────────────────

cmd_start() {
  echo -e "${BOLD}${CYAN}Starting CareerDock Development Environment${NC}"
  echo ""

  # Pre-flight checks
  require_go_and_air

  if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
    error "No .env file found. Run './scripts/dev.sh setup' first."
    exit 3
  fi

  if ! docker info >/dev/null 2>&1; then
    error "Docker daemon is not running. Start Docker Desktop or the docker service."
    exit 1
  fi

  ensure_dev_dirs

  # Truncate logs on fresh start
  for f in api.log worker.log frontend.log watchdog.log; do
    : > "$LOG_DIR/$f"
  done

  # ── Docker Infrastructure ──
  echo -e "${BOLD}Starting infrastructure...${NC}"
  cd "$PROJECT_ROOT"
  docker compose up -d 2>/dev/null

  # Wait for health checks
  wait_for "PostgreSQL" \
    "docker compose exec -T postgres pg_isready -U careerdock" \
    "$HEALTH_TIMEOUT" 2 || { error "PostgreSQL failed to start"; exit 1; }

  wait_for "Redis" \
    "docker compose exec -T redis redis-cli ping 2>/dev/null | grep -q PONG" \
    "$HEALTH_TIMEOUT" 2 || { error "Redis failed to start"; exit 1; }

  wait_for "MinIO" \
    "curl -sf http://localhost:9000/minio/health/live" \
    "$HEALTH_TIMEOUT" 2 || warn "MinIO health check failed (may still be initializing)"

  echo ""

  # ── Database Migrations ──
  echo -e "${BOLD}Running database migrations...${NC}"
  # Use `if` so set -e doesn't abort on non-zero exit before we can inspect it.
  # Output streams live so the terminal doesn't appear frozen during Go compile.
  if make -C "$PROJECT_ROOT" migrate; then
    success "Migrations complete"
  else
    # Check whether the failure is a dirty-state — a previous run was interrupted.
    _ver_out=$(cd "$PROJECT_ROOT/backend" && go run ./cmd/migrate/ version 2>&1) || true
    if printf '%s' "$_ver_out" | grep -q "Dirty: true"; then
      _dirty_ver=$(printf '%s' "$_ver_out" | grep -oE 'Version: [0-9]+' | grep -oE '[0-9]+')
      warn "Dirty migration at version ${_dirty_ver} — auto-recovering..."
      if (cd "$PROJECT_ROOT/backend" && go run ./cmd/migrate/ force "$_dirty_ver") && \
         make -C "$PROJECT_ROOT" migrate; then
        success "Migrations complete (recovered)"
      else
        warn "Migration failed after recovery — run: make migrate"
      fi
    else
      warn "Migration returned non-zero (may already be up to date)"
    fi
  fi

  # ── Auto-seed companies (only if table is empty) ──
  _company_count=$(docker compose -f "$PROJECT_ROOT/docker-compose.yml" exec -T postgres \
    psql -U careerdock -d careerdock -t -c "SELECT COUNT(*) FROM companies;" 2>/dev/null | tr -d ' \n')
  if [ "${_company_count:-0}" = "0" ]; then
    echo -e "${BOLD}Seeding initial company data...${NC}"
    if make -C "$PROJECT_ROOT" seed; then
      success "Seed complete"
    else
      warn "Seed failed — run: make seed"
    fi
  else
    info "Companies already seeded (${_company_count} records — skipping)"
  fi

  # ── Ensure admin user exists ──
  # Credentials are read from .env (ADMIN_EMAIL / ADMIN_NAME / ADMIN_PASSWORD).
  # The password is bcrypt-hashed at runtime using a throwaway //go:build ignore
  # Go file so the project's existing golang.org/x/crypto/bcrypt dep is reused
  # — no extra tooling required.
  _admin_email=$(grep -E '^ADMIN_EMAIL='    "$PROJECT_ROOT/.env" 2>/dev/null | cut -d= -f2- | tr -d '"'"'" | xargs)
  _admin_name=$(grep  -E '^ADMIN_NAME='     "$PROJECT_ROOT/.env" 2>/dev/null | cut -d= -f2- | tr -d '"'"'" | xargs)
  _admin_pass=$(grep  -E '^ADMIN_PASSWORD=' "$PROJECT_ROOT/.env" 2>/dev/null | cut -d= -f2- | tr -d '"'"'" | xargs)

  if [ -z "$_admin_email" ] || [ -z "$_admin_pass" ]; then
    warn "ADMIN_EMAIL / ADMIN_PASSWORD not set in .env — skipping admin seed"
  else
    _admin_name="${_admin_name:-Root Admin}"
    _admin_exists=$(docker compose -f "$PROJECT_ROOT/docker-compose.yml" exec -T postgres \
      psql -U careerdock -d careerdock -t -c \
      "SELECT COUNT(*) FROM users WHERE email='$_admin_email';" 2>/dev/null | tr -d ' \n')

    if [ "${_admin_exists:-0}" = "0" ]; then
      # Hash the plaintext password from .env using Go's bcrypt package.
      _bcrypt_src=$(mktemp /tmp/bcrypt_XXXXXX.go)
      cat > "$_bcrypt_src" << 'GOEOF'
//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	h, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 12)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(h))
}
GOEOF
      _admin_hash=$(cd "$PROJECT_ROOT/backend" && go run "$_bcrypt_src" "$_admin_pass" 2>/dev/null)
      rm -f "$_bcrypt_src"

      if [ -n "$_admin_hash" ]; then
        # Pipe the SQL through stdin so the bcrypt $ signs are never touched by the
        # shell (printf %s inserts them as-is; SQL single-quoted strings allow $).
        printf "INSERT INTO users (email, password_hash, name, role, email_verified)\nVALUES ('%s', '%s', '%s', 'admin', true)\nON CONFLICT (email) DO NOTHING;\n" \
          "$_admin_email" "$_admin_hash" "$_admin_name" | \
          docker compose -f "$PROJECT_ROOT/docker-compose.yml" exec -T postgres \
          psql -U careerdock -d careerdock 2>/dev/null

        echo ""
        echo -e "  ${BOLD}${YELLOW}╔══════════════════════════════════════════════════╗${NC}"
        echo -e "  ${BOLD}${YELLOW}║       DEV ADMIN CREDENTIALS (first run)         ║${NC}"
        echo -e "  ${BOLD}${YELLOW}╠══════════════════════════════════════════════════╣${NC}"
        printf   "  ${BOLD}${YELLOW}║  Email:    %-38s ║${NC}\n" "$_admin_email"
        printf   "  ${BOLD}${YELLOW}║  Password: %-38s ║${NC}\n" "$_admin_pass"
        echo -e "  ${BOLD}${YELLOW}╠══════════════════════════════════════════════════╣${NC}"
        echo -e "  ${BOLD}${YELLOW}║  Source:   .env  (ADMIN_EMAIL / ADMIN_PASSWORD) ║${NC}"
        echo -e "  ${BOLD}${YELLOW}╚══════════════════════════════════════════════════╝${NC}"
      else
        warn "Could not hash admin password — is Go installed?"
      fi
    else
      info "Dev admin user already exists ($_admin_email)"
    fi
  fi
  echo ""

  # ── Application Processes ──
  echo -e "${BOLD}Starting application processes...${NC}"
  local started=()

  # API Server
  if is_running api; then
    info "API server already running (PID $(get_pid api))"
    mark_managed api
  elif port_in_use "$API_PORT"; then
    warn "Port $API_PORT already in use by PID $(port_pid $API_PORT) — skipping API"
  else
    launch_component api
    success "API server started (PID $STARTED_PID)"
    started+=(api)
  fi

  # Worker
  if is_running worker; then
    info "Worker already running (PID $(get_pid worker))"
    mark_managed worker
  else
    launch_component worker
    success "Worker started (PID $STARTED_PID)"
    started+=(worker)
  fi

  # Sync frontend env before launch
  sync_frontend_env

  # Frontend
  if is_running frontend; then
    info "Frontend already running (PID $(get_pid frontend))"
    mark_managed frontend
  elif port_in_use "$FRONTEND_PORT"; then
    warn "Port $FRONTEND_PORT already in use by PID $(port_pid $FRONTEND_PORT) — skipping frontend"
  else
    launch_component frontend
    success "Frontend started (PID $STARTED_PID)"
    started+=(frontend)
  fi

  echo ""

  # Wait for API health
  if [[ " ${started[*]:-} " =~ " api " ]] || is_running api; then
    wait_for "API health" \
      "curl -sf http://localhost:${API_PORT}/api/health" \
      "$API_HEALTH_TIMEOUT" 3 || warn "API health check not yet passing (check logs: ./scripts/dev.sh logs api)"
  fi

  # ── Health Watchdog ──
  if is_running watchdog; then
    info "Watchdog already running (PID $(get_pid watchdog))"
  else
    _run_watchdog &
    echo $! > "$PID_DIR/watchdog.pid"
    disown
    success "Health watchdog started (PID $!)"
  fi

  echo ""
  _print_status_table
}

# ─── Stop Command ────────────────────────────────────────────────────────────

cmd_stop() {
  echo -e "${BOLD}${CYAN}Stopping CareerDock Development Environment${NC}"
  echo ""

  # Stop app processes in reverse order
  echo -e "${BOLD}Stopping application processes...${NC}"
  for component in watchdog frontend worker api; do
    if is_running "$component"; then
      graceful_stop "$component"
    fi
    unmark_managed "$component"
  done

  echo ""

  # Stop Docker services (preserve data)
  echo -e "${BOLD}Stopping Docker services...${NC}"
  cd "$PROJECT_ROOT"
  docker compose stop 2>/dev/null
  success "Docker services stopped (data preserved)"

  echo ""
  echo -e "${DIM}Run 'make clean' to remove Docker volumes and data${NC}"
  echo -e "${DIM}Logs preserved in .dev/logs/${NC}"
}

# ─── Restart Command ─────────────────────────────────────────────────────────

cmd_restart() {
  local component="${1:-}"

  if [[ -z "$component" ]]; then
    cmd_stop
    echo ""
    cmd_start
    return
  fi

  case "$component" in
    api|worker|frontend)
      echo -e "${BOLD}Restarting $component...${NC}"
      if is_running "$component"; then
        graceful_stop "$component"
      fi

      ensure_dev_dirs
      : > "$LOG_DIR/${component}.log"

      # Sync frontend env before restarting frontend
      if [[ "$component" == "frontend" ]]; then
        sync_frontend_env
      fi

      launch_component "$component"
      success "$component restarted (PID $STARTED_PID)"
      ;;
    *)
      error "Unknown component: $component (use: api, worker, frontend)"
      exit 1
      ;;
  esac
}

# ─── Status Command ──────────────────────────────────────────────────────────

cmd_status() {
  echo -e "${BOLD}${CYAN}CareerDock Development Environment Status${NC}"
  echo ""
  _print_status_table
}

_print_status_table() {
  local fmt="  %-16s %-12s %-8s %s\n"

  printf "  ${BOLD}%-16s %-12s %-8s %s${NC}\n" "COMPONENT" "STATUS" "PORT" "URL"
  echo "  ──────────────────────────────────────────────────────────────"

  # Docker services
  local docker_running=false
  if docker compose ps --format json 2>/dev/null | head -1 | grep -q "running" 2>/dev/null; then
    docker_running=true
  fi

  _status_line "PostgreSQL" \
    "$(docker compose ps -q postgres 2>/dev/null | xargs -r docker inspect --format '{{.State.Status}}' 2>/dev/null || echo stopped)" \
    "5432" ""

  _status_line "Redis" \
    "$(docker compose ps -q redis 2>/dev/null | xargs -r docker inspect --format '{{.State.Status}}' 2>/dev/null || echo stopped)" \
    "6379" ""

  _status_line "MinIO" \
    "$(docker compose ps -q minio 2>/dev/null | xargs -r docker inspect --format '{{.State.Status}}' 2>/dev/null || echo stopped)" \
    "9000" "http://localhost:9001"

  _status_line "Mailhog" \
    "$(docker compose ps -q mailhog 2>/dev/null | xargs -r docker inspect --format '{{.State.Status}}' 2>/dev/null || echo stopped)" \
    "1025" "http://localhost:8025"

  echo "  ──────────────────────────────────────────────────────────────"

  # API
  local api_status="stopped"
  if is_running api; then
    if curl -sf "http://localhost:${API_PORT}/api/health" >/dev/null 2>&1; then
      api_status="healthy"
    else
      api_status="starting"
    fi
  fi
  _status_line "API Server" "$api_status" "$API_PORT" "http://localhost:${API_PORT}"

  # Worker
  local worker_status="stopped"
  if is_running worker; then
    worker_status="running"
  fi
  _status_line "Worker" "$worker_status" "-" ""

  # Frontend
  local frontend_status="stopped"
  if is_running frontend; then
    if curl -sf "http://localhost:${FRONTEND_PORT}" >/dev/null 2>&1; then
      frontend_status="healthy"
    else
      frontend_status="starting"
    fi
  fi
  _status_line "Frontend" "$frontend_status" "$FRONTEND_PORT" "http://localhost:${FRONTEND_PORT}"

  # Watchdog
  local watchdog_status="stopped"
  if is_running watchdog; then
    watchdog_status="running"
  fi
  _status_line "Watchdog" "$watchdog_status" "-" ""

  echo ""
  echo -e "  ${DIM}Logs: .dev/logs/  |  Stop: ./scripts/dev.sh stop${NC}"
}

_status_line() {
  local name="$1" status="$2" port="$3" url="$4"
  local color

  case "$status" in
    running|healthy)   color="$GREEN" ;;
    starting)          color="$YELLOW" ;;
    stopped|exited|"") color="$RED"; status="${status:-stopped}" ;;
    *)                 color="$YELLOW" ;;
  esac

  printf "  %-16s ${color}%-12s${NC} %-8s %s\n" "$name" "$status" "$port" "$url"
}

# ─── Logs Command ────────────────────────────────────────────────────────────

cmd_logs() {
  local component="${1:-all}"

  case "$component" in
    api|worker|frontend|watchdog)
      local logfile="$LOG_DIR/${component}.log"
      if [[ ! -f "$logfile" ]]; then
        error "No log file for $component (has it been started?)"
        exit 1
      fi
      info "Tailing $component logs (Ctrl+C to stop)"
      tail -f "$logfile"
      ;;
    all)
      local logfiles=()
      for c in api worker frontend watchdog; do
        if [[ -f "$LOG_DIR/${c}.log" ]]; then
          logfiles+=("$LOG_DIR/${c}.log")
        fi
      done
      if [[ ${#logfiles[@]} -eq 0 ]]; then
        error "No log files found (has the environment been started?)"
        exit 1
      fi
      info "Tailing all logs (Ctrl+C to stop)"
      tail -f "${logfiles[@]}"
      ;;
    *)
      error "Unknown component: $component (use: api, worker, frontend, watchdog, all)"
      exit 1
      ;;
  esac
}

# ─── Health Watchdog ─────────────────────────────────────────────────────────

_run_watchdog() {
  # Avoid Bash 4-only associative arrays so the script still works on macOS.
  local api_restart_count=0
  local worker_restart_count=0
  local frontend_restart_count=0
  local api_last_restart_time=0
  local worker_last_restart_time=0
  local frontend_last_restart_time=0

  while true; do
    local now
    now=$(date +%s)
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    for component in api worker frontend; do
      local count last_time

      if ! is_managed "$component"; then
        continue
      fi

      case "$component" in
        api)
          count=$api_restart_count
          last_time=$api_last_restart_time
          ;;
        worker)
          count=$worker_restart_count
          last_time=$worker_last_restart_time
          ;;
        frontend)
          count=$frontend_restart_count
          last_time=$frontend_last_restart_time
          ;;
      esac

      if ! is_running "$component"; then
        # Reset counter if enough time has passed since last restart
        if (( now - last_time > RESTART_WINDOW )); then
          count=0
        fi

        if (( count >= MAX_RESTART_ATTEMPTS )); then
          echo "[$timestamp] ALERT: $component has crashed $count times in ${RESTART_WINDOW}s — giving up" >> "$LOG_DIR/watchdog.log"
          continue
        fi

        echo "[$timestamp] $component died, restarting (attempt $((count + 1))/$MAX_RESTART_ATTEMPTS)..." >> "$LOG_DIR/watchdog.log"

        launch_component "$component"

        count=$((count + 1))
        case "$component" in
          api)
            api_restart_count=$count
            api_last_restart_time=$now
            ;;
          worker)
            worker_restart_count=$count
            worker_last_restart_time=$now
            ;;
          frontend)
            frontend_restart_count=$count
            frontend_last_restart_time=$now
            ;;
        esac

        echo "[$timestamp] $component restarted (PID $STARTED_PID)" >> "$LOG_DIR/watchdog.log"
      fi
    done

    # API health check (only if process is alive)
    if is_running api; then
      if ! curl -sf "http://localhost:${API_PORT}/api/health" >/dev/null 2>&1; then
        echo "[$timestamp] WARN: API process alive but health check failed" >> "$LOG_DIR/watchdog.log"
      fi
    fi

    sleep "$WATCHDOG_INTERVAL"
  done
}

# ─── Help ────────────────────────────────────────────────────────────────────

cmd_help() {
  cat <<EOF
${BOLD}${CYAN}CareerDock Development Environment Manager${NC}

${BOLD}Usage:${NC} ./scripts/dev.sh <command> [args]

${BOLD}Commands:${NC}
  setup                Check prerequisites and set up development environment
  start                Start all services (Docker + API + worker + frontend)
  stop                 Gracefully stop all services
  restart [component]  Restart all or a specific component (api|worker|frontend)
  status               Show status of all components
  logs [component]     Tail logs (api|worker|frontend|watchdog|all)
  help                 Show this help message

${BOLD}Examples:${NC}
  ./scripts/dev.sh setup            # First-time setup
  ./scripts/dev.sh start            # Start everything
  ./scripts/dev.sh status           # Check what's running
  ./scripts/dev.sh logs api         # Tail API server logs
  ./scripts/dev.sh restart frontend # Restart just the frontend
  ./scripts/dev.sh stop             # Stop everything

${BOLD}Runtime files:${NC}
  .dev/pids/   PID files for each process
  .dev/logs/   Log files for each process
  .dev/state/  Watchdog state for managed processes
EOF
}

# ─── Main Dispatch ───────────────────────────────────────────────────────────

main() {
  cd "$PROJECT_ROOT"

  local command="${1:-help}"
  shift || true

  case "$command" in
    setup)   cmd_setup "$@" ;;
    start)   cmd_start "$@" ;;
    stop)    cmd_stop "$@" ;;
    restart) cmd_restart "$@" ;;
    status)  cmd_status "$@" ;;
    logs)    cmd_logs "$@" ;;
    help|--help|-h) cmd_help ;;
    *)
      error "Unknown command: $command"
      echo ""
      cmd_help
      exit 1
      ;;
  esac
}

main "$@"
