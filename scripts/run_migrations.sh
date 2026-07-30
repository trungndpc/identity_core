#!/usr/bin/env bash
# Run SQL migrations for identity_core (idempotent).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="$ROOT/scripts/migrations"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-identity}"
DB_PASSWORD="${DB_PASSWORD:-identity}"
DB_NAME="${DB_NAME:-identity_core}"

export PGPASSWORD="$DB_PASSWORD"

run_sql() {
  local file="$1"
  echo "→ $(basename "$file")"
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -f "$file"
}

if ! command -v psql >/dev/null 2>&1; then
  echo "✗ psql not found. Install PostgreSQL client first." >&2
  exit 1
fi

shopt -s nullglob
files=("$MIGRATIONS_DIR"/*.sql)
shopt -u nullglob

if [[ ${#files[@]} -eq 0 ]]; then
  echo "✗ No migration files in $MIGRATIONS_DIR" >&2
  exit 1
fi

IFS=$'\n' files=($(printf '%s\n' "${files[@]}" | sort))
unset IFS

for file in "${files[@]}"; do
  run_sql "$file"
done

echo "✓ Migrations completed"
