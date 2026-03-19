# Shipment Service - Testing Guide

## 1. RUN UNIT TESTS

### Run all tests
```bash
go test -v ./tests/
```

### Run specific test
```bash
go test -v ./tests/ -run TestNewShipment
```

### Run with coverage report
```bash
go test -v -cover ./tests/
```

### Run with detailed coverage
```bash
go test -coverprofile=coverage.out ./tests/
go tool cover -html=coverage.out
```

---

## 2. BUILD & START SERVER

### Build binary
```bash
go build -o bin/server ./cmd/server
```

### Run server (will listen on :50051)
```bash
./bin/server
```

### Or run directly
```bash
go run ./cmd/server/main.go
```

---

## 3. GRPC ENDPOINT TESTING WITH GRPCURL

### First, install grpcurl if not already installed
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### List available services
```bash
grpcurl -plaintext localhost:50051 list
```

### Describe the service
```bash
grpcurl -plaintext localhost:50051 describe shipment.v1.ShipmentService
```

---

## 4. TEST ENDPOINTS WITH REQUEST BODIES

### A. CREATE SHIPMENT
```bash
grpcurl -plaintext \
  -d '{
    "reference_number": "SHIP-2024-001",
    "origin": "New York",
    "destination": "Los Angeles",
    "driver_name": "John Doe",
    "vehicle_id": "VEH-001",
    "amount": 1000.50,
    "driver_revenue": 100.00
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/CreateShipment
```

**Expected Response:**
```json
{
  "id": "unique-id-here",
  "reference_number": "SHIP-2024-001",
  "origin": "New York",
  "destination": "Los Angeles",
  "driver_name": "John Doe",
  "vehicle_id": "VEH-001",
  "amount": 1000.5,
  "driver_revenue": 100,
  "current_status": "PENDING",
  "created_at": "2024-03-19T...",
  "updated_at": "2024-03-19T..."
}
```

---

### B. GET SHIPMENT
```bash
# Replace {SHIPMENT_ID} with actual ID from CreateShipment response
grpcurl -plaintext \
  -d '{"id": "{SHIPMENT_ID}"}' \
  localhost:50051 \
  shipment.v1.ShipmentService/GetShipment
```

**Expected Response:**
```json
{
  "id": "{SHIPMENT_ID}",
  "reference_number": "SHIP-2024-001",
  ...
}
```

---

### C. ADD SHIPMENT EVENT (Update Status)

#### Status Codes:
- 1 = PENDING (initial state)
- 2 = PICKED_UP
- 3 = IN_TRANSIT
- 4 = DELIVERED (terminal)
- 5 = CANCELLED (terminal)

#### Example: Move from PENDING → PICKED_UP
```bash
grpcurl -plaintext \
  -d '{
    "shipment_id": "{SHIPMENT_ID}",
    "status": 2,
    "note": "Package picked up from warehouse"
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent
```

#### Example: Move from PICKED_UP → IN_TRANSIT
```bash
grpcurl -plaintext \
  -d '{
    "shipment_id": "{SHIPMENT_ID}",
    "status": 3,
    "note": "Package is now in transit"
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent
```

#### Example: Move from IN_TRANSIT → DELIVERED
```bash
grpcurl -plaintext \
  -d '{
    "shipment_id": "{SHIPMENT_ID}",
    "status": 4,
    "note": "Package delivered successfully"
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent
```

**Expected Response on Success:**
```json
{}
```

**Expected Error on Invalid Transition:**
```json
{
  "code": 9,
  "message": "rpc error: code = FailedPrecondition desc = invalid transition"
}
```

---

### D. LIST SHIPMENT EVENTS
```bash
grpcurl -plaintext \
  -d '{"shipment_id": "{SHIPMENT_ID}"}' \
  localhost:50051 \
  shipment.v1.ShipmentService/ListShipmentEvents
```

