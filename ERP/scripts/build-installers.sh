#!/usr/bin/env bash
# Build the Go server binary + Tauri desktop installer for all platforms.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-$(git -C "$REPO_ROOT" describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
OUT="$REPO_ROOT/dist"

mkdir -p "$OUT"

echo "==> Building erp-server binaries (version: $VERSION)..."
cd "$REPO_ROOT/apps/server"

build_server() {
  local GOOS=$1 GOARCH=$2 EXT=${3:-}
  local BIN="erp-server-${GOOS}-${GOARCH}${EXT}"
  echo "    $GOOS/$GOARCH → $BIN"
  GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags="-s -w -X main.version=${VERSION}" \
    -o "$OUT/$BIN" ./cmd/server
}

build_server linux   amd64
build_server linux   arm64
build_server darwin  amd64
build_server darwin  arm64
build_server windows amd64 .exe

echo "==> Building web frontend..."
cd "$REPO_ROOT/apps/web"
npm ci --silent
npm run build

echo "==> Building Tauri desktop app..."
cd "$REPO_ROOT/apps/desktop"
# Copy the Linux server binary as the sidecar
mkdir -p src-tauri/binaries
cp "$OUT/erp-server-linux-amd64" src-tauri/binaries/erp-server-x86_64-unknown-linux-gnu
chmod +x src-tauri/binaries/erp-server-x86_64-unknown-linux-gnu

npm install --silent
npm run build

echo ""
echo "==> Artifacts in $OUT/"
ls -lh "$OUT/"
