package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"shipment-service/api/proto/shipmentpb"
	"shipment-service/internal/application"
	"shipment-service/internal/domain"
)

type ShipmentHandler struct {
	service *application.ShipmentService
	shipmentpb.UnimplementedShipmentServiceServer
}

func NewShipmentHandler(service *application.ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{
		service: service,
	}
}

func (h *ShipmentHandler) CreateShipment(ctx context.Context, req *shipmentpb.CreateShipmentRequest) (*shipmentpb.CreateShipmentResponse, error) {
	if req.ReferenceNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "reference_number is required")
	}
	if req.Origin == "" {
		return nil, status.Error(codes.InvalidArgument, "origin is required")
	}
	if req.Destination == "" {
		return nil, status.Error(codes.InvalidArgument, "destination is required")
	}
	if req.DriverName == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_name is required")
	}
	if req.VehicleId == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_id is required")
	}

	dto := &application.CreateShipmentDTO{
		ID:              domain.GenerateID(),
		ReferenceNumber: req.ReferenceNumber,
		Origin:          req.Origin,
		Destination:     req.Destination,
		DriverName:      req.DriverName,
		VehicleID:       req.VehicleId,
		Amount:          req.Amount,
		DriverRevenue:   req.DriverRevenue,
	}

	// Call service
	shipmentDTO, err := h.service.CreateShipmentUseCase(ctx, dto)
	if err != nil {
		return nil, h.mapDomainErrorToGRPC(err)
	}

	return &shipmentpb.CreateShipmentResponse{
		Shipment: h.shipmentDTOToProto(shipmentDTO),
	}, nil
}

func (h *ShipmentHandler) GetShipment(ctx context.Context, req *shipmentpb.GetShipmentRequest) (*shipmentpb.GetShipmentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	shipmentDTO, err := h.service.GetShipmentUseCase(ctx, req.Id)
	if err != nil {
		return nil, h.mapDomainErrorToGRPC(err)
	}

	return &shipmentpb.GetShipmentResponse{
		Shipment: h.shipmentDTOToProto(shipmentDTO),
	}, nil
}

func (h *ShipmentHandler) AddShipmentEvent(ctx context.Context, req *shipmentpb.AddShipmentEventRequest) (*shipmentpb.AddShipmentEventResponse, error) {
	if req.ShipmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "shipment_id is required")
	}
	if req.Status == shipmentpb.Status_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	dto := &application.AddShipmentEventDTO{
		EventID:    domain.GenerateID(),
		ShipmentID: req.ShipmentId,
		Status:     int32(req.Status),
		Note:       req.Note,
		OccurredAt: time.Now(),
	}

	eventDTO, shipmentDTO, err := h.service.AddShipmentEventUseCase(ctx, dto)
	if err != nil {
		return nil, h.mapDomainErrorToGRPC(err)
	}

	return &shipmentpb.AddShipmentEventResponse{
		Event:    h.shipmentEventDTOToProto(eventDTO),
		Shipment: h.shipmentDTOToProto(shipmentDTO),
	}, nil
}

func (h *ShipmentHandler) ListShipmentEvents(ctx context.Context, req *shipmentpb.ListShipmentEventsRequest) (*shipmentpb.ListShipmentEventsResponse, error) {
	if req.ShipmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "shipment_id is required")
	}

	events, err := h.service.ListShipmentEventsUseCase(ctx, req.ShipmentId)
	if err != nil {
		return nil, h.mapDomainErrorToGRPC(err)
	}

	protoEvents := make([]*shipmentpb.ShipmentEvent, len(events))
	for i, event := range events {
		protoEvents[i] = h.shipmentEventDTOToProto(event)
	}

	return &shipmentpb.ListShipmentEventsResponse{
		Events: protoEvents,
	}, nil
}

func (h *ShipmentHandler) shipmentDTOToProto(dto *application.ShipmentDTO) *shipmentpb.Shipment {
	if dto == nil {
		return nil
	}

	events := make([]*shipmentpb.ShipmentEvent, len(dto.Events))
	for i, event := range dto.Events {
		events[i] = h.shipmentEventDTOToProto(event)
	}

	return &shipmentpb.Shipment{
		Id:              dto.ID,
		ReferenceNumber: dto.ReferenceNum,
		Origin:          dto.Origin,
		Destination:     dto.Destination,
		DriverName:      dto.DriverName,
		VehicleId:       dto.VehicleID,
		Amount:          dto.Amount,
		DriverRevenue:   dto.DriverRevenue,
		CurrentStatus:   shipmentpb.Status(dto.CurrentStatus),
		CreatedAt:       timestamppb.New(dto.CreatedAt),
		UpdatedAt:       timestamppb.New(dto.UpdatedAt),
	}
}

func (h *ShipmentHandler) shipmentEventDTOToProto(dto *application.ShipmentEventDTO) *shipmentpb.ShipmentEvent {
	if dto == nil {
		return nil
	}

	return &shipmentpb.ShipmentEvent{
		Id:         dto.ID,
		ShipmentId: dto.ShipmentID,
		Status:     shipmentpb.Status(dto.Status),
		Note:       dto.Note,
		OccurredAt: timestamppb.New(dto.OccurredAt),
	}
}

func (h *ShipmentHandler) mapDomainErrorToGRPC(err error) error {
	switch err {
	case domain.ErrShipmentNotFound:
		return status.Error(codes.NotFound, "shipment not found")
	case domain.ErrInvalidTransition:
		return status.Error(codes.InvalidArgument, "invalid status transition")
	case domain.ErrShipmentTerminal:
		return status.Error(codes.FailedPrecondition, "shipment has reached terminal status")
	case domain.ErrDuplicateStatus:
		return status.Error(codes.InvalidArgument, "cannot add event with same status as current")
	case domain.ErrReferenceNumberExists:
		return status.Error(codes.AlreadyExists, "reference number already exists")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
