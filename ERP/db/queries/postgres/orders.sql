-- name: CreateOrder :one
INSERT INTO orders (
    location_id, shift_id, barista_id, customer_id,
    subtotal, discount_total, loyalty_points_used, total, cost_total,
    receipt_no, status, client_uuid, created_by, is_comp, comp_recipient
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9,
    $10, 'open', $11, $3, $12, $13
) RETURNING *;

-- name: PayOrder :one
UPDATE orders
SET status = 'paid', updated_at = NOW()
WHERE id = $1 AND status = 'open'
RETURNING *;

-- name: CancelOrder :one
UPDATE orders
SET status       = 'cancelled',
    cancel_reason = $2,
    cancelled_by  = $3,
    cancelled_at  = NOW()
WHERE id = $1 AND status IN ('open','paid')
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders WHERE id = $1 AND deleted_at IS NULL;

-- name: ListOrdersByShift :many
SELECT o.*, u.full_name AS barista_name
FROM orders o
JOIN users u ON u.id = o.barista_id
WHERE o.shift_id = $1 AND o.deleted_at IS NULL
ORDER BY o.created_at DESC;

-- name: ListOrdersByLocation :many
SELECT o.*, u.full_name AS barista_name
FROM orders o
JOIN users u ON u.id = o.barista_id
WHERE o.location_id = $1
  AND o.created_at BETWEEN $2 AND $3
  AND o.deleted_at IS NULL
ORDER BY o.created_at DESC
LIMIT $4 OFFSET $5;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, qty, unit_price_snapshot, line_total, line_cost, client_uuid)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetOrderItems :many
SELECT oi.*,
       p.name AS product_name
FROM order_items oi
JOIN products p ON p.id = oi.product_id
WHERE oi.order_id = $1;

-- name: CreatePayment :one
INSERT INTO payments (order_id, payment_method_id, amount, external_ref)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateRefund :one
INSERT INTO refunds (order_id, amount, reason, approved_by, client_uuid)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetNextReceiptNo :one
SELECT COALESCE(
    LPAD((MAX(REGEXP_REPLACE(receipt_no, '[^0-9]', '', 'g'))::bigint + 1)::text, 6, '0'),
    '000001'
) AS next_receipt_no
FROM orders
WHERE location_id = $1 AND DATE(created_at) = CURRENT_DATE;

-- name: GetOrderRevenueSummary :one
SELECT
    COUNT(*)          AS orders_count,
    SUM(total)        AS revenue,
    SUM(cost_total)   AS cost,
    SUM(total - cost_total) AS margin,
    AVG(total)        AS avg_check
FROM orders
WHERE location_id = $1
  AND status = 'paid'
  AND created_at BETWEEN $2 AND $3
  AND deleted_at IS NULL;

-- name: CreateOrderItemModifier :one
INSERT INTO order_item_modifiers (order_item_id, modifier_option_id, price_delta_snapshot)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetAllOrderItemModifiers :many
SELECT oim.order_item_id, oim.id, oim.modifier_option_id,
       oim.price_delta_snapshot, mo.name AS option_name
FROM order_item_modifiers oim
JOIN modifier_options mo ON mo.id = oim.modifier_option_id
JOIN order_items oi ON oi.id = oim.order_item_id
WHERE oi.order_id = $1;

-- name: GetOrderPayments :many
SELECT p.*, pm.code AS method_code, pm.name AS method_name
FROM payments p
JOIN payment_methods pm ON pm.id = p.payment_method_id
WHERE p.order_id = $1;

-- name: ListPaymentMethods :many
SELECT * FROM payment_methods WHERE is_active = TRUE ORDER BY sort_order;

-- name: GetPaymentMethodByCode :one
SELECT * FROM payment_methods WHERE code = $1 AND is_active = TRUE;

-- name: GetCustomerByPhone :one
SELECT * FROM customers WHERE phone = $1 AND deleted_at IS NULL;

-- name: CreateCustomer :one
INSERT INTO customers (phone, name) VALUES ($1, $2) RETURNING *;

-- name: CreateLoyaltyAccount :one
INSERT INTO loyalty_accounts (customer_id) VALUES ($1) RETURNING *;

-- name: GetLoyaltyAccountByCustomer :one
SELECT * FROM loyalty_accounts WHERE customer_id = $1;

-- name: DeductLoyaltyPoints :one
UPDATE loyalty_accounts
SET points_balance = points_balance - $2,
    updated_at     = NOW()
WHERE customer_id = $1 AND points_balance >= $2
RETURNING *;

-- name: AddLoyaltyPoints :one
UPDATE loyalty_accounts
SET points_balance = points_balance + $2,
    total_visits   = total_visits + $3,
    updated_at     = NOW()
WHERE customer_id = $1
RETURNING *;

-- name: CreateLoyaltyTransaction :one
INSERT INTO loyalty_transactions (loyalty_account_id, order_id, kind, points_delta, note)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetLoyaltyRules :many
SELECT * FROM loyalty_rules ORDER BY code;

-- name: GetLoyaltyRuleByCode :one
SELECT * FROM loyalty_rules WHERE code = $1;

-- name: SetLoyaltyRule :one
UPDATE loyalty_rules
SET params    = $2,
    is_active = $3,
    updated_at = NOW()
WHERE code = $1
RETURNING *;

-- name: ApplyLoyaltyOrder :one
UPDATE loyalty_accounts
SET points_balance  = $2,
    free_drinks_left = $3,
    coffee_punches   = $4,
    total_visits     = total_visits + 1,
    updated_at       = NOW()
WHERE customer_id = $1
RETURNING *;
