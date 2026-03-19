# Shipment Tracking gRPC Microservice

A production-ready gRPC microservice for tracking shipments, built with Go following Clean Architecture principles.

## Overview

This service provides a complete shipment tracking solution with:
- Event-driven shipment status lifecycle management
- Type-safe gRPC API
- Domain-driven design with clean separation of concerns
- Comprehensive test coverage
- Docker support

## Project Structure

```
shipment-service/
├── api/
│   └── proto/                      # Protocol Buffer definitions
│       ├── shipment.proto
│       └── shipmentpb/            # Generated gRPC code
│           ├── shipment.pb.go
│           └── shipment_grpc.pb.go
├── cmd/
│   └── server/
│       └── main.go                # Server entry point
├── internal/
│   ├── domain/                    # Domain layer (entities, value objects, business logic)
│   │   ├── shipment.go            # Shipment entity
│   │   ├── status.go              # Status enum and transitions
│   │   ├── event.go               # ShipmentEvent entity
│   │   └── errors.go              # Domain-specific errors
│   ├── application/               # Application layer (use cases)
│   │   ├── service.go             # Business logic orchestration
│   │   ├── dto.go                 # Data transfer objects
│   │   └── ports.go               # Repository interface definitions
│   └── infrastructure/            # Infrastructure layer (implementations)
│       ├── grpc/
│       │   └── handler.go         # gRPC service handler
│       └── repository/
│           └── memory.go          # In-memory repository implementation
├── tests/
│   ├── domain_test.go             # Domain logic tests
│   └── application_test.go        # Use case tests
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build automation
├── Dockerfile                     # Container image definition
├── docker-compose.yaml            # Docker orchestration
└── README.md                      # This file
```

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────┐
│  Infrastructure (gRPC, Repository)  │
├─────────────────────────────────────┤
│  Application (Service, DTOs)        │
├─────────────────────────────────────┤
│  Domain (Entities, Value Objects)   │
└─────────────────────────────────────┘

Dependency direction: Infrastructure → Application → Domain
Domain has zero external dependencies.
```

### Shipment Status Lifecycle

```
                    ┌─────────┐
                    │ PENDING │
                    └────┬────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
    ┌───────────┐  ┌──────────┐  ┌──────────────┐
    │ PICKED_UP │  │CANCELLED │  │  (invalid)   │
    └─────┬─────┘  └──────────┘  └──────────────┘
          │
          │
    ┌─────▼──────┐
    │ IN_TRANSIT │
    └─────┬──────┘
          │
    ┌─────┴─────────────┐
    │                   │
    ▼                   ▼
┌──────────┐      ┌──────────────┐
│DELIVERED │      │  CANCELLED   │
│(terminal)│      │  (terminal)  │
└──────────┘      └──────────────┘
```

### Domain Rules

1. **New shipments** start with `PENDING` status
2. **Valid transitions** only:
   - `PENDING` → `PICKED_UP` or `CANCELLED`
   - `PICKED_UP` → `IN_TRANSIT` or `CANCELLED`
   - `IN_TRANSIT` → `DELIVERED` or `CANCELLED`
   - `DELIVERED` and `CANCELLED` are terminal (no transitions)

3. **Each status change** creates a `ShipmentEvent`
4. **CurrentStatus** always reflects the latest event
5. **Duplicate statuses** are rejected (can't add event with same status as current)
6. **Terminal statuses** block further updates

## API Specification

### gRPC Service Definition

```protobuf
service ShipmentService {
  rpc CreateShipment(CreateShipmentRequest) returns (CreateShipmentResponse);
  rpc GetShipment(GetShipmentRequest) returns (GetShipmentResponse);
  rpc AddShipmentEvent(AddShipmentEventRequest) returns (AddShipmentEventResponse);
  rpc ListShipmentEvents(ListShipmentEventsRequest) returns (ListShipmentEventsResponse);
}
```

### Message Types

**Shipment**
```
{
  id: string (UUID)
  reference_number: string (unique)
  origin: string
  destination: string
  driver_name: string
  vehicle_id: string
  amount: float64
  driver_revenue: float64
  current_status: Status enum
  created_at: Timestamp
  updated_at: Timestamp
}
```

**ShipmentEvent**
```
{
  id: string (UUID)
  shipment_id: string
  status: Status enum
  note: string
  occurred_at: Timestamp
}
```

**Status Enum**
```
UNSPECIFIED (0)
PENDING (1)
PICKED_UP (2)
IN_TRANSIT (3)
DELIVERED (4)
CANCELLED (5)
```

## Getting Started

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose (optional)
- `protoc` compiler (for regenerating protobuf files)

### Local Setup

1. **Clone the repository**
   ```bash
   git clone <repo-url>
   cd shipment-service
   ```

2. **Download dependencies**
   ```bash
   go mod download
   ```

3. **Build the server**
   ```bash
   make build
   ```

4. **Run the server**
   ```bash
   make run
   ```

Server will listen on `localhost:50051`

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

### Docker Setup

1. **Build Docker image**
   ```bash
   make docker-build
   ```

2. **Run with docker-compose**
   ```bash
   make docker-up
   ```

3. **View logs**
   ```bash
   make docker-logs
   ```

4. **Stop services**
   ```bash
   make docker-down
   ```

## Usage Examples

### Using grpcurl

Install grpcurl: `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`

**Create a shipment**
```bash
grpcurl -plaintext \
  -d '{"reference_number":"SHIP-001","origin":"NYC","destination":"LA","driver_name":"John","vehicle_id":"VH-001","amount":1000.0,"driver_revenue":100.0}' \
  localhost:50051 shipment.v1.ShipmentService/CreateShipment
