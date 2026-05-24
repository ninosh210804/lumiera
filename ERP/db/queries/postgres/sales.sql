-- name: ListSaleEvents :many
SELECT se.*,
    (SELECT COUNT(*) FROM sale_event_items sei WHERE sei.sale_event_id = se.id) AS item_count
FROM sale_events se
WHERE se.location_id = $1
ORDER BY se.created_at DESC;

-- name: GetSaleEvent :one
SELECT * FROM sale_events WHERE id = $1;

-- name: CreateSaleEvent :one
INSERT INTO sale_events (location_id, name, is_active, starts_at, ends_at, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: SetSaleEventActive :one
UPDATE sale_events SET is_active = $2 WHERE id = $1 RETURNING *;

-- name: DeleteSaleEvent :exec
DELETE FROM sale_events WHERE id = $1;

-- name: ListSaleEventItems :many
SELECT sei.sale_event_id, sei.product_id, sei.sale_price, p.name AS product_name, p.base_price
FROM sale_event_items sei
JOIN products p ON p.id = sei.product_id
WHERE sei.sale_event_id = $1
ORDER BY p.name;

-- name: AddSaleEventItem :one
INSERT INTO sale_event_items (sale_event_id, product_id, sale_price)
VALUES ($1, $2, $3)
ON CONFLICT (sale_event_id, product_id) DO UPDATE SET sale_price = EXCLUDED.sale_price
RETURNING *;

-- name: RemoveSaleEventItem :exec
DELETE FROM sale_event_items WHERE sale_event_id = $1 AND product_id = $2;
