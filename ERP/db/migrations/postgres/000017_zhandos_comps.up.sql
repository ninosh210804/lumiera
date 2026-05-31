-- "Comps": products taken without payment (e.g. Zhandos took stock off the
-- shelf). They still consume ingredients from the warehouse via the recipe,
-- but they are NOT sales — so they must stay out of revenue/margin reporting
-- and be tracked separately as a running tally per recipient.

ALTER TABLE orders ADD COLUMN IF NOT EXISTS is_comp        BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS comp_recipient TEXT    NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_orders_comp
    ON orders (location_id, comp_recipient)
    WHERE is_comp = TRUE AND deleted_at IS NULL;

-- Recreate the live analytics views so comps never count as revenue.
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
WHERE status = 'paid' AND deleted_at IS NULL AND is_comp = FALSE
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
  AND o.is_comp = FALSE
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
WHERE status = 'paid' AND deleted_at IS NULL AND is_comp = FALSE
GROUP BY 1, 2;