**Expected Response:**
```json
{
  "events": [
    {
      "id": "event-1",
      "shipment_id": "{SHIPMENT_ID}",
      "status": "PENDING",
      "note": "",
      "occurred_at": "2024-03-19T..."
    },
    {
      "id": "event-2",
      "shipment_id": "{SHIPMENT_ID}",
      "status": "PICKED_UP",
      "note": "Package picked up from warehouse",
      "occurred_at": "2024-03-19T..."
    }
  ]
}
```

---

## 5. FULL WORKFLOW EXAMPLE

```bash
#!/bin/bash

# Start server in background
./bin/server &
SERVER_PID=$!
sleep 1

# 1. Create shipment
echo "=== Creating Shipment ==="
RESPONSE=$(grpcurl -plaintext \
  -d '{
    "reference_number": "DEMO-001",
    "origin": "NYC",
    "destination": "LA",
    "driver_name": "Alice",
    "vehicle_id": "TRUCK-01",
    "amount": 5000,
    "driver_revenue": 500
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/CreateShipment)

echo "$RESPONSE"

# Extract ID (adjust parsing based on actual response)
SHIPMENT_ID=$(echo "$RESPONSE" | grep '"id"' | head -1 | cut -d'"' -f4)
echo "Created Shipment ID: $SHIPMENT_ID"

# 2. Get shipment
echo -e "\n=== Getting Shipment ==="
grpcurl -plaintext \
  -d "{\"id\": \"$SHIPMENT_ID\"}" \
  localhost:50051 \
  shipment.v1.ShipmentService/GetShipment

# 3. Add event - PICKED_UP
echo -e "\n=== Moving to PICKED_UP ==="
grpcurl -plaintext \
  -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 2, \"note\": \"Picked up\"}" \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent

# 4. Add event - IN_TRANSIT
echo -e "\n=== Moving to IN_TRANSIT ==="
grpcurl -plaintext \
  -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 3, \"note\": \"In transit\"}" \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent

# 5. List events
echo -e "\n=== Listing All Events ==="
grpcurl -plaintext \
  -d "{\"shipment_id\": \"$SHIPMENT_ID\"}" \
  localhost:50051 \
  shipment.v1.ShipmentService/ListShipmentEvents

# 6. Add event - DELIVERED
echo -e "\n=== Moving to DELIVERED ==="
grpcurl -plaintext \
  -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 4, \"note\": \"Delivered\"}" \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent

# 7. Try invalid transition (should fail)
echo -e "\n=== Attempting Invalid Transition (should fail) ==="
grpcurl -plaintext \
  -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 2, \"note\": \"Invalid\"}" \
  localhost:50051 \
  shipment.v1.ShipmentService/AddShipmentEvent

# Cleanup
kill $SERVER_PID
```

---

## 6. ERROR TESTING

### Test duplicate reference number
```bash
# Create first shipment
grpcurl -plaintext \
  -d '{
    "reference_number": "UNIQUE-REF",
    "origin": "NYC",
    "destination": "LA",
    "driver_name": "Driver1",
    "vehicle_id": "VEH1",
    "amount": 1000,
    "driver_revenue": 100
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/CreateShipment

# Try creating another with same reference (should fail with AlreadyExists)
grpcurl -plaintext \
  -d '{
    "reference_number": "UNIQUE-REF",
    "origin": "NYC",
    "destination": "LA",
    "driver_name": "Driver2",
    "vehicle_id": "VEH2",
    "amount": 2000,
    "driver_revenue": 200
  }' \
  localhost:50051 \
  shipment.v1.ShipmentService/CreateShipment
```

### Test non-existent shipment
```bash
grpcurl -plaintext \
  -d '{"id": "invalid-id-12345"}' \
  localhost:50051 \
  shipment.v1.ShipmentService/GetShipment
```

---

## 7. DOCKER TESTING

### Build Docker image
```bash
docker build -t shipment-service:latest .
```

### Run with Docker
```bash
docker run -p 50051:50051 shipment-service:latest
```

### Run with Docker Compose
```bash
docker-compose up -d
```

### View logs
```bash
docker-compose logs -f shipment-service
```

### Stop services
```bash
docker-compose down
```

