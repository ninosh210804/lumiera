package tests

import (
	"context"
	"testing"
	"time"

	"shipment-service/internal/application"
	"shipment-service/internal/domain"
	"shipment-service/internal/infrastructure/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShipment_Success(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	dto := &application.CreateShipmentDTO{
		ID:              domain.GenerateID(),
		ReferenceNumber: "SHIP-001",
		Origin:          "New York",
		Destination:     "Los Angeles",
		DriverName:      "John Doe",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}

	shipmentDTO, err := service.CreateShipmentUseCase(ctx, dto)

	require.NoError(t, err)
	assert.Equal(t, dto.ReferenceNumber, shipmentDTO.ReferenceNum)
	assert.Equal(t, dto.Origin, shipmentDTO.Origin)
	assert.Equal(t, dto.Destination, shipmentDTO.Destination)
	assert.Equal(t, dto.DriverName, shipmentDTO.DriverName)
	assert.Equal(t, dto.VehicleID, shipmentDTO.VehicleID)
	assert.Equal(t, dto.Amount, shipmentDTO.Amount)
	assert.Equal(t, dto.DriverRevenue, shipmentDTO.DriverRevenue)
	assert.Equal(t, int32(domain.StatusPending), shipmentDTO.CurrentStatus)
	assert.Len(t, shipmentDTO.Events, 1)
	assert.Equal(t, int32(domain.StatusPending), shipmentDTO.Events[0].Status)
}

func TestCreateShipment_DuplicateReferenceNumber(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	dto := &application.CreateShipmentDTO{
		ID:              domain.GenerateID(),
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}

	// Create first shipment
	_, err := service.CreateShipmentUseCase(ctx, dto)
	require.NoError(t, err)

	// Try to create another with same reference number
	dto.ID = domain.GenerateID()
	_, err = service.CreateShipmentUseCase(ctx, dto)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReferenceNumberExists, err)
}

// Test GetShipmentUseCase
func TestGetShipment_Success(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create a shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}

	createdShipment, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)

	// Get the shipment
	retrievedShipment, err := service.GetShipmentUseCase(ctx, shipmentID)

	require.NoError(t, err)
	assert.Equal(t, createdShipment.ID, retrievedShipment.ID)
	assert.Equal(t, createdShipment.ReferenceNum, retrievedShipment.ReferenceNum)
	assert.Equal(t, int32(domain.StatusPending), retrievedShipment.CurrentStatus)
}

func TestGetShipment_NotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	_, err := service.GetShipmentUseCase(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrShipmentNotFound, err)
}

// Test AddShipmentEventUseCase - Valid Transitions
func TestAddShipmentEvent_PendingToPickedUp(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create a shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	_, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)

	// Add PICKED_UP event
	addEventDTO := &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusPickedUp),
		Note:       "Package picked up",
		OccurredAt: time.Now(),
	}

	eventDTO, shipmentDTO, err := service.AddShipmentEventUseCase(ctx, addEventDTO)

	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusPickedUp), eventDTO.Status)
	assert.Equal(t, int32(domain.StatusPickedUp), shipmentDTO.CurrentStatus)
	assert.Len(t, shipmentDTO.Events, 2) // Initial PENDING + PICKED_UP
}

func TestAddShipmentEvent_CompleteValidTransitionChain(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create a shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	createdShipment, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusPending), createdShipment.CurrentStatus)

	// PENDING -> PICKED_UP
	addEventDTO := &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusPickedUp),
		Note:       "Picked up",
		OccurredAt: time.Now(),
	}
	_, shipmentDTO, err := service.AddShipmentEventUseCase(ctx, addEventDTO)
	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusPickedUp), shipmentDTO.CurrentStatus)

	// PICKED_UP -> IN_TRANSIT
	addEventDTO = &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusInTransit),
		Note:       "In transit",
		OccurredAt: time.Now(),
	}
	_, shipmentDTO, err = service.AddShipmentEventUseCase(ctx, addEventDTO)
	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusInTransit), shipmentDTO.CurrentStatus)

	// IN_TRANSIT -> DELIVERED
	addEventDTO = &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusDelivered),
		Note:       "Delivered",
		OccurredAt: time.Now(),
	}
	_, shipmentDTO, err = service.AddShipmentEventUseCase(ctx, addEventDTO)
	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusDelivered), shipmentDTO.CurrentStatus)
	assert.Len(t, shipmentDTO.Events, 4) // Initial + 3 transitions
}

