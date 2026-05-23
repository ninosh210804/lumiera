PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;

-- ─── Locations ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS locations (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    address     TEXT NOT NULL DEFAULT '',
    city        TEXT NOT NULL DEFAULT 'Almaty',
    timezone    TEXT NOT NULL DEFAULT 'Asia/Almaty',
    phone       TEXT,
    is_active   INTEGER NOT NULL DEFAULT 1,
    settings    TEXT NOT NULL DEFAULT '{}',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

-- ─── Roles & Permissions ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
    id          TEXT PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    is_system   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS permissions (
    id          TEXT PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- ─── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id                  TEXT PRIMARY KEY,
    default_location_id TEXT NOT NULL REFERENCES locations(id),
    role_id             TEXT NOT NULL REFERENCES roles(id),
    full_name           TEXT NOT NULL,
    email               TEXT UNIQUE,
    pin_hash            TEXT NOT NULL,
    is_active           INTEGER NOT NULL DEFAULT 1,
    created_at          TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at          TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at          TEXT,
    created_by          TEXT REFERENCES users(id)
);

-- ─── Categories ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS categories (
    id          TEXT PRIMARY KEY,
    location_id TEXT NOT NULL REFERENCES locations(id),
    name        TEXT NOT NULL,
    icon        TEXT NOT NULL DEFAULT '',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    is_active   INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

-- ─── Products ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS products (
    id              TEXT PRIMARY KEY,
    location_id     TEXT NOT NULL REFERENCES locations(id),
    category_id     TEXT NOT NULL REFERENCES categories(id),
    recipe_id       TEXT,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    sku             TEXT,
    base_price      REAL NOT NULL DEFAULT 0,
    is_active       INTEGER NOT NULL DEFAULT 1,
    is_stop_listed  INTEGER NOT NULL DEFAULT 0,
    image_url       TEXT,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at      TEXT
);

-- ─── Modifier groups & options ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS product_modifier_groups (
    id              TEXT PRIMARY KEY,
    product_id      TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    selection_type  TEXT NOT NULL DEFAULT 'single',
    required        INTEGER NOT NULL DEFAULT 0,
    min_select      INTEGER NOT NULL DEFAULT 0,
    max_select      INTEGER NOT NULL DEFAULT 1,
    sort_order      INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS modifier_options (
    id                      TEXT PRIMARY KEY,
    group_id                TEXT NOT NULL REFERENCES product_modifier_groups(id) ON DELETE CASCADE,
    name                    TEXT NOT NULL,
    price_delta             REAL NOT NULL DEFAULT 0,
    linked_ingredient_id    TEXT REFERENCES ingredients(id),
    ingredient_qty_delta    REAL NOT NULL DEFAULT 0,
    is_active               INTEGER NOT NULL DEFAULT 1,
    sort_order              INTEGER NOT NULL DEFAULT 0,
    created_at              TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Ingredients ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ingredients (
    id                      TEXT PRIMARY KEY,
    location_id             TEXT NOT NULL REFERENCES locations(id),
    name                    TEXT NOT NULL,
    unit                    TEXT NOT NULL DEFAULT 'pcs',
    is_perishable           INTEGER NOT NULL DEFAULT 0,
    default_shelf_life_days INTEGER,
    min_stock_alert         REAL NOT NULL DEFAULT 0,
    current_avg_cost        REAL NOT NULL DEFAULT 0,
    current_qty             REAL NOT NULL DEFAULT 0,
    is_active               INTEGER NOT NULL DEFAULT 1,
    created_at              TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at              TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at              TEXT
);

-- ─── Recipes ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS recipes (
    id          TEXT PRIMARY KEY,
    location_id TEXT NOT NULL REFERENCES locations(id),
    name        TEXT NOT NULL,
    recipe_type TEXT NOT NULL DEFAULT 'product',
    yield_qty   REAL NOT NULL DEFAULT 1,
    yield_unit  TEXT NOT NULL DEFAULT 'pcs',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

CREATE TABLE IF NOT EXISTS recipe_items (
    id              TEXT PRIMARY KEY,
    recipe_id       TEXT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id   TEXT REFERENCES ingredients(id),
    sub_recipe_id   TEXT REFERENCES recipes(id),
    qty             REAL NOT NULL,
    unit            TEXT NOT NULL,
    sort_order      INTEGER NOT NULL DEFAULT 0
);

-- ─── Payment methods ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payment_methods (
    id            TEXT PRIMARY KEY,
    code          TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    opens_drawer  INTEGER NOT NULL DEFAULT 0,
    is_active     INTEGER NOT NULL DEFAULT 1,
    sort_order    INTEGER NOT NULL DEFAULT 0
);

-- ─── Customers & Loyalty ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS customers (
    id          TEXT PRIMARY KEY,
    phone       TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    birthday    TEXT,
    source      TEXT NOT NULL DEFAULT 'manual',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

CREATE TABLE IF NOT EXISTS loyalty_accounts (
    id                  TEXT PRIMARY KEY,
    customer_id         TEXT NOT NULL UNIQUE REFERENCES customers(id),
    points_balance      REAL NOT NULL DEFAULT 0,
    free_drinks_left    INTEGER NOT NULL DEFAULT 0,
    total_visits        INTEGER NOT NULL DEFAULT 0,
    created_at          TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at          TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS loyalty_rules (
    id          TEXT PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    params      TEXT NOT NULL DEFAULT '{}',
    is_active   INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Shifts ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shifts (
    id                      TEXT PRIMARY KEY,
    location_id             TEXT NOT NULL REFERENCES locations(id),
    user_id                 TEXT NOT NULL REFERENCES users(id),
    opened_at               TEXT NOT NULL DEFAULT (datetime('now')),
    closed_at               TEXT,
    opening_cash            REAL NOT NULL DEFAULT 0,
    closing_cash_expected   REAL,
    closing_cash_actual     REAL,
    auto_opened             INTEGER NOT NULL DEFAULT 0,
    client_uuid             TEXT NOT NULL UNIQUE,
    created_at              TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at              TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Orders ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS orders (
    id                      TEXT PRIMARY KEY,
    location_id             TEXT NOT NULL REFERENCES locations(id),
    shift_id                TEXT REFERENCES shifts(id),
    barista_id              TEXT NOT NULL REFERENCES users(id),
    customer_id             TEXT REFERENCES customers(id),
    status                  TEXT NOT NULL DEFAULT 'open',
    subtotal                REAL NOT NULL DEFAULT 0,
    discount_total          REAL NOT NULL DEFAULT 0,
    loyalty_points_used     REAL NOT NULL DEFAULT 0,
    total                   REAL NOT NULL DEFAULT 0,
    cost_total              REAL NOT NULL DEFAULT 0,
    receipt_no              TEXT NOT NULL,
    cancel_reason           TEXT,
    cancelled_by            TEXT REFERENCES users(id),
    cancelled_at            TEXT,
    client_uuid             TEXT NOT NULL UNIQUE,
    created_at              TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at              TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at              TEXT,
    created_by              TEXT REFERENCES users(id),
    synced                  INTEGER NOT NULL DEFAULT 0  -- 0=pending sync, 1=synced
);

CREATE TABLE IF NOT EXISTS order_items (
    id                      TEXT PRIMARY KEY,
    order_id                TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id              TEXT NOT NULL REFERENCES products(id),
    qty                     REAL NOT NULL DEFAULT 1,
    unit_price_snapshot     REAL NOT NULL,
    line_total              REAL NOT NULL,
    line_cost               REAL NOT NULL DEFAULT 0,
    client_uuid             TEXT NOT NULL UNIQUE,
    created_at              TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS order_item_modifiers (
    id                      TEXT PRIMARY KEY,
    order_item_id           TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    modifier_option_id      TEXT NOT NULL REFERENCES modifier_options(id),
    price_delta_snapshot    REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
    id                  TEXT PRIMARY KEY,
    order_id            TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_method_id   TEXT NOT NULL REFERENCES payment_methods(id),
    amount              REAL NOT NULL,
    external_ref        TEXT,
    created_at          TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS refunds (
    id              TEXT PRIMARY KEY,
    order_id        TEXT NOT NULL REFERENCES orders(id),
    amount          REAL NOT NULL,
    reason          TEXT NOT NULL,
    approved_by     TEXT REFERENCES users(id),
    client_uuid     TEXT NOT NULL UNIQUE,
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS loyalty_transactions (
    id                  TEXT PRIMARY KEY,
    loyalty_account_id  TEXT NOT NULL REFERENCES loyalty_accounts(id),
    order_id            TEXT REFERENCES orders(id),
    kind                TEXT NOT NULL,
    points_delta        REAL NOT NULL,
    rule_id             TEXT REFERENCES loyalty_rules(id),
    note                TEXT,
    client_uuid         TEXT NOT NULL UNIQUE,
    created_at          TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Stock movements (local) ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS stock_movements (
    id                  TEXT PRIMARY KEY,
    location_id         TEXT NOT NULL REFERENCES locations(id),
    ingredient_id       TEXT NOT NULL REFERENCES ingredients(id),
    qty_delta           REAL NOT NULL,
    unit_cost_snapshot  REAL NOT NULL DEFAULT 0,
    reason              TEXT NOT NULL,
    order_id            TEXT REFERENCES orders(id),
    note                TEXT,
    client_uuid         TEXT NOT NULL UNIQUE,
    created_at          TEXT NOT NULL DEFAULT (datetime('now')),
    created_by          TEXT REFERENCES users(id),
    synced              INTEGER NOT NULL DEFAULT 0
);

-- ─── Cash drawer ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cash_drawer_operations (
    id          TEXT PRIMARY KEY,
    shift_id    TEXT NOT NULL REFERENCES shifts(id),
    kind        TEXT NOT NULL,
    amount      REAL NOT NULL,
    reason      TEXT NOT NULL DEFAULT '',
    client_uuid TEXT NOT NULL UNIQUE,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    created_by  TEXT REFERENCES users(id),
    synced      INTEGER NOT NULL DEFAULT 0
);

-- ─── Sync outbox ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sync_outbox (
    client_uuid     TEXT PRIMARY KEY,
    sequence        INTEGER NOT NULL,
    event_type      TEXT NOT NULL,
    payload         TEXT NOT NULL,
    device_ts       TEXT NOT NULL DEFAULT (datetime('now')),
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending|sent|accepted|conflict
    retry_count     INTEGER NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Sync conflicts (local) ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sync_conflicts (
    id              TEXT PRIMARY KEY,
    client_uuid     TEXT NOT NULL REFERENCES sync_outbox(client_uuid),
    kind            TEXT NOT NULL,
    details         TEXT NOT NULL DEFAULT '{}',
    needs_review    INTEGER NOT NULL DEFAULT 1,
    resolved_at     TEXT,
    resolution      TEXT,
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Device info (self) ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS device_info (
    id          TEXT PRIMARY KEY,
    server_id   TEXT,   -- UUID from server device_registry
    device_name TEXT NOT NULL,
    platform    TEXT NOT NULL DEFAULT 'unknown',
    app_version TEXT NOT NULL DEFAULT '0.0.0',
    location_id TEXT REFERENCES locations(id),
    sequence    INTEGER NOT NULL DEFAULT 0,  -- monotonic counter for this device
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ─── Indexes ──────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_orders_location     ON orders (location_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_pending_sync ON orders (synced) WHERE synced = 0;
CREATE INDEX IF NOT EXISTS idx_orders_shift        ON orders (shift_id);
CREATE INDEX IF NOT EXISTS idx_products_active     ON products (location_id, is_active);
CREATE INDEX IF NOT EXISTS idx_ingredients_loc     ON ingredients (location_id, is_active);
CREATE INDEX IF NOT EXISTS idx_outbox_pending      ON sync_outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_movements_pending   ON stock_movements (synced) WHERE synced = 0;
CREATE INDEX IF NOT EXISTS idx_shifts_open         ON shifts (user_id, closed_at) WHERE closed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_phone     ON customers (phone);
