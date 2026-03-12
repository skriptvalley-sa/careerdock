#!/usr/bin/env bash
set -euo pipefail

NAME="$1"
TIMESTAMP=$(date +%Y%m%d%H%M%S)
DIR="backend/migrations"

touch "${DIR}/${TIMESTAMP}_${NAME}.up.sql"
touch "${DIR}/${TIMESTAMP}_${NAME}.down.sql"

echo "Created migration files:"
echo "  ${DIR}/${TIMESTAMP}_${NAME}.up.sql"
echo "  ${DIR}/${TIMESTAMP}_${NAME}.down.sql"