// Test AddShipmentEventUseCase - Invalid Transitions
func TestAddShipmentEvent_InvalidTransition(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create a shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	_, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)

	// Try PENDING -> IN_TRANSIT (should fail, must go through PICKED_UP)
	addEventDTO := &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusInTransit),
		Note:       "Invalid transition",
		OccurredAt: time.Now(),
	}

	_, _, err = service.AddShipmentEventUseCase(ctx, addEventDTO)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidTransition, err)
}

func TestAddShipmentEvent_DuplicateStatus(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create a shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	_, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err) // Shipment created with PENDING status

	// Try to add PENDING again (should fail)
	addEventDTO := &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusPending),
		Note:       "Duplicate status",
		OccurredAt: time.Now(),
	}

	_, _, err = service.AddShipmentEventUseCase(ctx, addEventDTO)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrDuplicateStatus, err)
}

func TestAddShipmentEvent_TerminalStatusBlocks(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create and move shipment to DELIVERED status
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	_, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)

	// Progress to DELIVERED
	service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusPickedUp),
		OccurredAt: time.Now(),
	})
	service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusInTransit),
		OccurredAt: time.Now(),
	})
	_, _, err = service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusDelivered),
		OccurredAt: time.Now(),
	})
	require.NoError(t, err)

	// Try to add another event (should fail - terminal)
	_, _, err = service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusCancelled),
		OccurredAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Equal(t, domain.ErrShipmentTerminal, err)
}

func TestAddShipmentEvent_ShipmentNotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	addEventDTO := &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: "non-existent-id",
		Status:     int32(domain.StatusPickedUp),
		OccurredAt: time.Now(),
	}

	_, _, err := service.AddShipmentEventUseCase(ctx, addEventDTO)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrShipmentNotFound, err)
}

func TestAddShipmentEvent_CancelFromDifferentStatuses(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus domain.Status
	}{
		{"Cancel from PENDING", domain.StatusPending},
		{"Cancel from PICKED_UP", domain.StatusPickedUp},
		{"Cancel from IN_TRANSIT", domain.StatusInTransit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMemoryRepository()
			service := application.NewShipmentService(repo)
			ctx := context.Background()

			// Create shipment
			shipmentID := domain.GenerateID()
			createDTO := &application.CreateShipmentDTO{
				ID:              shipmentID,
				ReferenceNumber: "SHIP-" + tt.name,
				Origin:          "NYC",
				Destination:     "LA",
				DriverName:      "John",
				VehicleID:       "VEH-001",
				Amount:          1000.0,
				DriverRevenue:   100.0,
			}
			_, err := service.CreateShipmentUseCase(ctx, createDTO)
			require.NoError(t, err)

			// Move to target status
			if tt.fromStatus != domain.StatusPending {
				service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
					EventID:    domain.GenerateID(),
					ShipmentID: shipmentID,
					Status:     int32(domain.StatusPickedUp),
					OccurredAt: time.Now(),
				})
			}
			if tt.fromStatus == domain.StatusInTransit {
				service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
					EventID:    domain.GenerateID(),
					ShipmentID: shipmentID,
					Status:     int32(domain.StatusInTransit),
					OccurredAt: time.Now(),
				})
			}

			// Cancel from that status
			_, shipmentDTO, err := service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
				EventID:    domain.GenerateID(),
				ShipmentID: shipmentID,
				Status:     int32(domain.StatusCancelled),
				OccurredAt: time.Now(),
			})

			require.NoError(t, err)
			assert.Equal(t, int32(domain.StatusCancelled), shipmentDTO.CurrentStatus)
			assert.True(t, domain.Status(shipmentDTO.CurrentStatus).IsTerminal())
		})
	}
}

