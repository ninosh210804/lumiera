-- ─── Recipes ──────────────────────────────────────────────────────────────────
CREATE TABLE recipes (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID        NOT NULL REFERENCES locations(id),
    name            TEXT        NOT NULL,
    recipe_type     TEXT        NOT NULL DEFAULT 'product' CHECK (recipe_type IN ('product','semi_finished')),
    yield_qty       NUMERIC(12,4) NOT NULL DEFAULT 1,
    yield_unit      TEXT        NOT NULL DEFAULT 'pcs',
    notes           TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      UUID        REFERENCES users(id)
);

-- ─── Recipe items (ingredients OR sub-recipes) ────────────────────────────────
CREATE TABLE recipe_items (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id       UUID        NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id   UUID        REFERENCES ingredients(id),
    sub_recipe_id   UUID        REFERENCES recipes(id),    -- self-reference for semi-finished
    qty             NUMERIC(12,4) NOT NULL CHECK (qty > 0),
    unit            TEXT        NOT NULL,
    sort_order      INTEGER     NOT NULL DEFAULT 0,
    -- exactly one of ingredient_id / sub_recipe_id must be set
    CONSTRAINT chk_recipe_item_source
        CHECK (
            (ingredient_id IS NOT NULL AND sub_recipe_id IS NULL) OR
            (ingredient_id IS NULL     AND sub_recipe_id IS NOT NULL)
        )
);

-- Link products to their recipes
ALTER TABLE products
    ADD COLUMN recipe_id UUID REFERENCES recipes(id);
