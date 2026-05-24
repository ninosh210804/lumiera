package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

// SaleService manages sale events that override product prices while active.
type SaleService struct {
	q *pgdb.Queries
}

func NewSaleService(q *pgdb.Queries) *SaleService {
	return &SaleService{q: q}
}

type SaleEventDTO struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	IsActive  bool               `json:"is_active"`
	ItemCount int64              `json:"item_count"`
	Items     []SaleEventItemDTO `json:"items,omitempty"`
}

type SaleEventItemDTO struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	BasePrice   float64   `json:"base_price"`
	SalePrice   float64   `json:"sale_price"`
}

func (s *SaleService) List(ctx context.Context, locationID uuid.UUID) ([]SaleEventDTO, error) {
	rows, err := s.q.ListSaleEvents(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list sale events: %w", err)
	}
	out := make([]SaleEventDTO, len(rows))
	for i, r := range rows {
		out[i] = SaleEventDTO{
			ID:        uuid.UUID(r.ID.Bytes),
			Name:      r.Name,
			IsActive:  r.IsActive,
			ItemCount: r.ItemCount,
		}
	}
	return out, nil
}

func (s *SaleService) Get(ctx context.Context, id uuid.UUID) (*SaleEventDTO, error) {
	ev, err := s.q.GetSaleEvent(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, domain.NewNotFound("sale event")
	}
	items, err := s.q.ListSaleEventItems(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list sale items: %w", err)
	}
	dto := &SaleEventDTO{
		ID:       uuid.UUID(ev.ID.Bytes),
		Name:     ev.Name,
		IsActive: ev.IsActive,
		Items:    make([]SaleEventItemDTO, len(items)),
	}
	for i, it := range items {
		dto.Items[i] = SaleEventItemDTO{
			ProductID:   uuid.UUID(it.ProductID.Bytes),
			ProductName: it.ProductName,
			BasePrice:   floatFromNumeric(it.BasePrice),
			SalePrice:   floatFromNumeric(it.SalePrice),
		}
	}
	dto.ItemCount = int64(len(items))
	return dto, nil
}

func (s *SaleService) Create(ctx context.Context, locationID, createdBy uuid.UUID, name string, active bool) (*SaleEventDTO, error) {
	if name == "" {
		return nil, domain.NewBadRequest("name is required")
	}
	ev, err := s.q.CreateSaleEvent(ctx, pgdb.CreateSaleEventParams{
		LocationID: pgtype.UUID{Bytes: locationID, Valid: true},
		Name:       name,
		IsActive:   active,
		CreatedBy:  pgtype.UUID{Bytes: createdBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create sale event: %w", err)
	}
	return &SaleEventDTO{ID: uuid.UUID(ev.ID.Bytes), Name: ev.Name, IsActive: ev.IsActive}, nil
}

func (s *SaleService) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	if _, err := s.q.SetSaleEventActive(ctx, pgdb.SetSaleEventActiveParams{
		ID:       pgtype.UUID{Bytes: id, Valid: true},
		IsActive: active,
	}); err != nil {
		return fmt.Errorf("set sale active: %w", err)
	}
	return nil
}

func (s *SaleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.q.DeleteSaleEvent(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

func (s *SaleService) AddItem(ctx context.Context, eventID, productID uuid.UUID, salePrice float64) error {
	if salePrice < 0 {
		return domain.NewBadRequest("sale_price must be >= 0")
	}
	if _, err := s.q.AddSaleEventItem(ctx, pgdb.AddSaleEventItemParams{
		SaleEventID: pgtype.UUID{Bytes: eventID, Valid: true},
		ProductID:   pgtype.UUID{Bytes: productID, Valid: true},
		SalePrice:   numericFromFloat(salePrice),
	}); err != nil {
		return fmt.Errorf("add sale item: %w", err)
	}
	return nil
}

func (s *SaleService) RemoveItem(ctx context.Context, eventID, productID uuid.UUID) error {
	return s.q.RemoveSaleEventItem(ctx, pgdb.RemoveSaleEventItemParams{
		SaleEventID: pgtype.UUID{Bytes: eventID, Valid: true},
		ProductID:   pgtype.UUID{Bytes: productID, Valid: true},
	})
}
