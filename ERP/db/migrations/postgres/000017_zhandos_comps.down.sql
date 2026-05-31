-- Restore the revenue views without the comp filter, then drop comp columns.
CREATE OR REPLACE VIEW mv_sales_heatmap AS
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

CREATE OR REPLACE VIEW mv_product_abc AS
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

CREATE OR REPLACE VIEW mv_daily_revenue AS
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

DROP INDEX IF EXISTS idx_orders_comp;
ALTER TABLE orders DROP COLUMN IF EXISTS comp_recipient;
ALTER TABLE orders DROP COLUMN IF EXISTS is_comp;
