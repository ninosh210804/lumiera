#!/bin/bash
# setup.sh - Setup script for Shipment Service

set -e

echo "📦 Shipment Service Setup"
echo "========================="
echo ""

# Check Go version
echo "✓ Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "  Go version: $GO_VERSION"

# Download dependencies
echo ""
echo "✓ Downloading dependencies..."
go mod download

# Tidy up go.mod and go.sum
echo "✓ Tidying go modules..."
go mod tidy

# Build the project
echo ""
echo "✓ Building server..."
go build -o bin/server ./cmd/server
echo "  Binary: bin/server"

# Run tests
echo ""
echo "✓ Running tests..."
go test -v ./tests/...

echo ""
echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo "  1. Start server:  go run ./cmd/server"
echo "  2. Run tests:     go test -v ./..."
echo "  3. Build binary:  go build -o bin/server ./cmd/server"
echo "  4. Docker build:  docker build -t shipment-service:latest ."
echo "  5. Docker compose: docker-compose up -d"
echo ""
echo "🔗 gRPC Server: localhost:50051"
