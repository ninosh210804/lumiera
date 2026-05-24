-- ─── Suppliers ────────────────────────────────────────────────────────────────
CREATE TABLE suppliers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    contact     TEXT        NOT NULL DEFAULT '',
    phone       TEXT,
    bin         TEXT,        -- БИН/ИИН Kazakhstan
    address     TEXT,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- ─── Ingredients ──────────────────────────────────────────────────────────────
CREATE TABLE ingredients (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id             UUID        NOT NULL REFERENCES locations(id),
    name                    TEXT        NOT NULL,
    unit                    TEXT        NOT NULL CHECK (unit IN ('g','kg','ml','l','pcs')),
    is_perishable           BOOLEAN     NOT NULL DEFAULT FALSE,
    default_shelf_life_days INTEGER,
    min_stock_alert         NUMERIC(12,4) NOT NULL DEFAULT 0,
    cost_method             TEXT        NOT NULL DEFAULT 'AVG' CHECK (cost_method IN ('AVG','FIFO')),
    current_avg_cost        NUMERIC(12,4) NOT NULL DEFAULT 0,
    current_qty             NUMERIC(12,4) NOT NULL DEFAULT 0,
    is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    created_by              UUID        REFERENCES users(id)
);

-- Wire modifier_options FK now that ingredients exists
ALTER TABLE modifier_options
    ADD CONSTRAINT fk_modifier_options_ingredient
    FOREIGN KEY (linked_ingredient_id) REFERENCES ingredients(id);

-- ─── Stock batches (FEFO) ─────────────────────────────────────────────────────
CREATE TABLE stock_batches (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ingredient_id           UUID        NOT NULL REFERENCES ingredients(id),
    purchase_order_item_id  UUID,       -- FK added after purchase_order_items exists
    qty_received            NUMERIC(12,4) NOT NULL CHECK (qty_received > 0),
    qty_remaining           NUMERIC(12,4) NOT NULL,
    unit_cost               NUMERIC(12,4) NOT NULL CHECK (unit_cost >= 0),
    received_at             DATE        NOT NULL DEFAULT CURRENT_DATE,
    expires_at              DATE,
    client_uuid             UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Purchase orders ──────────────────────────────────────────────────────────
CREATE TABLE purchase_orders (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID        NOT NULL REFERENCES locations(id),
    supplier_id     UUID        REFERENCES suppliers(id),
    status          TEXT        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','received','cancelled')),
    received_at     DATE,
    total_amount    NUMERIC(12,2) NOT NULL DEFAULT 0,
    notes           TEXT        NOT NULL DEFAULT '',
    client_uuid     UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      UUID        REFERENCES users(id)
);

-- ─── Purchase order items ─────────────────────────────────────────────────────
CREATE TABLE purchase_order_items (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id   UUID        NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    ingredient_id       UUID        NOT NULL REFERENCES ingredients(id),
    qty                 NUMERIC(12,4) NOT NULL CHECK (qty > 0),
    unit_cost           NUMERIC(12,4) NOT NULL CHECK (unit_cost >= 0),
    expires_at          DATE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Wire purchase_order_item FK on stock_batches
ALTER TABLE stock_batches
    ADD CONSTRAINT fk_stock_batches_poi
    FOREIGN KEY (purchase_order_item_id) REFERENCES purchase_order_items(id);

-- ─── Stock movements ──────────────────────────────────────────────────────────
CREATE TABLE stock_movements (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id         UUID        NOT NULL REFERENCES locations(id),
    ingredient_id       UUID        NOT NULL REFERENCES ingredients(id),
    batch_id            UUID        REFERENCES stock_batches(id),
    qty_delta           NUMERIC(12,4) NOT NULL,  -- positive = in, negative = out
    unit_cost_snapshot  NUMERIC(12,4) NOT NULL DEFAULT 0,
    reason              TEXT        NOT NULL CHECK (reason IN ('sale','waste','spill','gift','staff','count','purchase','adjustment')),
    order_id            UUID,       -- FK added after orders table
    inventory_count_id  UUID,       -- FK added after inventory_counts
    note                TEXT,
    client_uuid         UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID        REFERENCES users(id)
);

-- ─── Inventory counts ─────────────────────────────────────────────────────────
CREATE TABLE inventory_counts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID        NOT NULL REFERENCES locations(id),
    status          TEXT        NOT NULL DEFAULT 'open' CHECK (status IN ('open','completed','cancelled')),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    notes           TEXT        NOT NULL DEFAULT '',
    client_uuid     UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    performed_by    UUID        REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory_count_items (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_count_id  UUID        NOT NULL REFERENCES inventory_counts(id) ON DELETE CASCADE,
    ingredient_id       UUID        NOT NULL REFERENCES ingredients(id),
    expected_qty        NUMERIC(12,4) NOT NULL DEFAULT 0,
    actual_qty          NUMERIC(12,4) NOT NULL DEFAULT 0,
    variance            NUMERIC(12,4) GENERATED ALWAYS AS (actual_qty - expected_qty) STORED,
    note                TEXT
);

-- Wire inventory_count FK on stock_movements
ALTER TABLE stock_movements
    ADD CONSTRAINT fk_stock_movements_count
    FOREIGN KEY (inventory_count_id) REFERENCES inventory_counts(id);
