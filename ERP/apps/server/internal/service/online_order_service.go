package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

// OnlineOrderService handles customer-placed delivery orders. These live in the
// online_orders table with their own fulfilment lifecycle
// (new → preparing → ready → completed). On completion the order is funnelled
// through OrderService.CreateOrder so revenue, stock, and loyalty flow through
// the proven POS path.
type OnlineOrderService struct {
	q      *pgdb.Queries
	pool   *pgxpool.Pool
	orders *OrderService
}

func NewOnlineOrderService(pool *pgxpool.Pool, q *pgdb.Queries, orders *OrderService) *OnlineOrderService {
	return &OnlineOrderService{q: q, pool: pool, orders: orders}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type OnlineOrderDTO struct {
	ID             uuid.UUID            `json:"id"`
	LocationID     uuid.UUID            `json:"location_id"`
	CustomerPhone  string               `json:"customer_phone"`
	Status         string               `json:"status"`
	DeliveryOffice string               `json:"delivery_office"`
	DeliveryNote   string               `json:"delivery_note"`
	Subtotal       float64              `json:"subtotal"`
	Total          float64              `json:"total"`
	OrderID        *uuid.UUID           `json:"order_id,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	Items          []OnlineOrderItemDTO `json:"items,omitempty"`
}

type OnlineOrderItemDTO struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Qty         float64   `json:"qty"`
	UnitPrice   float64   `json:"unit_price"`
	LineTotal   float64   `json:"line_total"`
}

// ─── Inputs ───────────────────────────────────────────────────────────────────

type PlaceOnlineOrderInput struct {
	LocationID     uuid.UUID
	CustomerPhone  string
	DeliveryOffice string
	DeliveryNote   string
	Items          []OrderItemInput
}

type CompleteOnlineOrderInput struct {
	OnlineOrderID uuid.UUID
	BaristaID     uuid.UUID
	LocationID    uuid.UUID
	ShiftID       *uuid.UUID
	PaymentMethod string
}

// Legal fulfilment transitions for staff actions.
var onlineOrderTransitions = map[string]map[string]bool{
	"new":       {"preparing": true, "rejected": true},
	"preparing": {"ready": true, "rejected": true},
	"ready":     {"rejected": true},
}

// ─── PlaceOrder (public) ────────────────────────────────────────────────────

func (svc *OnlineOrderService) PlaceOrder(ctx context.Context, in PlaceOnlineOrderInput) (*OnlineOrderDTO, error) {
	if normalizePhone(in.CustomerPhone) == "" {
		return nil, domain.NewBadRequest("phone is required")
	}
	if len(in.Items) == 0 {
		return nil, domain.NewBadRequest("order must have at least one item")
	}
	if in.LocationID == (uuid.UUID{}) {
		return nil, domain.NewBadRequest("location is required")
	}

	tx, err := svc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := svc.q.WithTx(tx)

	type calcItem struct {
		productID   uuid.UUID
		productName string
		qty         float64
		unitPrice   float64
		lineTotal   float64
	}
	var calc []calcItem
	var subtotal float64
	for _, item := range in.Items {
		if item.Qty <= 0 {
			return nil, domain.NewBadRequest("item qty must be > 0")
		}
		prod, err := q.GetProduct(ctx, pgtype.UUID{Bytes: item.ProductID, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.NewNotFound("product")
			}
			return nil, fmt.Errorf("get product: %w", err)
		}
		if !prod.IsActive || prod.IsStopListed {
			return nil, domain.NewBadRequest(fmt.Sprintf("product '%s' is not available", prod.Name))
		}
		unitPrice := floatFromNumeric(prod.BasePrice)
		if prod.SalePrice.Valid {
			unitPrice = floatFromNumeric(prod.SalePrice)
		}
		lineTotal := math.Round(unitPrice*item.Qty*100) / 100
		subtotal += lineTotal
		calc = append(calc, calcItem{
			productID:   item.ProductID,
			productName: prod.Name,
			qty:         item.Qty,
			unitPrice:   unitPrice,
			lineTotal:   lineTotal,
		})
	}

	// Ensure the customer + loyalty account exist (phone signs them in for bonuses).
	custID, _, _, err := getOrCreateCustomer(ctx, q, in.CustomerPhone)
	if err != nil {
		return nil, err
	}

	oo, err := q.CreateOnlineOrder(ctx, pgdb.CreateOnlineOrderParams{
		LocationID:     pgtype.UUID{Bytes: in.LocationID, Valid: true},
		CustomerID:     pgtype.UUID{Bytes: custID, Valid: true},
		CustomerPhone:  normalizePhone(in.CustomerPhone),
		DeliveryOffice: in.DeliveryOffice,
		DeliveryNote:   in.DeliveryNote,
		Subtotal:       numericFromFloat(subtotal),
		Total:          numericFromFloat(subtotal),
	})
	if err != nil {
		return nil, fmt.Errorf("create online order: %w", err)
	}

	items := make([]OnlineOrderItemDTO, 0, len(calc))
	for _, ci := range calc {
		if _, err := q.CreateOnlineOrderItem(ctx, pgdb.CreateOnlineOrderItemParams{
			OnlineOrderID:     oo.ID,
			ProductID:         pgtype.UUID{Bytes: ci.productID, Valid: true},
			Qty:               numericFromFloat(ci.qty),
			UnitPriceSnapshot: numericFromFloat(ci.unitPrice),
		}); err != nil {
			return nil, fmt.Errorf("create online order item: %w", err)
		}
		items = append(items, OnlineOrderItemDTO{
			ProductID:   ci.productID,
			ProductName: ci.productName,
			Qty:         ci.qty,
			UnitPrice:   ci.unitPrice,
			LineTotal:   ci.lineTotal,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	dto := onlineOrderToDTO(oo)
	dto.Items = items
	return &dto, nil
}

// ─── Track (public) ─────────────────────────────────────────────────────────

func (svc *OnlineOrderService) Track(ctx context.Context, id uuid.UUID) (*OnlineOrderDTO, error) {
	oo, err := svc.q.GetOnlineOrder(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("online order")
		}
		return nil, fmt.Errorf("get online order: %w", err)
	}
	dto := onlineOrderToDTO(oo)
	items, err := svc.loadItems(ctx, oo.ID)
	if err != nil {
		return nil, err
	}
	dto.Items = items
	return &dto, nil
}

// ─── ListActive (staff) ─────────────────────────────────────────────────────

func (svc *OnlineOrderService) ListActive(ctx context.Context, locationID uuid.UUID) ([]OnlineOrderDTO, error) {
	rows, err := svc.q.ListActiveOnlineOrders(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list active online orders: %w", err)
	}
	out := make([]OnlineOrderDTO, 0, len(rows))
	for _, oo := range rows {
		dto := onlineOrderToDTO(oo)
		items, err := svc.loadItems(ctx, oo.ID)
		if err != nil {
			return nil, err
		}
		dto.Items = items
		out = append(out, dto)
	}
	return out, nil
}

// ─── Advance (staff) ────────────────────────────────────────────────────────

func (svc *OnlineOrderService) Advance(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) (*OnlineOrderDTO, error) {
	oo, err := svc.q.GetOnlineOrder(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("online order")
		}
		return nil, fmt.Errorf("get online order: %w", err)
	}
	if !onlineOrderTransitions[oo.Status][status] {
		return nil, domain.NewBadRequest(fmt.Sprintf("cannot change status from '%s' to '%s'", oo.Status, status))
	}
	updated, err := svc.q.SetOnlineOrderStatus(ctx, pgdb.SetOnlineOrderStatusParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		Status:  status,
		Column3: pgtype.UUID{Bytes: userID, Valid: userID != (uuid.UUID{})},
	})
	if err != nil {
		return nil, fmt.Errorf("set online order status: %w", err)
	}
	dto := onlineOrderToDTO(updated)
	return &dto, nil
}

// ─── Complete (staff) ───────────────────────────────────────────────────────

// Complete collects payment on handover: it books a real POS order (revenue +
// stock deduction + loyalty earn/redeem) via OrderService, then marks the online
// order completed and links it to that order.
func (svc *OnlineOrderService) Complete(ctx context.Context, in CompleteOnlineOrderInput) (*OnlineOrderDTO, error) {
	oo, err := svc.q.GetOnlineOrder(ctx, pgtype.UUID{Bytes: in.OnlineOrderID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("online order")
		}
		return nil, fmt.Errorf("get online order: %w", err)
	}
	if oo.Status == "completed" {
		return nil, domain.NewBadRequest("online order is already completed")
	}
	if oo.Status == "rejected" {
		return nil, domain.NewBadRequest("online order was rejected")
	}
	if in.PaymentMethod == "" {
		return nil, domain.NewBadRequest("payment method is required")
	}

	itemRows, err := svc.q.GetOnlineOrderItems(ctx, oo.ID)
	if err != nil {
		return nil, fmt.Errorf("get online order items: %w", err)
	}
	if len(itemRows) == 0 {
		return nil, domain.NewBadRequest("online order has no items")
	}
	phone := oo.CustomerPhone

	orderItems := make([]OrderItemInput, 0, len(itemRows))
	for _, it := range itemRows {
		orderItems = append(orderItems, OrderItemInput{
			ProductID: uuid.UUID(it.ProductID.Bytes),
			Qty:       floatFromNumeric(it.Qty),
		})
	}

	// Price it exactly as CreateOrder will (promo/free-coffee/loyalty applied),
	// so the collected payment matches the order total.
	quote, err := svc.orders.Quote(ctx, QuoteInput{CustomerPhone: phone, Items: orderItems})
	if err != nil {
		return nil, fmt.Errorf("quote online order: %w", err)
	}

	created, err := svc.orders.CreateOrder(ctx, CreateOrderInput{
		LocationID:    in.LocationID,
		ShiftID:       in.ShiftID,
		BaristaID:     in.BaristaID,
		CustomerPhone: phone,
		ClientUUID:    uuid.New(),
		Items:         orderItems,
		Payments: []PaymentInput{{
			MethodCode: in.PaymentMethod,
			Amount:     quote.Total,
		}},
	})
	if err != nil {
		return nil, err
	}

	completed, err := svc.q.CompleteOnlineOrder(ctx, pgdb.CompleteOnlineOrderParams{
		ID:      pgtype.UUID{Bytes: in.OnlineOrderID, Valid: true},
		OrderID: pgtype.UUID{Bytes: created.ID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("complete online order: %w", err)
	}
	dto := onlineOrderToDTO(completed)
	return &dto, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (svc *OnlineOrderService) loadItems(ctx context.Context, onlineOrderID pgtype.UUID) ([]OnlineOrderItemDTO, error) {
	rows, err := svc.q.GetOnlineOrderItems(ctx, onlineOrderID)
	if err != nil {
		return nil, fmt.Errorf("get online order items: %w", err)
	}
	items := make([]OnlineOrderItemDTO, len(rows))
	for i, it := range rows {
		qty := floatFromNumeric(it.Qty)
		unit := floatFromNumeric(it.UnitPriceSnapshot)
		items[i] = OnlineOrderItemDTO{
			ProductID:   uuid.UUID(it.ProductID.Bytes),
			ProductName: it.ProductName,
			Qty:         qty,
			UnitPrice:   unit,
			LineTotal:   math.Round(unit*qty*100) / 100,
		}
	}
	return items, nil
}

func onlineOrderToDTO(oo pgdb.OnlineOrder) OnlineOrderDTO {
	dto := OnlineOrderDTO{
		ID:             uuid.UUID(oo.ID.Bytes),
		LocationID:     uuid.UUID(oo.LocationID.Bytes),
		CustomerPhone:  oo.CustomerPhone,
		Status:         oo.Status,
		DeliveryOffice: oo.DeliveryOffice,
		DeliveryNote:   oo.DeliveryNote,
		Subtotal:       floatFromNumeric(oo.Subtotal),
		Total:          floatFromNumeric(oo.Total),
		CreatedAt:      oo.CreatedAt.Time,
	}
	if oo.OrderID.Valid {
		id := uuid.UUID(oo.OrderID.Bytes)
		dto.OrderID = &id
	}
	return dto
}
