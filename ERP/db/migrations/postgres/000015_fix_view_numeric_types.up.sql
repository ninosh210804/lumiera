-- Fix analytics 500s (dashboard + analytics pages).
--
-- orders.total / cost_total are NUMERIC(12,2) and order_items.qty is
-- NUMERIC(12,3), so the aggregates in the analytics materialized views are
-- NUMERIC. The Go layer scans revenue/margin/total_cost/units_sold into int64,
-- and pgx refuses to scan a NUMERIC with a fractional part into int64. As soon
-- as a fractional value appears (recipe costs -> fractional margin/cost,
-- weighed quantities), GetDailyRevenue / GetProductABC / GetSalesHeatmap fail
-- with a 500. (GetBaristaStats survived only because it scans whole-number
-- revenue and a float64 avg_check.)
--
-- Recast the int64-bound aggregates to BIGINT (rounded to whole tenge / units),
-- matching how the app stores and displays money. avg_check, margin_pct and
-- base_price stay NUMERIC (scanned into float64). Materialized views can't be
-- altered in place, so drop and recreate, restoring the unique indexes that
-- REFRESH MATERIALIZED VIEW CONCURRENTLY requires.

DROP MATERIALIZED VIEW IF EXISTS mv_sales_heatmap;
CREATE MATERIALIZED VIEW mv_sales_heatmap AS
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
GROUP BY 1, 2, 3, 4
WITH DATA;

CREATE UNIQUE INDEX ON mv_sales_heatmap (location_id, hour_bucket);

DROP MATERIALIZED VIEW IF EXISTS mv_product_abc;
CREATE MATERIALIZED VIEW mv_product_abc AS
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
GROUP BY 1, 2
WITH DATA;

CREATE UNIQUE INDEX ON mv_product_abc (location_id, product_id);

DROP MATERIALIZED VIEW IF EXISTS mv_daily_revenue;
CREATE MATERIALIZED VIEW mv_daily_revenue AS
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
GROUP BY 1, 2
WITH DATA;

CREATE UNIQUE INDEX ON mv_daily_revenue (location_id, day);
