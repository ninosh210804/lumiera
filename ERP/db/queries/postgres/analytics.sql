-- name: GetSalesHeatmap :many
SELECT weekday, hour_of_day, SUM(orders_count) AS orders_count, SUM(revenue) AS revenue, SUM(margin) AS margin
FROM mv_sales_heatmap
WHERE location_id = $1
  AND hour_bucket BETWEEN $2::timestamptz AND $3::timestamptz
GROUP BY weekday, hour_of_day
ORDER BY weekday, hour_of_day;

-- name: GetProductABC :many
SELECT
    pa.*,
    p.name AS product_name,
    p.base_price,
    c.name AS category_name
FROM mv_product_abc pa
JOIN products p ON p.id = pa.product_id
JOIN categories c ON c.id = p.category_id
WHERE pa.location_id = $1
ORDER BY pa.revenue DESC;

-- name: GetDailyRevenue :many
SELECT * FROM mv_daily_revenue
WHERE location_id = $1 AND day BETWEEN $2 AND $3
ORDER BY day;

-- name: GetBaristaStats :many
SELECT
    u.id,
    u.full_name,
    COUNT(o.id)         AS orders_count,
    SUM(o.total)        AS revenue,
    AVG(o.total)        AS avg_check,
    COUNT(DISTINCT o.shift_id) AS shifts_count
FROM orders o
JOIN users u ON u.id = o.barista_id
WHERE o.location_id = $1
  AND o.status = 'paid'
  AND o.created_at BETWEEN $2 AND $3
  AND o.deleted_at IS NULL
  AND o.is_comp = FALSE
GROUP BY u.id, u.full_name
ORDER BY revenue DESC;

-- name: GetCompSummary :many
-- Per-recipient tally of products taken without payment (e.g. Zhandos).
SELECT
    o.comp_recipient,
    COUNT(*)::bigint               AS orders_count,
    ROUND(SUM(o.total))::bigint    AS total_value,
    ROUND(SUM(o.cost_total))::bigint AS total_cost
FROM orders o
WHERE o.location_id = $1
  AND o.is_comp = TRUE
  AND o.status = 'paid'
  AND o.deleted_at IS NULL
  AND o.created_at BETWEEN $2 AND $3
GROUP BY o.comp_recipient
ORDER BY total_value DESC;

-- name: GetCompItems :many
-- What exactly each recipient took, aggregated by product.
SELECT
    o.comp_recipient,
    oi.product_id,
    p.name                            AS product_name,
    ROUND(SUM(oi.qty))::bigint        AS qty,
    ROUND(SUM(oi.line_total))::bigint AS total_value
FROM order_items oi
JOIN orders o   ON o.id = oi.order_id
JOIN products p ON p.id = oi.product_id
WHERE o.location_id = $1
  AND o.is_comp = TRUE
  AND o.status = 'paid'
  AND o.deleted_at IS NULL
  AND o.created_at BETWEEN $2 AND $3
GROUP BY o.comp_recipient, oi.product_id, p.name
ORDER BY total_value DESC;

-- The analytics views are now plain VIEWs (see migration 000016), so refresh
-- is a no-op. We keep the queries so existing code paths and the admin
-- "Обновить" button still compile and respond; SELECT 1 is just a fast ping
-- that keeps the response shape and lets the frontend display "Обновлено".

-- name: RefreshSalesHeatmap :exec
SELECT 1;

-- name: RefreshProductABC :exec
SELECT 1;

-- name: RefreshDailyRevenue :exec
SELECT 1;
