package tests

import (
	"testing"
	"time"

	"shipment-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewShipment_StartsWithPendingStatus(t *testing.T) {
	shipment := domain.NewShipment(
		"test-id",
		"REF-001",
		"New York",
		"Los Angeles",
		"John Doe",
		"VEH-001",
		1000.0,
		100.0,
	)

	assert.Equal(t, domain.StatusPending, shipment.CurrentStatus)
	assert.Len(t, shipment.Events, 1)
	assert.Equal(t, domain.StatusPending, shipment.Events[0].Status)
}

func TestShipment_ValidTransitions(t *testing.T) {
	shipment := domain.NewShipment("id", "REF", "A", "B", "Driver", "VEH", 100, 10)

	err := shipment.AddEvent(domain.NewShipmentEvent("event-1", shipment.ID, domain.StatusPickedUp, "Picked up", time.Now()))
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusPickedUp, shipment.CurrentStatus)

	err = shipment.AddEvent(domain.NewShipmentEvent("event-2", shipment.ID, domain.StatusInTransit, "In transit", time.Now()))
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusInTransit, shipment.CurrentStatus)

	err = shipment.AddEvent(domain.NewShipmentEvent("event-3", shipment.ID, domain.StatusDelivered, "Delivered", time.Now()))
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusDelivered, shipment.CurrentStatus)
}

func TestShipment_InvalidTransition(t *testing.T) {
	shipment := domain.NewShipment("id", "REF", "A", "B", "Driver", "VEH", 100, 10)

	err := shipment.AddEvent(domain.NewShipmentEvent("event-1", shipment.ID, domain.StatusInTransit, "Invalid", time.Now()))
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidTransition, err)
	assert.Equal(t, domain.StatusPending, shipment.CurrentStatus)
}

func TestShipment_TerminalStatusBlocks(t *testing.T) {
	shipment := domain.NewShipment("id", "REF", "A", "B", "Driver", "VEH", 100, 10)

	shipment.AddEvent(domain.NewShipmentEvent("e1", shipment.ID, domain.StatusPickedUp, "", time.Now()))
	shipment.AddEvent(domain.NewShipmentEvent("e2", shipment.ID, domain.StatusInTransit, "", time.Now()))
	shipment.AddEvent(domain.NewShipmentEvent("e3", shipment.ID, domain.StatusDelivered, "", time.Now()))

	err := shipment.AddEvent(domain.NewShipmentEvent("e4", shipment.ID, domain.StatusCancelled, "", time.Now()))
	assert.Error(t, err)
	assert.Equal(t, domain.ErrShipmentTerminal, err)
}

func TestShipment_DuplicateStatusBlocks(t *testing.T) {
	shipment := domain.NewShipment("id", "REF", "A", "B", "Driver", "VEH", 100, 10)

	err := shipment.AddEvent(domain.NewShipmentEvent("event-1", shipment.ID, domain.StatusPending, "Duplicate", time.Now()))
	assert.Error(t, err)
	assert.Equal(t, domain.ErrDuplicateStatus, err)
}

func TestShipment_CancelledIsTerminal(t *testing.T) {
	shipment := domain.NewShipment("id", "REF", "A", "B", "Driver", "VEH", 100, 10)

	shipment.AddEvent(domain.NewShipmentEvent("e1", shipment.ID, domain.StatusPickedUp, "", time.Now()))
	err := shipment.AddEvent(domain.NewShipmentEvent("e2", shipment.ID, domain.StatusCancelled, "", time.Now()))
	assert.NoError(t, err)
	assert.True(t, shipment.CurrentStatus.IsTerminal())

	err = shipment.AddEvent(domain.NewShipmentEvent("e3", shipment.ID, domain.StatusInTransit, "", time.Now()))
	assert.Error(t, err)
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.Status
		to       domain.Status
		expected bool
	}{
		{"PENDING to PICKED_UP", domain.StatusPending, domain.StatusPickedUp, true},
		{"PENDING to IN_TRANSIT", domain.StatusPending, domain.StatusInTransit, false},
		{"PENDING to CANCELLED", domain.StatusPending, domain.StatusCancelled, true},
		{"PICKED_UP to PENDING", domain.StatusPickedUp, domain.StatusPending, false},
		{"PICKED_UP to IN_TRANSIT", domain.StatusPickedUp, domain.StatusInTransit, true},
		{"PICKED_UP to CANCELLED", domain.StatusPickedUp, domain.StatusCancelled, true},
		{"IN_TRANSIT to DELIVERED", domain.StatusInTransit, domain.StatusDelivered, true},
		{"IN_TRANSIT to PENDING", domain.StatusInTransit, domain.StatusPending, false},
		{"DELIVERED to PENDING", domain.StatusDelivered, domain.StatusPending, false},
		{"DELIVERED to CANCELLED", domain.StatusDelivered, domain.StatusCancelled, false},
		{"CANCELLED to PENDING", domain.StatusCancelled, domain.StatusPending, false},
		{"CANCELLED to DELIVERED", domain.StatusCancelled, domain.StatusDelivered, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.expected, result, "%s: expected %v", tt.name, tt.expected)
		})
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   domain.Status
		expected bool
	}{
		{domain.StatusPending, false},
		{domain.StatusPickedUp, false},
		{domain.StatusInTransit, false},
		{domain.StatusDelivered, true},
		{domain.StatusCancelled, true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.status.IsTerminal(), "Status %v", tt.status)
	}
}
