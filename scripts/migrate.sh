#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATION="${1:-$ROOT/migrations/001_plaid.sql}"

if [[ ! -f "$MIGRATION" ]]; then
  echo "Migration file not found: $MIGRATION"
  exit 1
fi

if docker ps --format '{{.Names}}' | grep -qx 'finances_db'; then
  echo "Running migration via Docker (finances_db)..."
  docker exec -i finances_db psql -U postgres -d finances < "$MIGRATION"
  echo "Done."
  exit 0
fi

if [[ -n "${DATABASE_URL:-}" ]]; then
  echo "Running migration via psql and DATABASE_URL..."
  psql "$DATABASE_URL" -f "$MIGRATION"
  echo "Done."
  exit 0
fi

echo "Could not run migration."
echo "Start the database: docker compose up -d"
echo "Or set DATABASE_URL and run: psql \"\$DATABASE_URL\" -f $MIGRATION"
exit 1
