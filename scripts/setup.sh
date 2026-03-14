#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "setup.sh is a compatibility wrapper. Forwarding to ./scripts/dev.sh setup..."
exec "$SCRIPT_DIR/dev.sh" setup "$@"
