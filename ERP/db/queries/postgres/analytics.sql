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
GROUP BY u.id, u.full_name
ORDER BY revenue DESC;

-- name: RefreshSalesHeatmap :exec
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_sales_heatmap;

-- name: RefreshProductABC :exec
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_product_abc;

-- name: RefreshDailyRevenue :exec
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_daily_revenue;
