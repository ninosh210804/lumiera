#!/usr/bin/env bash
# Create a release: tag, build, package for flash-drive distribution.
# Usage: VERSION=v1.2.3 ./scripts/release.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:?set VERSION=vX.Y.Z}"
OUT="$REPO_ROOT/dist"
FLASH="$OUT/flash-${VERSION}"

echo "==> Release $VERSION"

# Tag if not already tagged
if ! git -C "$REPO_ROOT" rev-parse "$VERSION" > /dev/null 2>&1; then
  git -C "$REPO_ROOT" tag -a "$VERSION" -m "Release $VERSION"
  echo "    Tagged $VERSION"
fi

# Build everything
VERSION="$VERSION" "$REPO_ROOT/scripts/build-installers.sh"

# Flash-drive layout
mkdir -p "$FLASH/server" "$FLASH/db/migrations"
echo "==> Packaging flash-drive bundle → $FLASH/"

cp "$OUT/erp-server-linux-amd64"    "$FLASH/server/erp-server"
cp "$OUT/erp-server-windows-amd64.exe" "$FLASH/server/erp-server.exe" 2>/dev/null || true
chmod +x "$FLASH/server/erp-server"

# Migrations
cp -r "$REPO_ROOT/db/migrations/postgres" "$FLASH/db/migrations/"

# Bootstrap script that runs on the target machine
cat > "$FLASH/start.sh" << 'BOOT'
#!/usr/bin/env bash
# Run on the target machine: ./start.sh
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
DB_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/coffeeshop?sslmode=disable}"
echo "==> Migrating database..."
migrate -path "$DIR/db/migrations/postgres" -database "$DB_URL" up 2>/dev/null || true
echo "==> Starting ERP server..."
exec "$DIR/server/erp-server" "$@"
BOOT
chmod +x "$FLASH/start.sh"

# README
cat > "$FLASH/README.txt" << 'DOC'
Coffeeshop ERP — Flash Drive Distribution
==========================================

Requirements on target machine:
  - PostgreSQL 15+ (or use the docker-compose.yml in the repo)
  - Linux x86_64 or Windows x64

Quick start:
  1. Set DATABASE_URL or use the default (postgres://postgres:postgres@localhost:5432/coffeeshop)
  2. Run:  ./start.sh          (Linux)
            start.bat          (Windows — create manually if needed)
  3. Open: http://localhost:8080
DOC

# Zip it
cd "$OUT"
ZIP_NAME="coffeeshop-erp-${VERSION}-flash.zip"
zip -r "$ZIP_NAME" "flash-${VERSION}/"
echo ""
echo "==> Flash zip: $OUT/$ZIP_NAME"
echo "==> Done."
