-- ─── Categories ───────────────────────────────────────────────────────────────
CREATE TABLE categories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID        NOT NULL REFERENCES locations(id),
    name        TEXT        NOT NULL,
    icon        TEXT        NOT NULL DEFAULT '',
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  UUID        REFERENCES users(id)
);

-- ─── Products ─────────────────────────────────────────────────────────────────
CREATE TABLE products (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID        NOT NULL REFERENCES locations(id),
    category_id     UUID        NOT NULL REFERENCES categories(id),
    name            TEXT        NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    sku             TEXT,
    base_price      NUMERIC(12,2) NOT NULL CHECK (base_price >= 0),
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    is_stop_listed  BOOLEAN     NOT NULL DEFAULT FALSE,
    image_url       TEXT,
    sort_order      INTEGER     NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      UUID        REFERENCES users(id),
    UNIQUE (location_id, sku)
);

-- ─── Modifier groups ──────────────────────────────────────────────────────────
CREATE TABLE product_modifier_groups (
    id              UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID    NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name            TEXT    NOT NULL,   -- Объём | Молоко | Сироп | Доп.шот
    selection_type  TEXT    NOT NULL DEFAULT 'single' CHECK (selection_type IN ('single','multi')),
    required        BOOLEAN NOT NULL DEFAULT FALSE,
    min_select      INTEGER NOT NULL DEFAULT 0,
    max_select      INTEGER NOT NULL DEFAULT 1,
    sort_order      INTEGER NOT NULL DEFAULT 0
);

-- ─── Modifier options ─────────────────────────────────────────────────────────
CREATE TABLE modifier_options (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id                UUID        NOT NULL REFERENCES product_modifier_groups(id) ON DELETE CASCADE,
    name                    TEXT        NOT NULL,   -- 250 мл | Овсяное | Ваниль
    price_delta             NUMERIC(12,2) NOT NULL DEFAULT 0,
    linked_ingredient_id    UUID,                   -- FK set after ingredients table exists
    ingredient_qty_delta    NUMERIC(12,4) NOT NULL DEFAULT 0,
    is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order              INTEGER     NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Price history ────────────────────────────────────────────────────────────
CREATE TABLE price_history (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID        NOT NULL REFERENCES products(id),
    price           NUMERIC(12,2) NOT NULL,
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to    TIMESTAMPTZ,
    changed_by      UUID        REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
