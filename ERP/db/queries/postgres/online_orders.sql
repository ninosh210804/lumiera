-- name: CreateOnlineOrder :one
INSERT INTO online_orders (
    location_id, customer_id, customer_phone,
    delivery_office, delivery_note, subtotal, total
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: CreateOnlineOrderItem :one
INSERT INTO online_order_items (
    online_order_id, product_id, qty, unit_price_snapshot
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetOnlineOrder :one
SELECT * FROM online_orders WHERE id = $1;

-- name: GetOnlineOrderItems :many
SELECT oi.*, p.name AS product_name
FROM online_order_items oi
JOIN products p ON p.id = oi.product_id
WHERE oi.online_order_id = $1
ORDER BY oi.created_at;

-- name: ListActiveOnlineOrders :many
SELECT * FROM online_orders
WHERE location_id = $1 AND status IN ('new','preparing','ready')
ORDER BY created_at ASC;

-- name: SetOnlineOrderStatus :one
UPDATE online_orders
SET status = $2, accepted_by = COALESCE($3::uuid, accepted_by), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CompleteOnlineOrder :one
UPDATE online_orders
SET status = 'completed', order_id = $2, updated_at = NOW()
WHERE id = $1 AND status <> 'completed'
RETURNING *;
