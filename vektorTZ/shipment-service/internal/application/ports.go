package application

import (
	"context"

	"shipment-service/internal/domain"
)

type ShipmentRepository interface {
	Save(ctx context.Context, shipment *domain.Shipment) error
	GetByID(ctx context.Context, id string) (*domain.Shipment, error)
	GetByReferenceNumber(ctx context.Context, referenceNumber string) (*domain.Shipment, error)
	List(ctx context.Context) ([]*domain.Shipment, error)
	Delete(ctx context.Context, id string) error
}