```

**Get a shipment**
```bash
grpcurl -plaintext \
  -d '{"id":"<shipment-id>"}' \
  localhost:50051 shipment.v1.ShipmentService/GetShipment
```

**Add shipment event**
```bash
grpcurl -plaintext \
  -d '{"shipment_id":"<shipment-id>","status":2,"note":"Package picked up"}' \
  localhost:50051 shipment.v1.ShipmentService/AddShipmentEvent
```

**List shipment events**
```bash
grpcurl -plaintext \
  -d '{"shipment_id":"<shipment-id>"}' \
  localhost:50051 shipment.v1.ShipmentService/ListShipmentEvents
```

## Testing

### Domain Tests
Located in `tests/domain_test.go`:
- Shipment creation with PENDING status
- Valid status transitions
- Invalid transition rejection
- Terminal status blocking
- Duplicate status rejection
- Status lifecycle rules

### Application Tests
Located in `tests/application_test.go`:
- CreateShipment use case
- GetShipment use case
- AddShipmentEvent with valid/invalid transitions
- ListShipmentEvents
- Error handling for all scenarios

Run tests with:
```bash
go test -v ./tests/...
```

## Error Handling

The service returns appropriate gRPC error codes:

| Error | gRPC Code | Description |
|-------|-----------|-------------|
| ShipmentNotFound | NotFound (5) | Shipment doesn't exist |
| InvalidTransition | InvalidArgument (3) | Invalid status transition |
| ShipmentTerminal | FailedPrecondition (9) | Shipment reached terminal status |
| DuplicateStatus | InvalidArgument (3) | Same status as current |
| ReferenceNumberExists | AlreadyExists (6) | Duplicate reference number |

## Design Decisions

### Clean Architecture
- **Domain Layer**: Pure Go with no external dependencies. Contains entities, value objects, and business rules.
- **Application Layer**: Orchestrates domain logic. Defines ports (interfaces) for external dependencies.
- **Infrastructure Layer**: Implements ports. Contains gRPC handlers and persistence.

### In-Memory Repository
- Thread-safe with `sync.RWMutex`
- Suitable for testing and small deployments
- Can be replaced with database implementation without changing domain/application layers

### Event Sourcing Pattern
- Each status change creates an immutable event
- Complete audit trail of shipment lifecycle
- Enables temporal queries and analytics

### UUID for IDs
- Globally unique identifiers using `github.com/google/uuid`
- Safe for distributed systems
- No database dependency for ID generation

## Extending the Service

### Adding a New Repository Implementation

1. Implement the `ShipmentRepository` interface:
   ```go
   type ShipmentRepository interface {
       Save(ctx context.Context, shipment *domain.Shipment) error
       GetByID(ctx context.Context, id string) (*domain.Shipment, error)
       GetByReferenceNumber(ctx context.Context, referenceNumber string) (*domain.Shipment, error)
       List(ctx context.Context) ([]*domain.Shipment, error)
       Delete(ctx context.Context, id string) error
   }
   ```

2. Replace in `cmd/server/main.go`:
   ```go
   repo := postgresql.NewPostgresRepository(connString)
   ```

### Adding Business Logic

1. Add methods to `ShipmentService` in `internal/application/service.go`
2. Call domain methods to enforce rules
3. Use repository to persist changes
4. Add corresponding gRPC handler in `internal/infrastructure/grpc/handler.go`

## Production Considerations

- **Logging**: Add structured logging with `github.com/sirupsen/logrus` or similar
- **Metrics**: Integrate Prometheus for monitoring
- **Tracing**: Add OpenTelemetry for distributed tracing
- **Database**: Replace in-memory repository with PostgreSQL/MongoDB implementation
- **Caching**: Add Redis caching layer
- **Rate Limiting**: Implement gRPC rate limiting
- **Authentication**: Add JWT/OAuth2 authentication
- **Validation**: Add gRPC validators for requests
- **Documentation**: Generate API docs with protoc-gen-doc

## Makefile Targets

```bash
make proto         # Generate protobuf files
make build         # Build the server binary
make run           # Run the server locally
make test          # Run all tests
make test-coverage # Run tests with coverage report
make clean         # Clean build artifacts
make docker-build  # Build Docker image
make docker-run    # Run in Docker
make docker-up     # Start with docker-compose
make docker-down   # Stop docker-compose
make docker-logs   # View docker-compose logs
make help          # Show help
```

## Dependencies

- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers
- `github.com/google/uuid` - UUID generation
- `github.com/stretchr/testify` - Testing utilities

## Contributing

1. Maintain clean architecture boundaries
2. Add tests for new features
3. Follow Go code style conventions
4. Update documentation

## License

This project is provided as-is for demonstration purposes.

## Contact

For questions or issues, please contact the development team.
