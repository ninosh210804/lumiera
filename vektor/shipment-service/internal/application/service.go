package application

import (
	"context"

	"shipment-service/internal/domain"
)

type ShipmentService struct {
	repository ShipmentRepository
}

func NewShipmentService(repository ShipmentRepository) *ShipmentService {
	return &ShipmentService{
		repository: repository,
	}
}

func (s *ShipmentService) CreateShipmentUseCase(ctx context.Context, dto *CreateShipmentDTO) (*ShipmentDTO, error) {
	_, err := s.repository.GetByReferenceNumber(ctx, dto.ReferenceNumber)
	if err == nil {
		return nil, domain.ErrReferenceNumberExists
	}

	shipment := domain.NewShipment(
		dto.ID,
		dto.ReferenceNumber,
		dto.Origin,
		dto.Destination,
		dto.DriverName,
		dto.VehicleID,
		dto.Amount,
		dto.DriverRevenue,
	)

	// Save to repository
	err = s.repository.Save(ctx, shipment)
	if err != nil {
		return nil, err
	}

	return toShipmentDTO(shipment), nil
}

func (s *ShipmentService) GetShipmentUseCase(ctx context.Context, id string) (*ShipmentDTO, error) {
	shipment, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toShipmentDTO(shipment), nil
}

func (s *ShipmentService) AddShipmentEventUseCase(ctx context.Context, dto *AddShipmentEventDTO) (*ShipmentEventDTO, *ShipmentDTO, error) {
	shipment, err := s.repository.GetByID(ctx, dto.ShipmentID)
	if err != nil {
		return nil, nil, err
	}

	event := domain.NewShipmentEvent(
		dto.EventID,
		dto.ShipmentID,
		domain.Status(dto.Status),
		dto.Note,
		dto.OccurredAt,
	)

	err = shipment.AddEvent(event)
	if err != nil {
		return nil, nil, err
	}

	err = s.repository.Save(ctx, shipment)
	if err != nil {
		return nil, nil, err
	}

	return toShipmentEventDTO(event), toShipmentDTO(shipment), nil
}

func (s *ShipmentService) ListShipmentEventsUseCase(ctx context.Context, shipmentID string) ([]*ShipmentEventDTO, error) {
	shipment, err := s.repository.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	events := make([]*ShipmentEventDTO, len(shipment.Events))
	for i, event := range shipment.Events {
		events[i] = toShipmentEventDTO(event)
	}

	return events, nil
}

func toShipmentDTO(shipment *domain.Shipment) *ShipmentDTO {
	events := make([]*ShipmentEventDTO, len(shipment.Events))
	for i, event := range shipment.Events {
		events[i] = toShipmentEventDTO(event)
	}

	return &ShipmentDTO{
		ID:            shipment.ID,
		ReferenceNum:  shipment.ReferenceNum,
		Origin:        shipment.Origin,
		Destination:   shipment.Destination,
		DriverName:    shipment.DriverName,
		VehicleID:     shipment.VehicleID,
		Amount:        shipment.Amount,
		DriverRevenue: shipment.DriverRevenue,
		CurrentStatus: int32(shipment.CurrentStatus),
		Events:        events,
		CreatedAt:     shipment.CreatedAt,
		UpdatedAt:     shipment.UpdatedAt,
	}
}

func toShipmentEventDTO(event *domain.ShipmentEvent) *ShipmentEventDTO {
	return &ShipmentEventDTO{
		ID:         event.ID,
		ShipmentID: event.ShipmentID,
		Status:     int32(event.Status),
		Note:       event.Note,
		OccurredAt: event.OccurredAt,
	}
}
