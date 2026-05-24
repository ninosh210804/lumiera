-- ─── Payment methods ──────────────────────────────────────────────────────────
CREATE TABLE payment_methods (
    id            UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    code          TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL,
    opens_drawer  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order    INTEGER NOT NULL DEFAULT 0
);

INSERT INTO payment_methods (code, name, opens_drawer, sort_order) VALUES
    ('cash',     'Наличные',  TRUE,  1),
    ('kaspi_qr', 'Kaspi QR',  FALSE, 2),
    ('loyalty',  'Бонусы',    FALSE, 3),
    ('gift',     'Подарок',   FALSE, 4);

-- ─── Customers ────────────────────────────────────────────────────────────────
CREATE TABLE customers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    phone       TEXT        UNIQUE NOT NULL,
    name        TEXT        NOT NULL DEFAULT '',
    birthday    DATE,
    source      TEXT        NOT NULL DEFAULT 'manual' CHECK (source IN ('manual','qr','app')),
    notes       TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- ─── Loyalty ──────────────────────────────────────────────────────────────────
CREATE TABLE loyalty_accounts (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id         UUID        NOT NULL UNIQUE REFERENCES customers(id),
    points_balance      NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (points_balance >= 0),
    free_drinks_left    INTEGER     NOT NULL DEFAULT 0,
    total_visits        INTEGER     NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loyalty_rules (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    params      JSONB   NOT NULL DEFAULT '{}',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO loyalty_rules (code, name, params) VALUES
    ('earn_1pct',     '1% кэшбэк баллами',       '{"percent": 1}'),
    ('every_6th_free','Каждый 6-й кофе бесплатно','{"every_n": 6, "product_category": "coffee"}'),
    ('birthday_x2',   'Двойные баллы в день рождения','{"multiplier": 2}');

-- ─── Orders ───────────────────────────────────────────────────────────────────
CREATE TABLE shifts (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id             UUID        NOT NULL REFERENCES locations(id),
    user_id                 UUID        NOT NULL REFERENCES users(id),
    opened_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at               TIMESTAMPTZ,
    opening_cash            NUMERIC(12,2) NOT NULL DEFAULT 0,
    closing_cash_expected   NUMERIC(12,2),
    closing_cash_actual     NUMERIC(12,2),
    variance                NUMERIC(12,2) GENERATED ALWAYS AS (closing_cash_actual - closing_cash_expected) STORED,
    auto_opened             BOOLEAN     NOT NULL DEFAULT FALSE,
    client_uuid             UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id             UUID        NOT NULL REFERENCES locations(id),
    shift_id                UUID        REFERENCES shifts(id),
    barista_id              UUID        NOT NULL REFERENCES users(id),
    customer_id             UUID        REFERENCES customers(id),
    status                  TEXT        NOT NULL DEFAULT 'open' CHECK (status IN ('open','paid','refunded','cancelled')),
    subtotal                NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_total          NUMERIC(12,2) NOT NULL DEFAULT 0,
    loyalty_points_used     NUMERIC(12,2) NOT NULL DEFAULT 0,
    total                   NUMERIC(12,2) NOT NULL DEFAULT 0,
    cost_total              NUMERIC(12,2) NOT NULL DEFAULT 0,
    receipt_no              TEXT        NOT NULL,
    cancel_reason           TEXT,
    cancelled_by            UUID        REFERENCES users(id),
    cancelled_at            TIMESTAMPTZ,
    client_uuid             UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    created_by              UUID        REFERENCES users(id)
);

CREATE TABLE order_items (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id              UUID        NOT NULL REFERENCES products(id),
    qty                     NUMERIC(12,3) NOT NULL DEFAULT 1 CHECK (qty > 0),
    unit_price_snapshot     NUMERIC(12,2) NOT NULL,
    line_total              NUMERIC(12,2) NOT NULL,
    line_cost               NUMERIC(12,2) NOT NULL DEFAULT 0,
    client_uuid             UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_item_modifiers (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id           UUID        NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    modifier_option_id      UUID        NOT NULL REFERENCES modifier_options(id),
    price_delta_snapshot    NUMERIC(12,2) NOT NULL
);

CREATE TABLE payments (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_method_id   UUID        NOT NULL REFERENCES payment_methods(id),
    amount              NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    external_ref        TEXT,        -- Kaspi transaction ID (manual entry)
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refunds (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID        NOT NULL REFERENCES orders(id),
    amount          NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    reason          TEXT        NOT NULL,
    approved_by     UUID        REFERENCES users(id),
    client_uuid     UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loyalty_transactions (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    loyalty_account_id  UUID        NOT NULL REFERENCES loyalty_accounts(id),
    order_id            UUID        REFERENCES orders(id),
    kind                TEXT        NOT NULL CHECK (kind IN ('earn','spend','adjust','gift')),
    points_delta        NUMERIC(12,2) NOT NULL,
    rule_id             UUID        REFERENCES loyalty_rules(id),
    note                TEXT,
    client_uuid         UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Wire order FK on stock_movements
ALTER TABLE stock_movements
    ADD CONSTRAINT fk_stock_movements_order
    FOREIGN KEY (order_id) REFERENCES orders(id);
