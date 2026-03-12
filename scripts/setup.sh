#!/usr/bin/env bash
set -euo pipefail

echo "=== CareerDock Local Setup ==="

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Go is required. Install from https://go.dev/dl/"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Node.js is required. Install from https://nodejs.org/"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required. Install from https://docker.com/"; exit 1; }

# Install Air for Go hot reload
echo "Installing Air..."
go install github.com/air-verse/air@latest

# Install golangci-lint
echo "Installing golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.0

# Copy env file
if [ ! -f .env ]; then
    cp .env.example .env
    echo "Created .env from .env.example — update values as needed"
fi

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd frontend && npm install

# Install pre-commit hooks
if command -v pre-commit >/dev/null 2>&1; then
    echo "Installing pre-commit hooks..."
    cd .. && pre-commit install
else
    echo "pre-commit not found. Install with: pip install pre-commit"
fi

echo ""
echo "=== Setup complete! ==="
echo "Run 'make dev' to start infrastructure"
echo "Then in separate terminals:"
echo "  make dev-api"
echo "  make dev-worker"
echo "  make dev-frontend"
