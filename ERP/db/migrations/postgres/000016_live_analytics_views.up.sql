-- Make analytics live.
--
-- mv_daily_revenue, mv_sales_heatmap and mv_product_abc were MATERIALIZED
-- views, so the dashboard and analytics pages only showed data through the
-- last REFRESH MATERIALIZED VIEW call. Today's orders never appeared until an
-- admin clicked "Обновить" — hence /1 (analytics page missing today) and /2
-- (dashboard "Сегодня" empty).
--
-- For a single-shop ERP these aggregates are tiny (one row per day / per
-- product-90d / per weekday-hour), so a regular view is fast enough and is
-- always live. Drop the materialized variants, recreate as plain views with
-- the same column shape. The Refresh* SQL queries become no-ops via a
-- subsequent code change.

DROP MATERIALIZED VIEW IF EXISTS mv_sales_heatmap;
CREATE VIEW mv_sales_heatmap AS
SELECT
    location_id,
    date_trunc('hour', created_at AT TIME ZONE 'Asia/Almaty') AS hour_bucket,
    EXTRACT(DOW  FROM created_at AT TIME ZONE 'Asia/Almaty')::INTEGER AS weekday,
    EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Almaty')::INTEGER AS hour_of_day,
    COUNT(*)                                  AS orders_count,
    ROUND(SUM(total))::BIGINT                 AS revenue,
    ROUND(SUM(total - cost_total))::BIGINT    AS margin
FROM orders
WHERE status = 'paid' AND deleted_at IS NULL
GROUP BY 1, 2, 3, 4;

DROP MATERIALIZED VIEW IF EXISTS mv_product_abc;
CREATE VIEW mv_product_abc AS
SELECT
    o.location_id,
    oi.product_id,
    ROUND(SUM(oi.qty))::BIGINT                        AS units_sold,
    ROUND(SUM(oi.line_total))::BIGINT                 AS revenue,
    ROUND(SUM(oi.line_total - oi.line_cost))::BIGINT  AS margin,
    ROUND(
        SUM(oi.line_total - oi.line_cost)
        / NULLIF(SUM(oi.line_total), 0) * 100, 2
    )                                                 AS margin_pct
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE o.status = 'paid'
  AND o.created_at >= NOW() - INTERVAL '90 days'
  AND o.deleted_at IS NULL
GROUP BY 1, 2;

DROP MATERIALIZED VIEW IF EXISTS mv_daily_revenue;
CREATE VIEW mv_daily_revenue AS
SELECT
    location_id,
    DATE(created_at AT TIME ZONE 'Asia/Almaty')   AS day,
    COUNT(*)                                       AS orders_count,
    ROUND(SUM(total))::BIGINT                      AS revenue,
    ROUND(SUM(total - cost_total))::BIGINT         AS margin,
    AVG(total)                                     AS avg_check,
    ROUND(SUM(cost_total))::BIGINT                 AS total_cost
FROM orders
WHERE status = 'paid' AND deleted_at IS NULL
GROUP BY 1, 2;
