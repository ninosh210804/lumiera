-- name: GetIngredient :one
SELECT * FROM ingredients WHERE id = $1 AND deleted_at IS NULL;

-- name: ListIngredients :many
SELECT * FROM ingredients
WHERE location_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: ListLowStockIngredients :many
SELECT * FROM ingredients
WHERE location_id = $1
  AND current_qty <= min_stock_alert
  AND is_active = TRUE
  AND deleted_at IS NULL
ORDER BY (current_qty / NULLIF(min_stock_alert, 0)) ASC;

-- name: CreateIngredient :one
INSERT INTO ingredients (location_id, name, unit, is_perishable, default_shelf_life_days, min_stock_alert, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateIngredient :one
UPDATE ingredients
SET name = $2, unit = $3, is_perishable = $4, default_shelf_life_days = $5, min_stock_alert = $6, is_active = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteIngredient :exec
UPDATE ingredients SET deleted_at = NOW() WHERE id = $1;

-- name: CreateStockMovement :one
INSERT INTO stock_movements (
    location_id, ingredient_id, batch_id,
    qty_delta, unit_cost_snapshot, reason,
    order_id, inventory_count_id, note, client_uuid, created_by
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetStockMovements :many
SELECT sm.*, i.name AS ingredient_name
FROM stock_movements sm
JOIN ingredients i ON i.id = sm.ingredient_id
WHERE sm.location_id = $1
  AND ($2::uuid IS NULL OR sm.ingredient_id = $2)
  AND sm.created_at BETWEEN $3 AND $4
ORDER BY sm.created_at DESC
LIMIT $5;

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (location_id, supplier_id, notes, client_uuid, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ReceivePurchaseOrder :one
UPDATE purchase_orders
SET status = 'received', received_at = CURRENT_DATE, total_amount = $2
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: CreateInventoryCount :one
INSERT INTO inventory_counts (location_id, performed_by, client_uuid)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CompleteInventoryCount :one
UPDATE inventory_counts
SET status = 'completed', completed_at = NOW()
WHERE id = $1 AND status = 'open'
RETURNING *;

-- name: GetStopListedProducts :many
SELECT id, name, location_id
FROM products
WHERE location_id = $1 AND is_stop_listed = TRUE AND deleted_at IS NULL;

-- name: GetRecipeForProduct :many
SELECT ri.*, i.name AS ingredient_name, i.unit, i.current_avg_cost, i.current_qty
FROM recipe_items ri
LEFT JOIN ingredients i ON i.id = ri.ingredient_id
WHERE ri.recipe_id = $1;

-- ─── Purchase orders ──────────────────────────────────────────────────────────

-- name: GetPurchaseOrder :one
SELECT * FROM purchase_orders WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPurchaseOrders :many
SELECT * FROM purchase_orders
WHERE location_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 50;

-- name: CancelPurchaseOrder :one
UPDATE purchase_orders
SET status = 'cancelled'
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: AddPurchaseOrderItem :one
INSERT INTO purchase_order_items (purchase_order_id, ingredient_id, qty, unit_cost, expires_at)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ListPurchaseOrderItems :many
SELECT poi.*, i.name AS ingredient_name, i.unit
FROM purchase_order_items poi
JOIN ingredients i ON i.id = poi.ingredient_id
WHERE poi.purchase_order_id = $1;

-- ─── Stock batches ────────────────────────────────────────────────────────────

-- name: CreateStockBatch :one
INSERT INTO stock_batches (ingredient_id, purchase_order_item_id, qty_received, qty_remaining, unit_cost, expires_at)
VALUES ($1, $2, $3, $3, $4, $5) RETURNING *;

-- ─── Suppliers ────────────────────────────────────────────────────────────────

-- name: ListSuppliers :many
SELECT * FROM suppliers WHERE deleted_at IS NULL ORDER BY name;

-- name: CreateSupplier :one
INSERT INTO suppliers (name, contact, phone, bin, address)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateSupplier :one
UPDATE suppliers
SET name = $2, contact = $3, phone = $4, bin = $5, address = $6, is_active = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteSupplier :exec
UPDATE suppliers SET deleted_at = NOW() WHERE id = $1;

-- ─── Inventory counts ─────────────────────────────────────────────────────────

-- name: GetInventoryCount :one
SELECT * FROM inventory_counts WHERE id = $1;

-- name: ListInventoryCounts :many
SELECT * FROM inventory_counts
WHERE location_id = $1
ORDER BY created_at DESC
LIMIT 20;

-- name: AddInventoryCountItem :one
INSERT INTO inventory_count_items (inventory_count_id, ingredient_id, expected_qty, actual_qty)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetInventoryCountItems :many
SELECT ici.*, i.name AS ingredient_name, i.unit
FROM inventory_count_items ici
JOIN ingredients i ON i.id = ici.ingredient_id
WHERE ici.inventory_count_id = $1
ORDER BY i.name;

-- ─── Stop-list automation ─────────────────────────────────────────────────────

-- name: GetProductsByIngredient :many
SELECT DISTINCT p.id, p.name, p.location_id
FROM products p
JOIN recipes r ON r.id = p.recipe_id
JOIN recipe_items ri ON ri.recipe_id = r.id
WHERE ri.ingredient_id = $1
  AND p.is_active = TRUE
  AND p.deleted_at IS NULL;
