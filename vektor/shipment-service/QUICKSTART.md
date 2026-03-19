# Shipment Service - Quick Reference Guide

## 🚀 Getting Started (30 seconds)

```bash
cd shipment-service
chmod +x setup.sh
./setup.sh
```

## 📝 Common Commands

```bash
# Run server locally
go run ./cmd/server

# Run tests
go test -v ./tests/...

# Build binary
go build -o bin/server ./cmd/server

# Run with Docker
docker-compose up -d

# View Docker logs
docker-compose logs -f shipment-service

# Stop Docker services
docker-compose down

# Clean build artifacts
make clean

# Full help
make help
```

## 🔌 gRPC API Examples

Server runs on `localhost:50051`

### Using grpcurl

Install: `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`

**Create Shipment:**
```bash
grpcurl -plaintext \
  -d '{
    "reference_number":"SHIP-2024-001",
    "origin":"New York",
    "destination":"Los Angeles",
    "driver_name":"John Doe",
    "vehicle_id":"VEH-001",
    "amount":1000.0,
    "driver_revenue":100.0
  }' \
  localhost:50051 shipment.v1.ShipmentService/CreateShipment
```

**Get Shipment:**
```bash
grpcurl -plaintext \
  -d '{"id":"<shipment-id>"}' \
  localhost:50051 shipment.v1.ShipmentService/GetShipment
```

**Add Event (PENDING → PICKED_UP):**
```bash
grpcurl -plaintext \
  -d '{
    "shipment_id":"<shipment-id>",
    "status":2,
    "note":"Package picked up from warehouse"
  }' \
  localhost:50051 shipment.v1.ShipmentService/AddShipmentEvent
```

**List Events:**
```bash
grpcurl -plaintext \
  -d '{"shipment_id":"<shipment-id>"}' \
  localhost:50051 shipment.v1.ShipmentService/ListShipmentEvents
```

## 📊 Status Codes Reference

| Code | Name | Description |
|------|------|-------------|
| 1 | PENDING | Initial state, ready for pickup |
| 2 | PICKED_UP | Package collected from origin |
| 3 | IN_TRANSIT | Currently being delivered |
| 4 | DELIVERED | Reached destination (terminal) |
| 5 | CANCELLED | Shipment cancelled (terminal) |

## ✅ Valid Transitions

```
PENDING     → PICKED_UP, CANCELLED
PICKED_UP   → IN_TRANSIT, CANCELLED
IN_TRANSIT  → DELIVERED, CANCELLED
DELIVERED   (terminal)
CANCELLED   (terminal)
```

## 📁 Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/server/main.go` | Server entry point | 38 |
| `internal/domain/shipment.go` | Shipment entity | 64 |
| `internal/domain/status.go` | Status state machine | 52 |
| `internal/application/service.go` | Business logic | 104 |
| `internal/infrastructure/grpc/handler.go` | gRPC implementation | 192 |
| `internal/infrastructure/repository/memory.go` | Data storage | 80 |
| `api/proto/shipment.proto` | API contract | 77 |
| `tests/domain_test.go` | Domain tests | 271 |
| `tests/application_test.go` | Use case tests | 225 |

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Find process using port 50051
lsof -i :50051

# Kill process (if needed)
kill -9 <PID>
```

### Dependencies Not Found
```bash
# Refresh go.mod and go.sum
go mod tidy

# Download all dependencies
go mod download
```

### Docker Build Issues
```bash
# Clean build (no cache)
docker build --no-cache -t shipment-service:latest .

# Check logs
docker-compose logs -f
```

## 🧪 Running Specific Tests

```bash
# Run only domain tests
go test -v ./tests/ -run TestNewShipment

# Run only application tests
go test -v ./tests/ -run TestCreateShipmentUseCase

# Run with coverage
go test -v -cover ./tests/...

# Generate coverage report
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out
```

## 📋 Makefile Targets

```bash
make proto         # Generate protobuf code
make build         # Build server binary
make run           # Run server
make test          # Run tests
make test-coverage # Tests with coverage
make clean         # Clean artifacts
make docker-build  # Build Docker image
make docker-run    # Run Docker container
make docker-up     # Start with docker-compose
make docker-down   # Stop docker-compose
make docker-logs   # View docker-compose logs
make help          # Show all targets
```

## 🔐 Error Codes

| gRPC Code | Meaning | Example |
|-----------|---------|---------|
| 3 | InvalidArgument | Missing required field |
| 5 | NotFound | Shipment doesn't exist |
| 6 | AlreadyExists | Reference number duplicate |
| 9 | FailedPrecondition | Status is terminal |
| 13 | Internal | Unexpected error |

## 📚 Architecture Layers

```
┌─────────────────────────────────────┐
│ gRPC Handler (Request/Response)     │ ← Infrastructure
├─────────────────────────────────────┤
│ Service Layer (Use Cases)           │ ← Application
├─────────────────────────────────────┤
│ Entity Logic (State Machine)        │ ← Domain
├─────────────────────────────────────┤
│ Repository (Data Access)            │ ← Infrastructure
└─────────────────────────────────────┘
```

## 🎯 Project Statistics

- **Total Lines**: 2,581
- **Go Files**: 10
- **Test Files**: 2
- **Test Cases**: 26+
- **API Endpoints**: 4
- **Domain Entities**: 2
- **Error Types**: 5
- **Status States**: 5

## 📖 Documentation Files

- `README.md` - Full documentation (350+ lines)
- `IMPLEMENTATION.md` - Implementation details
- `setup.sh` - Automated setup script
- This file - Quick reference

## 💡 Next Steps

1. **Local Development**
   ```bash
   chmod +x setup.sh && ./setup.sh
   go run ./cmd/server
   ```

2. **Testing**
   ```bash
   go test -v ./tests/...
   ```

3. **Docker Deployment**
   ```bash
   docker-compose up -d
   ```

4. **Production Setup**
   - Add database persistence
   - Implement proper logging
   - Add metrics/monitoring
   - Setup CI/CD pipeline

## 🆘 Need Help?

- Check `README.md` for detailed documentation
- Review `IMPLEMENTATION.md` for architecture details
- Look at test files for usage examples
- See error messages for specific issues

---

**Service Status: Ready for Development & Testing ✅**
