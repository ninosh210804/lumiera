-- ─── Sales heatmap (hourly, refreshed by cron) ────────────────────────────────
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_sales_heatmap AS
SELECT
    location_id,
    date_trunc('hour', created_at AT TIME ZONE 'Asia/Almaty') AS hour_bucket,
    EXTRACT(DOW  FROM created_at AT TIME ZONE 'Asia/Almaty')::INTEGER AS weekday,
    EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Almaty')::INTEGER AS hour_of_day,
    COUNT(*)                        AS orders_count,
    SUM(total)                      AS revenue,
    SUM(total - cost_total)         AS margin
FROM orders
WHERE status = 'paid' AND deleted_at IS NULL
GROUP BY 1, 2, 3, 4
WITH DATA;

CREATE UNIQUE INDEX ON mv_sales_heatmap (location_id, hour_bucket);

-- ─── Product ABC analysis (last 90 days) ──────────────────────────────────────
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_product_abc AS
SELECT
    o.location_id,
    oi.product_id,
    SUM(oi.qty)                             AS units_sold,
    SUM(oi.line_total)                      AS revenue,
    SUM(oi.line_total - oi.line_cost)       AS margin,
    ROUND(
        SUM(oi.line_total - oi.line_cost)
        / NULLIF(SUM(oi.line_total), 0) * 100, 2
    )                                       AS margin_pct
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE o.status = 'paid'
  AND o.created_at >= NOW() - INTERVAL '90 days'
  AND o.deleted_at IS NULL
GROUP BY 1, 2
WITH DATA;

CREATE UNIQUE INDEX ON mv_product_abc (location_id, product_id);

-- ─── Daily revenue summary ────────────────────────────────────────────────────
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_daily_revenue AS
SELECT
    location_id,
    DATE(created_at AT TIME ZONE 'Asia/Almaty')  AS day,
    COUNT(*)                                      AS orders_count,
    SUM(total)                                    AS revenue,
    SUM(total - cost_total)                       AS margin,
    AVG(total)                                    AS avg_check,
    SUM(cost_total)                               AS total_cost
FROM orders
WHERE status = 'paid' AND deleted_at IS NULL
GROUP BY 1, 2
WITH DATA;

CREATE UNIQUE INDEX ON mv_daily_revenue (location_id, day);
