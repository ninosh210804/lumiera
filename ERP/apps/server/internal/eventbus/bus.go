package eventbus

import (
	"sync"
	"time"
)

// EventKind identifies what happened.
type EventKind string

const (
	EventSalePaid      EventKind = "sale.paid"
	EventStockReceived EventKind = "stock.received"
	EventShiftOpened   EventKind = "shift.opened"
	EventShiftClosed   EventKind = "shift.closed"
)

// Event carries context about what happened.
type Event struct {
	Kind      EventKind
	OccurredAt time.Time
	// Payload is specific to the EventKind; consumers type-assert.
	Payload any
}

// SalePaidPayload is published with EventSalePaid.
type SalePaidPayload struct {
	OrderID    string
	LocationID string
	Total      float64
	CostTotal  float64
}

// StockReceivedPayload is published with EventStockReceived.
type StockReceivedPayload struct {
	PurchaseOrderID string
	LocationID      string
	TotalAmount     float64
}

// Handler is a function that processes an event.
type Handler func(Event)

// Bus is a simple in-process pub/sub bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[EventKind][]Handler
}

func New() *Bus {
	return &Bus{handlers: make(map[EventKind][]Handler)}
}

// Subscribe registers h to receive events of kind k.
func (b *Bus) Subscribe(k EventKind, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[k] = append(b.handlers[k], h)
}

// Publish sends e to all subscribers. Never blocks: handlers run in a goroutine.
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	hs := b.handlers[e.Kind]
	b.mu.RUnlock()
	for _, h := range hs {
		go h(e)
	}
}
