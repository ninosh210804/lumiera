.PHONY: dev migrate-up migrate-down migrate-create generate build build-cloud test test-integration lint install-tools seed

# ─── Config ────────────────────────────────────────────────────────────────────
-include ../../.env
DATABASE_URL ?= postgres://postgres:postgres@localhost:5434/coffeeshop?sslmode=disable
PG_MIGRATIONS   = ../../db/migrations/postgres
SQLITE_MIGRATIONS = ../../db/migrations/sqlite

# ─── Development ───────────────────────────────────────────────────────────────
dev:            ## hot-reload server (requires air)
	air -c .air.toml

# ─── Migrations ────────────────────────────────────────────────────────────────
migrate-up:     ## apply all pending PostgreSQL migrations
	migrate -path $(PG_MIGRATIONS) -database "$(DATABASE_URL)" up

migrate-down:   ## roll back last PostgreSQL migration
	migrate -path $(PG_MIGRATIONS) -database "$(DATABASE_URL)" down 1

migrate-force:  ## force migration version (usage: make migrate-force V=3)
	migrate -path $(PG_MIGRATIONS) -database "$(DATABASE_URL)" force $(V)

migrate-create: ## create new migration (usage: make migrate-create NAME=add_foo)
	migrate create -ext sql -dir $(PG_MIGRATIONS) -seq $(NAME)
	migrate create -ext sql -dir $(SQLITE_MIGRATIONS) -seq $(NAME)

migrate-status:
	migrate -path $(PG_MIGRATIONS) -database "$(DATABASE_URL)" version

# ─── Code generation ───────────────────────────────────────────────────────────
generate:       ## run sqlc codegen
	sqlc generate -f sqlc.yaml
	go generate ./...

# ─── Build ─────────────────────────────────────────────────────────────────────
build:          ## cross-compile sidecar binaries for Tauri
	@echo "Building Windows x64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" \
		-o ../../apps/desktop/src-tauri/binaries/server-x86_64-pc-windows-msvc.exe ./cmd/server
	@echo "Building macOS arm64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" \
		-o ../../apps/desktop/src-tauri/binaries/server-aarch64-apple-darwin ./cmd/server
	@echo "Building Linux x64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" \
		-o ../../apps/desktop/src-tauri/binaries/server-x86_64-unknown-linux-gnu ./cmd/server

build-cloud:    ## build for Fly.io deployment
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# ─── Tests ─────────────────────────────────────────────────────────────────────
test:           ## run unit tests
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

test-integration: ## run integration tests (needs running PostgreSQL)
	go test -tags=integration -race ./test/integration/...

# ─── Lint ──────────────────────────────────────────────────────────────────────
lint:           ## run golangci-lint
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

# ─── Tools installation ────────────────────────────────────────────────────────
install-tools:  ## install dev tools (sqlc, air, migrate)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/air-verse/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ─── Seed ──────────────────────────────────────────────────────────────────────
seed:           ## insert demo data (location, admin user, sample menu)
	go run ./test/fixtures/seed.go

help:           ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
