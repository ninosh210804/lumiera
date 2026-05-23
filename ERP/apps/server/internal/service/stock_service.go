package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

type StockService struct {
	q    *pgdb.Queries
	pool *pgxpool.Pool
}

func NewStockService(pool *pgxpool.Pool, q *pgdb.Queries) *StockService {
	return &StockService{q: q, pool: pool}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type StockMovementDTO struct {
	ID             uuid.UUID  `json:"id"`
	LocationID     uuid.UUID  `json:"location_id"`
	IngredientID   uuid.UUID  `json:"ingredient_id"`
	IngredientName string     `json:"ingredient_name"`
	QtyDelta       float64    `json:"qty_delta"`
	UnitCost       float64    `json:"unit_cost"`
	Reason         string     `json:"reason"`
	Note           string     `json:"note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type SupplierDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Contact string    `json:"contact"`
	Phone   string    `json:"phone,omitempty"`
	BIN     string    `json:"bin,omitempty"`
	Address string    `json:"address,omitempty"`
	IsActive bool     `json:"is_active"`
}

type PurchaseReceiptDTO struct {
	PurchaseOrderID uuid.UUID              `json:"purchase_order_id"`
	TotalAmount     float64                `json:"total_amount"`
	Items           []PurchaseReceiptItem  `json:"items"`
}

type PurchaseReceiptItem struct {
	IngredientID   uuid.UUID `json:"ingredient_id"`
	IngredientName string    `json:"ingredient_name"`
	Qty            float64   `json:"qty"`
	UnitCost       float64   `json:"unit_cost"`
	LineCost       float64   `json:"line_cost"`
}

type InventoryCountDTO struct {
	ID          uuid.UUID              `json:"id"`
	LocationID  uuid.UUID              `json:"location_id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Items       []InventoryCountItemDTO `json:"items,omitempty"`
}

type InventoryCountItemDTO struct {
	IngredientID   uuid.UUID `json:"ingredient_id"`
	IngredientName string    `json:"ingredient_name"`
	Unit           string    `json:"unit"`
	ExpectedQty    float64   `json:"expected_qty"`
	ActualQty      float64   `json:"actual_qty"`
	Variance       float64   `json:"variance"`
}

// ─── Inputs ───────────────────────────────────────────────────────────────────

type WriteOffInput struct {
	LocationID   uuid.UUID
	IngredientID uuid.UUID
	Qty          float64
	Reason       string
	Note         string
	CreatedBy    uuid.UUID
}

type ReceiveStockInput struct {
	LocationID uuid.UUID
	SupplierID *uuid.UUID
	Notes      string
	CreatedBy  uuid.UUID
	Items      []ReceiveStockItemInput
}

type ReceiveStockItemInput struct {
	IngredientID uuid.UUID
	Qty          float64
	UnitCost     float64
	ExpiresAt    *time.Time
}

type CreateCountInput struct {
	LocationID  uuid.UUID
	PerformedBy uuid.UUID
	Items       []CountItemInput
}

type CountItemInput struct {
	IngredientID uuid.UUID
	ActualQty    float64
}

// ─── WriteOff ─────────────────────────────────────────────────────────────────

var validWriteOffReasons = map[string]bool{
	"waste": true, "spill": true, "gift": true, "staff": true, "adjustment": true,
}

func (svc *StockService) WriteOff(ctx context.Context, in WriteOffInput) (*StockMovementDTO, error) {
	if in.Qty <= 0 {
		return nil, domain.NewBadRequest("qty must be > 0")
	}
	if !validWriteOffReasons[in.Reason] {
		return nil, domain.NewBadRequest("reason must be one of: waste, spill, gift, staff, adjustment")
	}

	var note *string
	if in.Note != "" {
		note = &in.Note
	}
	mv, err := svc.q.CreateStockMovement(ctx, pgdb.CreateStockMovementParams{
		LocationID:       pgtype.UUID{Bytes: in.LocationID, Valid: true},
		IngredientID:     pgtype.UUID{Bytes: in.IngredientID, Valid: true},
		BatchID:          pgtype.UUID{},
		QtyDelta:         numericFromFloat(-in.Qty),
		UnitCostSnapshot: numericFromFloat(0),
		Reason:           in.Reason,
		OrderID:          pgtype.UUID{},
		InventoryCountID: pgtype.UUID{},
		Note:             note,
		ClientUuid:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedBy:        pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create stock movement: %w", err)
	}

	// Auto stop-list any affected products if ingredient hits zero
	ingredient, err := svc.q.GetIngredient(ctx, pgtype.UUID{Bytes: in.IngredientID, Valid: true})
	if err == nil && floatFromNumeric(ingredient.CurrentQty) <= 0 {
		products, _ := svc.q.GetProductsByIngredient(ctx, pgtype.UUID{Bytes: in.IngredientID, Valid: true})
		for _, p := range products {
			_ = svc.q.SetProductStopList(ctx, pgdb.SetProductStopListParams{
				ID:           p.ID,
				IsStopListed: true,
			})
		}
	}

	return &StockMovementDTO{
		ID:           uuid.UUID(mv.ID.Bytes),
		LocationID:   uuid.UUID(mv.LocationID.Bytes),
		IngredientID: uuid.UUID(mv.IngredientID.Bytes),
		QtyDelta:     floatFromNumeric(mv.QtyDelta),
		UnitCost:     floatFromNumeric(mv.UnitCostSnapshot),
		Reason:       mv.Reason,
		Note:         in.Note,
		CreatedAt:    mv.CreatedAt.Time,
	}, nil
}

// ─── ReceiveStock ─────────────────────────────────────────────────────────────

func (svc *StockService) ReceiveStock(ctx context.Context, in ReceiveStockInput) (*PurchaseReceiptDTO, error) {
	if len(in.Items) == 0 {
		return nil, domain.NewBadRequest("items is required")
	}

	tx, err := svc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := svc.q.WithTx(tx)

	supplierID := pgtype.UUID{}
	if in.SupplierID != nil {
		supplierID = pgtype.UUID{Bytes: *in.SupplierID, Valid: true}
	}

	po, err := q.CreatePurchaseOrder(ctx, pgdb.CreatePurchaseOrderParams{
		LocationID: pgtype.UUID{Bytes: in.LocationID, Valid: true},
		SupplierID: supplierID,
		Notes:      in.Notes,
		ClientUuid: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedBy:  pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create purchase order: %w", err)
	}
	poID := pgtype.UUID{Bytes: uuid.UUID(po.ID.Bytes), Valid: true}

	var totalAmount float64
	var dtoItems []PurchaseReceiptItem

	for _, item := range in.Items {
		expiresAt := pgtype.Date{}
		if item.ExpiresAt != nil {
			expiresAt = pgtype.Date{Time: *item.ExpiresAt, Valid: true}
		}

		poi, err := q.AddPurchaseOrderItem(ctx, pgdb.AddPurchaseOrderItemParams{
			PurchaseOrderID: poID,
			IngredientID:    pgtype.UUID{Bytes: item.IngredientID, Valid: true},
			Qty:             numericFromFloat(item.Qty),
			UnitCost:        numericFromFloat(item.UnitCost),
			ExpiresAt:       expiresAt,
		})
		if err != nil {
			return nil, fmt.Errorf("add purchase order item: %w", err)
		}
		poiID := pgtype.UUID{Bytes: uuid.UUID(poi.ID.Bytes), Valid: true}

		batch, err := q.CreateStockBatch(ctx, pgdb.CreateStockBatchParams{
			IngredientID:        pgtype.UUID{Bytes: item.IngredientID, Valid: true},
			PurchaseOrderItemID: poiID,
			QtyReceived:         numericFromFloat(item.Qty),
			UnitCost:            numericFromFloat(item.UnitCost),
			ExpiresAt:           expiresAt,
		})
		if err != nil {
			return nil, fmt.Errorf("create stock batch: %w", err)
		}

		_, err = q.CreateStockMovement(ctx, pgdb.CreateStockMovementParams{
			LocationID:       pgtype.UUID{Bytes: in.LocationID, Valid: true},
			IngredientID:     pgtype.UUID{Bytes: item.IngredientID, Valid: true},
			BatchID:          pgtype.UUID{Bytes: uuid.UUID(batch.ID.Bytes), Valid: true},
			QtyDelta:         numericFromFloat(item.Qty),
			UnitCostSnapshot: numericFromFloat(item.UnitCost),
			Reason:           "purchase",
			OrderID:          pgtype.UUID{},
			InventoryCountID: pgtype.UUID{},
			Note:             nil,
			ClientUuid:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			CreatedBy:        pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("create stock movement: %w", err)
		}

		lineCost := item.Qty * item.UnitCost
		totalAmount += lineCost

		ing, _ := q.GetIngredient(ctx, pgtype.UUID{Bytes: item.IngredientID, Valid: true})
		ingName := ""
		if ing.ID.Valid {
			ingName = ing.Name
		}
		dtoItems = append(dtoItems, PurchaseReceiptItem{
			IngredientID:   item.IngredientID,
			IngredientName: ingName,
			Qty:            item.Qty,
			UnitCost:       item.UnitCost,
			LineCost:       lineCost,
		})
	}

	if _, err := q.ReceivePurchaseOrder(ctx, pgdb.ReceivePurchaseOrderParams{
		ID:          poID,
		TotalAmount: numericFromFloat(totalAmount),
	}); err != nil {
		return nil, fmt.Errorf("receive purchase order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &PurchaseReceiptDTO{
		PurchaseOrderID: uuid.UUID(po.ID.Bytes),
		TotalAmount:     totalAmount,
		Items:           dtoItems,
	}, nil
}

// ─── CreateCount ──────────────────────────────────────────────────────────────

func (svc *StockService) CreateCount(ctx context.Context, in CreateCountInput) (*InventoryCountDTO, error) {
	if len(in.Items) == 0 {
		return nil, domain.NewBadRequest("items is required")
	}

	tx, err := svc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := svc.q.WithTx(tx)

	count, err := q.CreateInventoryCount(ctx, pgdb.CreateInventoryCountParams{
		LocationID:  pgtype.UUID{Bytes: in.LocationID, Valid: true},
		PerformedBy: pgtype.UUID{Bytes: in.PerformedBy, Valid: true},
		ClientUuid:  pgtype.UUID{Bytes: uuid.New(), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create inventory count: %w", err)
	}
	countID := pgtype.UUID{Bytes: uuid.UUID(count.ID.Bytes), Valid: true}

	var dtoItems []InventoryCountItemDTO
	for _, item := range in.Items {
		ing, err := q.GetIngredient(ctx, pgtype.UUID{Bytes: item.IngredientID, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.NewNotFound("ingredient")
			}
			return nil, fmt.Errorf("get ingredient: %w", err)
		}

		expectedQty := floatFromNumeric(ing.CurrentQty)
		variance := item.ActualQty - expectedQty

		_, err = q.AddInventoryCountItem(ctx, pgdb.AddInventoryCountItemParams{
			InventoryCountID: countID,
			IngredientID:     pgtype.UUID{Bytes: item.IngredientID, Valid: true},
			ExpectedQty:      numericFromFloat(expectedQty),
			ActualQty:        numericFromFloat(item.ActualQty),
		})
		if err != nil {
			return nil, fmt.Errorf("add count item: %w", err)
		}

		if variance != 0 {
			note := "inventory count adjustment"
			_, err = q.CreateStockMovement(ctx, pgdb.CreateStockMovementParams{
				LocationID:       pgtype.UUID{Bytes: in.LocationID, Valid: true},
				IngredientID:     pgtype.UUID{Bytes: item.IngredientID, Valid: true},
				BatchID:          pgtype.UUID{},
				QtyDelta:         numericFromFloat(variance),
				UnitCostSnapshot: ing.CurrentAvgCost,
				Reason:           "count",
				OrderID:          pgtype.UUID{},
				InventoryCountID: countID,
				Note:             &note,
				ClientUuid:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
				CreatedBy:        pgtype.UUID{Bytes: in.PerformedBy, Valid: true},
			})
			if err != nil {
				return nil, fmt.Errorf("create adjustment movement: %w", err)
			}
		}

		// Auto stop-list if actual_qty hits zero
		if item.ActualQty <= 0 {
			products, _ := q.GetProductsByIngredient(ctx, pgtype.UUID{Bytes: item.IngredientID, Valid: true})
			for _, p := range products {
				_ = q.SetProductStopList(ctx, pgdb.SetProductStopListParams{
					ID:           p.ID,
					IsStopListed: true,
				})
			}
		}

		dtoItems = append(dtoItems, InventoryCountItemDTO{
			IngredientID:   item.IngredientID,
			IngredientName: ing.Name,
			Unit:           ing.Unit,
			ExpectedQty:    expectedQty,
			ActualQty:      item.ActualQty,
			Variance:       variance,
		})
	}

	completed, err := q.CompleteInventoryCount(ctx, countID)
	if err != nil {
		return nil, fmt.Errorf("complete count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	dto := inventoryCountToDTO(completed)
	dto.Items = dtoItems
	return &dto, nil
}

// ─── Movements ────────────────────────────────────────────────────────────────

func (svc *StockService) ListMovements(ctx context.Context, locationID uuid.UUID, ingredientID *uuid.UUID, from, to time.Time) ([]StockMovementDTO, error) {
	ingFilter := pgtype.UUID{}
	if ingredientID != nil {
		ingFilter = pgtype.UUID{Bytes: *ingredientID, Valid: true}
	}
	rows, err := svc.q.GetStockMovements(ctx, pgdb.GetStockMovementsParams{
		LocationID:  pgtype.UUID{Bytes: locationID, Valid: true},
		Column2:     ingFilter,
		CreatedAt:   pgtype.Timestamptz{Time: from, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: to, Valid: true},
		Limit:       200,
	})
	if err != nil {
		return nil, fmt.Errorf("list stock movements: %w", err)
	}
	out := make([]StockMovementDTO, len(rows))
	for i, r := range rows {
		note := ""
		if r.Note != nil {
			note = *r.Note
		}
		out[i] = StockMovementDTO{
			ID:             uuid.UUID(r.ID.Bytes),
			LocationID:     uuid.UUID(r.LocationID.Bytes),
			IngredientID:   uuid.UUID(r.IngredientID.Bytes),
			IngredientName: r.IngredientName,
			QtyDelta:       floatFromNumeric(r.QtyDelta),
			UnitCost:       floatFromNumeric(r.UnitCostSnapshot),
			Reason:         r.Reason,
			Note:           note,
			CreatedAt:      r.CreatedAt.Time,
		}
	}
	return out, nil
}

// ─── Suppliers ────────────────────────────────────────────────────────────────

func (svc *StockService) ListSuppliers(ctx context.Context) ([]SupplierDTO, error) {
	rows, err := svc.q.ListSuppliers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	out := make([]SupplierDTO, len(rows))
	for i, r := range rows {
		out[i] = supplierToDTO(r)
	}
	return out, nil
}

type CreateSupplierInput struct {
	Name    string
	Contact string
	Phone   string
	BIN     string
	Address string
}

func (svc *StockService) CreateSupplier(ctx context.Context, in CreateSupplierInput) (*SupplierDTO, error) {
	if in.Name == "" {
		return nil, domain.NewBadRequest("name is required")
	}
	s, err := svc.q.CreateSupplier(ctx, pgdb.CreateSupplierParams{
		Name:    in.Name,
		Contact: in.Contact,
		Phone:   nullableStr(in.Phone),
		Bin:     nullableStr(in.BIN),
		Address: nullableStr(in.Address),
	})
	if err != nil {
		return nil, fmt.Errorf("create supplier: %w", err)
	}
	dto := supplierToDTO(s)
	return &dto, nil
}

// ─── List counts ──────────────────────────────────────────────────────────────

func (svc *StockService) ListCounts(ctx context.Context, locationID uuid.UUID) ([]InventoryCountDTO, error) {
	rows, err := svc.q.ListInventoryCounts(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list counts: %w", err)
	}
	out := make([]InventoryCountDTO, len(rows))
	for i, r := range rows {
		out[i] = inventoryCountToDTO(r)
	}
	return out, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func supplierToDTO(s pgdb.Supplier) SupplierDTO {
	return SupplierDTO{
		ID:       uuid.UUID(s.ID.Bytes),
		Name:     s.Name,
		Contact:  s.Contact,
		Phone:    derefStr(s.Phone),
		BIN:      derefStr(s.Bin),
		Address:  derefStr(s.Address),
		IsActive: s.IsActive,
	}
}

func inventoryCountToDTO(c pgdb.InventoryCount) InventoryCountDTO {
	dto := InventoryCountDTO{
		ID:         uuid.UUID(c.ID.Bytes),
		LocationID: uuid.UUID(c.LocationID.Bytes),
		Status:     c.Status,
		CreatedAt:  c.CreatedAt.Time,
	}
	if c.CompletedAt.Valid {
		t := c.CompletedAt.Time
		dto.CompletedAt = &t
	}
	return dto
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

