#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DIR="$PROJECT_ROOT/backend/migrations"

usage() {
  cat <<EOF
Usage: ./scripts/gen-migration.sh <name>

Creates the next sequential migration pair in backend/migrations/.
Example:
  ./scripts/gen-migration.sh add_interview_rounds_table
EOF
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  usage >&2
  exit 1
fi

NAME="$1"
NAME="${NAME// /_}"
NAME="${NAME//-/_}"

if [[ ! "$NAME" =~ ^[a-z0-9_]+$ ]]; then
  echo "Migration name must contain only lowercase letters, numbers, underscores, spaces, or hyphens." >&2
  exit 1
fi

if [[ ! -d "$DIR" ]]; then
  echo "Migration directory not found: $DIR" >&2
  exit 1
fi

LAST_NUMBER=$(
  find "$DIR" -maxdepth 1 -type f -name '*.up.sql' \
    | sed -E 's#.*/([0-9]{6})_.*#\1#' \
    | sort -n \
    | tail -1
)

if [[ -z "$LAST_NUMBER" ]]; then
  NEXT_NUMBER="000001"
else
  NEXT_NUMBER=$(printf "%06d" $((10#$LAST_NUMBER + 1)))
fi

UP_FILE="$DIR/${NEXT_NUMBER}_${NAME}.up.sql"
DOWN_FILE="$DIR/${NEXT_NUMBER}_${NAME}.down.sql"

touch "$UP_FILE" "$DOWN_FILE"

echo "Created migration files:"
echo "  ${UP_FILE#$PROJECT_ROOT/}"
echo "  ${DOWN_FILE#$PROJECT_ROOT/}"
