package domain

import (
	"time"

	"github.com/google/uuid"
)

// RoleType defines the user's role
type RoleType string

const (
	RoleAdmin   RoleType = "admin"
	RoleManager RoleType = "manager"
	RoleBarista RoleType = "barista"
)

// CostMethod defines stock valuation method
type CostMethod string

const (
	CostMethodAVG  CostMethod = "AVG"
	CostMethodFIFO CostMethod = "FIFO"
)

// StockReason defines why a stock movement happened
type StockReason string

const (
	StockReasonSale          StockReason = "sale"
	StockReasonWaste         StockReason = "waste"
	StockReasonSpill         StockReason = "spill"
	StockReasonGift          StockReason = "gift"
	StockReasonStaff         StockReason = "staff"
	StockReasonCount         StockReason = "count"
	StockReasonPurchase      StockReason = "purchase"
	StockReasonAdjustment    StockReason = "adjustment"
)

// OrderStatus represents the lifecycle of an order
type OrderStatus string

const (
	OrderStatusOpen      OrderStatus = "open"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// PaymentMethodCode identifies payment types
type PaymentMethodCode string

const (
	PaymentCash     PaymentMethodCode = "cash"
	PaymentKaspiQR  PaymentMethodCode = "kaspi_qr"
	PaymentLoyalty  PaymentMethodCode = "loyalty"
	PaymentGift     PaymentMethodCode = "gift"
)

// SyncEventType identifies what kind of event was synced
type SyncEventType string

const (
	EventOrderCreated   SyncEventType = "ORDER_CREATED"
	EventOrderRefunded  SyncEventType = "ORDER_REFUNDED"
	EventOrderCancelled SyncEventType = "ORDER_CANCELLED"
	EventStockReceived  SyncEventType = "STOCK_RECEIVED"
	EventStockAdjusted  SyncEventType = "STOCK_ADJUSTED"
	EventStockWrittenOff SyncEventType = "STOCK_WRITTEN_OFF"
	EventPriceChanged   SyncEventType = "PRICE_CHANGED"
	EventMenuPublished  SyncEventType = "MENU_PUBLISHED"
	EventShiftOpened    SyncEventType = "SHIFT_OPENED"
	EventShiftClosed    SyncEventType = "SHIFT_CLOSED"
	EventCashDrawerOp   SyncEventType = "CASH_DRAWER_OP"
	EventLoyaltyEarned  SyncEventType = "LOYALTY_EARNED"
	EventLoyaltySpent   SyncEventType = "LOYALTY_SPENT"
	EventRecipeUpdated  SyncEventType = "RECIPE_UPDATED"
)

// SyncEventStatus
type SyncEventStatus string

const (
	SyncStatusPending  SyncEventStatus = "pending"
	SyncStatusAccepted SyncEventStatus = "accepted"
	SyncStatusConflict SyncEventStatus = "conflict"
	SyncStatusRejected SyncEventStatus = "rejected"
)

// ConflictKind
type ConflictKind string

const (
	ConflictNegativeStock ConflictKind = "negative_stock"
	ConflictStalePrice    ConflictKind = "stale_price"
	ConflictDuplicateUUID ConflictKind = "duplicate_uuid"
	ConflictUnknownProduct ConflictKind = "unknown_product"
	ConflictClockSkew     ConflictKind = "clock_skew"
	ConflictShiftNotOpen  ConflictKind = "shift_not_open"
)

// Claims is embedded in JWT tokens
type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	LocationID uuid.UUID `json:"location_id"`
	DeviceID   uuid.UUID `json:"device_id,omitempty"`
	Role       RoleType  `json:"role"`
	IssuedAt   time.Time `json:"iat"`
	ExpiresAt  time.Time `json:"exp"`
}
