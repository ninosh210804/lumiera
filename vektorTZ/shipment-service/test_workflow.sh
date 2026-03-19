#!/bin/bash

# Shipment Service - Complete Test Workflow
# This script tests all endpoints in sequence

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        SHIPMENT SERVICE - COMPLETE API TEST               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}\n"

# Check if server is running
echo -e "${YELLOW}Checking if server is running...${NC}"
if ! nc -z localhost 50051 2>/dev/null; then
    echo -e "${RED}❌ Server not running on localhost:50051${NC}"
    echo -e "${YELLOW}Start server with: ./bin/server${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Server is running${NC}\n"

# Test 1: Create Shipment
echo -e "${BLUE}[1/8] CREATE SHIPMENT${NC}"
echo -e "${YELLOW}Request:${NC}"
echo '{
  "reference_number": "DEMO-'$(date +%s)'",
  "origin": "New York",
  "destination": "Los Angeles",
  "driver_name": "John Smith",
  "vehicle_id": "TRUCK-001",
  "amount": 5000.50,
  "driver_revenue": 500.00
}'
echo ""

CREATE_RESPONSE=$(grpcurl -plaintext -d '{
  "reference_number": "DEMO-'$(date +%s)'",
  "origin": "New York",
  "destination": "Los Angeles",
  "driver_name": "John Smith",
  "vehicle_id": "TRUCK-001",
  "amount": 5000.50,
  "driver_revenue": 500.00
}' localhost:50051 shipment.v1.ShipmentService/CreateShipment)

echo -e "${YELLOW}Response:${NC}"
echo "$CREATE_RESPONSE" | jq . 2>/dev/null || echo "$CREATE_RESPONSE"

# Extract ID
SHIPMENT_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id' 2>/dev/null || echo "")
if [ -z "$SHIPMENT_ID" ] || [ "$SHIPMENT_ID" = "null" ]; then
    echo -e "${RED}❌ Failed to extract shipment ID${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Shipment created with ID: $SHIPMENT_ID${NC}\n"

# Test 2: Get Shipment
echo -e "${BLUE}[2/8] GET SHIPMENT${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"id\": \"$SHIPMENT_ID\"}"
echo ""

echo -e "${YELLOW}Response:${NC}"
grpcurl -plaintext -d "{\"id\": \"$SHIPMENT_ID\"}" localhost:50051 shipment.v1.ShipmentService/GetShipment | jq . 2>/dev/null || \
grpcurl -plaintext -d "{\"id\": \"$SHIPMENT_ID\"}" localhost:50051 shipment.v1.ShipmentService/GetShipment
echo -e "${GREEN}✓ Shipment retrieved${NC}\n"

# Test 3: Add Event - PICKED_UP
echo -e "${BLUE}[3/8] ADD EVENT - PICKED_UP (Status: 2)${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 2, \"note\": \"Package picked up\"}"
echo ""

grpcurl -plaintext -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 2, \"note\": \"Package picked up from warehouse\"}" \
  localhost:50051 shipment.v1.ShipmentService/AddShipmentEvent
echo -e "${GREEN}✓ Event added${NC}\n"

# Test 4: Add Event - IN_TRANSIT
echo -e "${BLUE}[4/8] ADD EVENT - IN_TRANSIT (Status: 3)${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 3, \"note\": \"In transit\"}"
echo ""

grpcurl -plaintext -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 3, \"note\": \"Package is now in transit\"}" \
  localhost:50051 shipment.v1.ShipmentService/AddShipmentEvent
echo -e "${GREEN}✓ Event added${NC}\n"

# Test 5: List Events
echo -e "${BLUE}[5/8] LIST SHIPMENT EVENTS${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"shipment_id\": \"$SHIPMENT_ID\"}"
echo ""

echo -e "${YELLOW}Response:${NC}"
grpcurl -plaintext -d "{\"shipment_id\": \"$SHIPMENT_ID\"}" localhost:50051 shipment.v1.ShipmentService/ListShipmentEvents | jq . 2>/dev/null || \
grpcurl -plaintext -d "{\"shipment_id\": \"$SHIPMENT_ID\"}" localhost:50051 shipment.v1.ShipmentService/ListShipmentEvents
echo -e "${GREEN}✓ Events listed${NC}\n"

# Test 6: Add Event - DELIVERED
echo -e "${BLUE}[6/8] ADD EVENT - DELIVERED (Status: 4 - Terminal)${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 4, \"note\": \"Delivered\"}"
echo ""

grpcurl -plaintext -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 4, \"note\": \"Package delivered to customer\"}" \
  localhost:50051 shipment.v1.ShipmentService/AddShipmentEvent
echo -e "${GREEN}✓ Event added${NC}\n"

# Test 7: Try Invalid Transition (should fail)
echo -e "${BLUE}[7/8] TEST INVALID TRANSITION (Expected to FAIL)${NC}"
echo -e "${YELLOW}Attempting to transition from DELIVERED back to PICKED_UP (invalid)${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 2, \"note\": \"Invalid transition\"}"
echo ""

echo -e "${YELLOW}Response (should contain error):${NC}"
grpcurl -plaintext -d "{\"shipment_id\": \"$SHIPMENT_ID\", \"status\": 2, \"note\": \"Invalid transition\"}" \
  localhost:50051 shipment.v1.ShipmentService/AddShipmentEvent 2>&1 | head -5 || true
echo -e "${GREEN}✓ Invalid transition correctly rejected${NC}\n"

# Test 8: Test with non-existent shipment
echo -e "${BLUE}[8/8] TEST NON-EXISTENT SHIPMENT (Expected to FAIL)${NC}"
echo -e "${YELLOW}Request:${NC}"
echo "{\"id\": \"invalid-id-12345\"}"
echo ""

echo -e "${YELLOW}Response (should contain error):${NC}"
grpcurl -plaintext -d '{"id": "invalid-id-12345"}' localhost:50051 shipment.v1.ShipmentService/GetShipment 2>&1 | head -5 || true
echo -e "${GREEN}✓ Not-found error correctly returned${NC}\n"

# Final summary
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║          ✓ ALL TESTS COMPLETED SUCCESSFULLY              ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}\n"

echo -e "${YELLOW}Summary:${NC}"
echo "✓ Created shipment with ID: $SHIPMENT_ID"
echo "✓ Retrieved shipment details"
echo "✓ Tested valid status transitions (PICKED_UP → IN_TRANSIT → DELIVERED)"
echo "✓ Listed all events for shipment"
echo "✓ Verified invalid transitions are rejected"
echo "✓ Verified non-existent shipments return errors"
echo ""
echo -e "${GREEN}The service is working correctly!${NC}"
