package repository

import (
	"context"
	"sync"

	"shipment-service/internal/domain"
)

type MemoryRepository struct {
	mu         sync.RWMutex
	shipments  map[string]*domain.Shipment
	refNumbers map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		shipments:  make(map[string]*domain.Shipment),
		refNumbers: make(map[string]string),
	}
}

func (r *MemoryRepository) Save(ctx context.Context, shipment *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.shipments[shipment.ID]; ok {
		if existing.ReferenceNum != shipment.ReferenceNum {
			delete(r.refNumbers, existing.ReferenceNum)
			r.refNumbers[shipment.ReferenceNum] = shipment.ID
		}
	} else {
		r.refNumbers[shipment.ReferenceNum] = shipment.ID
	}

	r.shipments[shipment.ID] = shipment
	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shipment, ok := r.shipments[id]
	if !ok {
		return nil, domain.ErrShipmentNotFound
	}

	return shipment, nil
}

func (r *MemoryRepository) GetByReferenceNumber(ctx context.Context, referenceNumber string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.refNumbers[referenceNumber]
	if !ok {
		return nil, domain.ErrShipmentNotFound
	}

	return r.shipments[id], nil
}

func (r *MemoryRepository) List(ctx context.Context) ([]*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shipments := make([]*domain.Shipment, 0, len(r.shipments))
	for _, shipment := range r.shipments {
		shipments = append(shipments, shipment)
	}

	return shipments, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	shipment, ok := r.shipments[id]
	if !ok {
		return domain.ErrShipmentNotFound
	}

	delete(r.refNumbers, shipment.ReferenceNum)
	delete(r.shipments, id)
	return nil
}
