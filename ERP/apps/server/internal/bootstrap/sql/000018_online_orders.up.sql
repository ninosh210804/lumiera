-- Customer-placed delivery/online orders. These have their own fulfilment
-- lifecycle (new → preparing → ready → completed) separate from orders.status
-- (a financial status). When a barista hands the order over and collects payment
-- it is converted into a real `orders` row (see OnlineOrderService.Complete), so
-- revenue, stock deduction, and loyalty all flow through the existing POS path.
CREATE TABLE online_orders (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID        NOT NULL REFERENCES locations(id),
    customer_id     UUID        REFERENCES customers(id),
    customer_phone  TEXT        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'new'
                    CHECK (status IN ('new','preparing','ready','completed','rejected')),
    delivery_office TEXT        NOT NULL DEFAULT '',
    delivery_note   TEXT        NOT NULL DEFAULT '',
    subtotal        NUMERIC(12,2) NOT NULL DEFAULT 0,   -- price snapshot for display
    total           NUMERIC(12,2) NOT NULL DEFAULT 0,
    order_id        UUID        REFERENCES orders(id),  -- set when completed → real order
    accepted_by     UUID        REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE online_order_items (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    online_order_id     UUID        NOT NULL REFERENCES online_orders(id) ON DELETE CASCADE,
    product_id          UUID        NOT NULL REFERENCES products(id),
    qty                 NUMERIC(12,3) NOT NULL DEFAULT 1 CHECK (qty > 0),
    unit_price_snapshot NUMERIC(12,2) NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fast lookup of the active queue for a location.
CREATE INDEX ix_online_orders_active
    ON online_orders (location_id, created_at)
    WHERE status IN ('new','preparing','ready');
