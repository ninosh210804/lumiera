-- name: CreateOrder :one
INSERT INTO orders (
    id, location_id, shift_id, barista_id, customer_id,
    subtotal, discount_total, loyalty_points_used, total, cost_total,
    receipt_no, status, client_uuid, created_by, synced
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'open', ?, ?, 0)
RETURNING *;

-- name: PayOrder :one
UPDATE orders
SET status = 'paid', updated_at = datetime('now')
WHERE id = ? AND status = 'open'
RETURNING *;

-- name: CancelOrder :exec
UPDATE orders
SET status        = 'cancelled',
    cancel_reason = ?,
    cancelled_by  = ?,
    cancelled_at  = datetime('now'),
    updated_at    = datetime('now')
WHERE id = ? AND status IN ('open','paid');

-- name: GetOrder :one
SELECT * FROM orders WHERE id = ? AND deleted_at IS NULL;

-- name: GetOrdersByShift :many
SELECT * FROM orders WHERE shift_id = ? AND deleted_at IS NULL ORDER BY created_at DESC;

-- name: CreateOrderItem :one
INSERT INTO order_items (id, order_id, product_id, qty, unit_price_snapshot, line_total, line_cost, client_uuid)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetOrderItems :many
SELECT oi.*, p.name AS product_name
FROM order_items oi
JOIN products p ON p.id = oi.product_id
WHERE oi.order_id = ?;

-- name: CreateOrderItemModifier :exec
INSERT INTO order_item_modifiers (id, order_item_id, modifier_option_id, price_delta_snapshot)
VALUES (?, ?, ?, ?);

-- name: CreatePayment :exec
INSERT INTO payments (id, order_id, payment_method_id, amount, external_ref)
VALUES (?, ?, ?, ?, ?);

-- name: CreateStockMovement :exec
INSERT INTO stock_movements (
    id, location_id, ingredient_id, qty_delta,
    unit_cost_snapshot, reason, order_id, note, client_uuid, created_by, synced
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0);

-- name: GetPendingSyncItems :many
SELECT client_uuid, sequence, event_type, payload, device_ts, retry_count
FROM sync_outbox
WHERE status = 'pending'
ORDER BY created_at
LIMIT ?;

-- name: InsertSyncEvent :exec
INSERT INTO sync_outbox (client_uuid, sequence, event_type, payload, device_ts)
VALUES (?, ?, ?, ?, ?);

-- name: MarkSyncAccepted :exec
UPDATE sync_outbox SET status = 'accepted' WHERE client_uuid = ?;

-- name: MarkSyncConflict :exec
UPDATE sync_outbox
SET status = 'conflict', last_error = ?
WHERE client_uuid = ?;

-- name: IncrementRetry :exec
UPDATE sync_outbox
SET retry_count = retry_count + 1, last_error = ?
WHERE client_uuid = ?;

-- name: GetDeviceInfo :one
SELECT * FROM device_info LIMIT 1;

-- name: IncrementDeviceSequence :one
UPDATE device_info SET sequence = sequence + 1 RETURNING sequence;
