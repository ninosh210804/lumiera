-- ─── Orders ───────────────────────────────────────────────────────────────────
CREATE INDEX idx_orders_location_time    ON orders (location_id, created_at DESC);
CREATE INDEX idx_orders_customer         ON orders (customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_orders_shift            ON orders (shift_id) WHERE shift_id IS NOT NULL;
CREATE INDEX idx_orders_barista          ON orders (barista_id, created_at DESC);
CREATE INDEX idx_orders_status           ON orders (status) WHERE status NOT IN ('paid');

-- ─── Products ─────────────────────────────────────────────────────────────────
CREATE INDEX idx_products_location       ON products (location_id, is_active);
CREATE INDEX idx_products_category       ON products (category_id);

-- ─── Stock ────────────────────────────────────────────────────────────────────
CREATE INDEX idx_stock_movements_ingr    ON stock_movements (ingredient_id, created_at DESC);
CREATE INDEX idx_stock_movements_order   ON stock_movements (order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_stock_batches_fefo      ON stock_batches (ingredient_id, expires_at)
    WHERE qty_remaining > 0;

-- ─── Sync ─────────────────────────────────────────────────────────────────────
CREATE UNIQUE INDEX uq_sync_events_uuid  ON sync_events (client_uuid);
CREATE INDEX idx_sync_events_device_seq  ON sync_events (device_id, sequence);
CREATE INDEX idx_sync_events_pending     ON sync_events (status) WHERE status = 'conflict';
CREATE INDEX idx_sync_conflicts_review   ON sync_conflicts (needs_review) WHERE needs_review = TRUE;

-- ─── Audit ────────────────────────────────────────────────────────────────────
CREATE INDEX idx_audit_entity            ON audit_log (entity, entity_id);
CREATE INDEX idx_audit_user_time         ON audit_log (user_id, created_at DESC);

-- ─── CRM / Loyalty ────────────────────────────────────────────────────────────
CREATE INDEX idx_customers_phone         ON customers (phone);
CREATE INDEX idx_loyalty_txn_account     ON loyalty_transactions (loyalty_account_id, created_at DESC);

-- ─── Shifts ───────────────────────────────────────────────────────────────────
CREATE INDEX idx_shifts_user_open        ON shifts (user_id, closed_at) WHERE closed_at IS NULL;
CREATE INDEX idx_shifts_location_time    ON shifts (location_id, opened_at DESC);

-- ─── Accounting ───────────────────────────────────────────────────────────────
CREATE INDEX idx_journal_entries_source  ON journal_entries (source_ref) WHERE source_ref IS NOT NULL;
CREATE INDEX idx_expenses_location_date  ON expenses (location_id, paid_on DESC);
