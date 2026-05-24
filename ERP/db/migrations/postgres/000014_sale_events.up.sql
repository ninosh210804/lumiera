-- Sale events: a togglable promotion that overrides prices for selected products.
CREATE TABLE IF NOT EXISTS sale_events (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID        NOT NULL REFERENCES locations(id),
    name        TEXT        NOT NULL,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    starts_at   TIMESTAMPTZ,
    ends_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID        REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS sale_event_items (
    sale_event_id UUID          NOT NULL REFERENCES sale_events(id) ON DELETE CASCADE,
    product_id    UUID          NOT NULL REFERENCES products(id),
    sale_price    NUMERIC(12,2) NOT NULL CHECK (sale_price >= 0),
    PRIMARY KEY (sale_event_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_sale_event_items_product ON sale_event_items(product_id);
