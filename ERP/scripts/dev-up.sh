#!/usr/bin/env bash
# Start the full development environment.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "==> Starting Docker services..."
docker compose up -d postgres

echo "==> Waiting for PostgreSQL to be ready..."
until docker compose exec -T postgres pg_isready -U postgres -d coffeeshop > /dev/null 2>&1; do
  sleep 1
done
echo "    PostgreSQL ready."

echo "==> Running migrations..."
cd apps/server
DATABASE_URL="postgres://postgres:postgres@localhost:5433/coffeeshop?sslmode=disable" \
  go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  -path ../../db/migrations/postgres \
  -database "$DATABASE_URL" up

echo "==> Starting API server (hot-reload)..."
cd "$REPO_ROOT/apps/server"
DATABASE_URL="postgres://postgres:postgres@localhost:5433/coffeeshop?sslmode=disable" \
  go run ./cmd/server &
SERVER_PID=$!

echo "==> Starting web frontend..."
cd "$REPO_ROOT/apps/web"
npm run dev &
WEB_PID=$!

echo ""
echo "  API  → http://localhost:8080"
echo "  Web  → http://localhost:5173"
echo "  pgAdmin → http://localhost:8081  (admin@local.dev / admin)"
echo ""
echo "Press Ctrl+C to stop all services."

trap "kill $SERVER_PID $WEB_PID 2>/dev/null; docker compose stop" INT TERM
wait