// Test ListShipmentEventsUseCase
func TestListShipmentEvents_Success(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	_, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)

	// Add events
	service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusPickedUp),
		Note:       "Picked up",
		OccurredAt: time.Now(),
	})

	service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: shipmentID,
		Status:     int32(domain.StatusInTransit),
		Note:       "In transit",
		OccurredAt: time.Now(),
	})

	// List events
	events, err := service.ListShipmentEventsUseCase(ctx, shipmentID)

	require.NoError(t, err)
	assert.Len(t, events, 3) // Initial PENDING + 2 added events
	assert.Equal(t, int32(domain.StatusPending), events[0].Status)
	assert.Equal(t, int32(domain.StatusPickedUp), events[1].Status)
	assert.Equal(t, int32(domain.StatusInTransit), events[2].Status)
}

func TestListShipmentEvents_EmptyShipmentNotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	_, err := service.ListShipmentEventsUseCase(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrShipmentNotFound, err)
}

func TestListShipmentEvents_NewShipmentHasInitialEvent(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "SHIP-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "John",
		VehicleID:       "VEH-001",
		Amount:          1000.0,
		DriverRevenue:   100.0,
	}
	_, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)

	// List events for newly created shipment
	events, err := service.ListShipmentEventsUseCase(ctx, shipmentID)

	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, int32(domain.StatusPending), events[0].Status)
	assert.Equal(t, "Shipment created", events[0].Note)
}

// Integration test: Full shipment lifecycle
func TestFullShipmentLifecycle(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	ctx := context.Background()

	// Create shipment
	shipmentID := domain.GenerateID()
	createDTO := &application.CreateShipmentDTO{
		ID:              shipmentID,
		ReferenceNumber: "LIFECYCLE-001",
		Origin:          "NYC",
		Destination:     "LA",
		DriverName:      "Jane Smith",
		VehicleID:       "VEH-999",
		Amount:          5000.0,
		DriverRevenue:   500.0,
	}

	shipmentDTO, err := service.CreateShipmentUseCase(ctx, createDTO)
	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusPending), shipmentDTO.CurrentStatus)

	// Verify initial state
	retrievedShipment, err := service.GetShipmentUseCase(ctx, shipmentID)
	require.NoError(t, err)
	assert.Equal(t, "LIFECYCLE-001", retrievedShipment.ReferenceNum)
	assert.Equal(t, "Jane Smith", retrievedShipment.DriverName)

	// Transition through lifecycle
	transitions := []int32{
		int32(domain.StatusPickedUp),
		int32(domain.StatusInTransit),
		int32(domain.StatusDelivered),
	}

	for _, status := range transitions {
		_, shipmentDTO, err := service.AddShipmentEventUseCase(ctx, &application.AddShipmentEventDTO{
			EventID:    domain.GenerateID(),
			ShipmentID: shipmentID,
			Status:     status,
			Note:       "Status update",
			OccurredAt: time.Now(),
		})
		require.NoError(t, err)
		assert.Equal(t, status, shipmentDTO.CurrentStatus)
	}

	// Verify final state
	finalShipment, err := service.GetShipmentUseCase(ctx, shipmentID)
	require.NoError(t, err)
	assert.Equal(t, int32(domain.StatusDelivered), finalShipment.CurrentStatus)

	// List all events
	events, err := service.ListShipmentEventsUseCase(ctx, shipmentID)
	require.NoError(t, err)
	assert.Len(t, events, 4) // Initial + 3 transitions
}
