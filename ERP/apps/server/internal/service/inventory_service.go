package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

// IngredientService handles ingredient CRUD and stock reads.
type IngredientService struct {
	q *pgdb.Queries
}

func NewIngredientService(q *pgdb.Queries) *IngredientService {
	return &IngredientService{q: q}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type IngredientDTO struct {
	ID                   uuid.UUID  `json:"id"`
	LocationID           uuid.UUID  `json:"location_id"`
	Name                 string     `json:"name"`
	Unit                 string     `json:"unit"`
	IsPerishable         bool       `json:"is_perishable"`
	DefaultShelfLifeDays *int32     `json:"default_shelf_life_days"`
	MinStockAlert        float64    `json:"min_stock_alert"`
	CostMethod           string     `json:"cost_method"`
	CurrentAvgCost       float64    `json:"current_avg_cost"`
	CurrentQty           float64    `json:"current_qty"`
	IsActive             bool       `json:"is_active"`
}

// ─── Queries ──────────────────────────────────────────────────────────────────

func (s *IngredientService) List(ctx context.Context, locationID uuid.UUID) ([]IngredientDTO, error) {
	rows, err := s.q.ListIngredients(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list ingredients: %w", err)
	}
	out := make([]IngredientDTO, len(rows))
	for i, r := range rows {
		out[i] = ingredientToDTO(r)
	}
	return out, nil
}

func (s *IngredientService) ListLowStock(ctx context.Context, locationID uuid.UUID) ([]IngredientDTO, error) {
	rows, err := s.q.ListLowStockIngredients(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list low-stock: %w", err)
	}
	out := make([]IngredientDTO, len(rows))
	for i, r := range rows {
		out[i] = ingredientToDTO(r)
	}
	return out, nil
}

func (s *IngredientService) Get(ctx context.Context, id uuid.UUID) (*IngredientDTO, error) {
	row, err := s.q.GetIngredient(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, domain.NewNotFound("ingredient")
	}
	dto := ingredientToDTO(row)
	return &dto, nil
}

// ─── Mutations ────────────────────────────────────────────────────────────────

type CreateIngredientInput struct {
	LocationID           uuid.UUID
	Name                 string
	Unit                 string
	IsPerishable         bool
	DefaultShelfLifeDays *int32
	MinStockAlert        float64
	CreatedBy            uuid.UUID
}

func (s *IngredientService) Create(ctx context.Context, in CreateIngredientInput) (*IngredientDTO, error) {
	row, err := s.q.CreateIngredient(ctx, pgdb.CreateIngredientParams{
		LocationID:           pgtype.UUID{Bytes: in.LocationID, Valid: true},
		Name:                 in.Name,
		Unit:                 in.Unit,
		IsPerishable:         in.IsPerishable,
		DefaultShelfLifeDays: in.DefaultShelfLifeDays,
		MinStockAlert:        numericFromFloat(in.MinStockAlert),
		CreatedBy:            pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create ingredient: %w", err)
	}
	dto := ingredientToDTO(row)
	return &dto, nil
}

type UpdateIngredientInput struct {
	ID                   uuid.UUID
	Name                 string
	Unit                 string
	IsPerishable         bool
	DefaultShelfLifeDays *int32
	MinStockAlert        float64
	IsActive             bool
	// Optional manual overrides used by the warehouse edit dialog (/5).
	// When non-nil they replace current_qty / current_avg_cost — qty via an
	// "adjustment" stock_movement so the trigger keeps batches consistent,
	// cost via a direct write (the trigger only weighted-averages on
	// positive movements, so there's no other way to correct it).
	CurrentQty     *float64
	CurrentAvgCost *float64
	CreatedBy      uuid.UUID
}

func (s *IngredientService) Update(ctx context.Context, in UpdateIngredientInput) (*IngredientDTO, error) {
	row, err := s.q.UpdateIngredient(ctx, pgdb.UpdateIngredientParams{
		ID:                   pgtype.UUID{Bytes: in.ID, Valid: true},
		Name:                 in.Name,
		Unit:                 in.Unit,
		IsPerishable:         in.IsPerishable,
		DefaultShelfLifeDays: in.DefaultShelfLifeDays,
		MinStockAlert:        numericFromFloat(in.MinStockAlert),
		IsActive:             in.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("update ingredient: %w", err)
	}

	// Stock qty override → log an adjustment movement so the trigger updates
	// current_qty and we keep an audit trail. Only fire when the requested
	// qty actually differs from what's there.
	if in.CurrentQty != nil {
		currentQty := floatFromNumeric(row.CurrentQty)
		delta := *in.CurrentQty - currentQty
		if delta != 0 {
			note := "manual stock override"
			_, err := s.q.CreateStockMovement(ctx, pgdb.CreateStockMovementParams{
				LocationID:       row.LocationID,
				IngredientID:     pgtype.UUID{Bytes: in.ID, Valid: true},
				BatchID:          pgtype.UUID{},
				QtyDelta:         numericFromFloat(delta),
				UnitCostSnapshot: row.CurrentAvgCost,
				Reason:           "adjustment",
				OrderID:          pgtype.UUID{},
				InventoryCountID: pgtype.UUID{},
				Note:             &note,
				ClientUuid:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
				CreatedBy:        pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
			})
			if err != nil {
				return nil, fmt.Errorf("adjust stock: %w", err)
			}
		}
	}

	// Cost override → direct write (cost-update trigger only kicks in on
	// positive movements, so we cannot piggy-back on the qty adjustment).
	if in.CurrentAvgCost != nil {
		row2, err := s.q.SetIngredientCost(ctx, pgdb.SetIngredientCostParams{
			ID:             pgtype.UUID{Bytes: in.ID, Valid: true},
			CurrentAvgCost: numericFromFloat(*in.CurrentAvgCost),
		})
		if err != nil {
			return nil, fmt.Errorf("set cost: %w", err)
		}
		row = row2
	}

	// Re-fetch in case the trigger updated qty above.
	if in.CurrentQty != nil {
		latest, err := s.q.GetIngredient(ctx, pgtype.UUID{Bytes: in.ID, Valid: true})
		if err == nil {
			row = latest
		}
	}

	dto := ingredientToDTO(row)
	return &dto, nil
}

func (s *IngredientService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.q.SoftDeleteIngredient(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func ingredientToDTO(r pgdb.Ingredient) IngredientDTO {
	return IngredientDTO{
		ID:                   uuid.UUID(r.ID.Bytes),
		LocationID:           uuid.UUID(r.LocationID.Bytes),
		Name:                 r.Name,
		Unit:                 r.Unit,
		IsPerishable:         r.IsPerishable,
		DefaultShelfLifeDays: r.DefaultShelfLifeDays,
		MinStockAlert:        floatFromNumeric(r.MinStockAlert),
		CostMethod:           r.CostMethod,
		CurrentAvgCost:       floatFromNumeric(r.CurrentAvgCost),
		CurrentQty:           floatFromNumeric(r.CurrentQty),
		IsActive:             r.IsActive,
	}
}
