-- name: OpenShift :one
INSERT INTO shifts (location_id, user_id, opening_cash, client_uuid)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CloseShift :one
UPDATE shifts
SET closed_at              = NOW(),
    closing_cash_expected  = $2,
    closing_cash_actual    = $3,
    updated_at             = NOW()
WHERE id = $1 AND closed_at IS NULL
RETURNING *;

-- name: GetActiveShift :one
SELECT * FROM shifts
WHERE user_id = $1 AND location_id = $2 AND closed_at IS NULL
LIMIT 1;

-- name: GetShift :one
SELECT * FROM shifts WHERE id = $1;

-- name: ListShifts :many
SELECT s.*, u.full_name AS user_name
FROM shifts s
JOIN users u ON u.id = s.user_id
WHERE s.location_id = $1
ORDER BY s.opened_at DESC
LIMIT 20;

-- name: GetShiftOrderStats :one
SELECT
    COUNT(*)::bigint          AS orders_count,
    COALESCE(SUM(total), 0)   AS revenue
FROM orders
WHERE shift_id = $1 AND status = 'paid';

-- name: CreateCashDrawerOperation :one
INSERT INTO cash_drawer_operations (shift_id, kind, amount, reason, client_uuid, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListCashDrawerOperations :many
SELECT * FROM cash_drawer_operations
WHERE shift_id = $1
ORDER BY created_at;

-- name: CreateTimeEntry :one
INSERT INTO time_entries (user_id, shift_id, clock_in)
VALUES ($1, $2, NOW())
RETURNING *;

-- name: CloseTimeEntry :one
UPDATE time_entries
SET clock_out = NOW(),
    hours     = EXTRACT(EPOCH FROM (NOW() - clock_in)) / 3600
WHERE id = $1 AND clock_out IS NULL
RETURNING *;

-- name: GetOpenTimeEntry :one
SELECT * FROM time_entries
WHERE user_id = $1 AND shift_id = $2 AND clock_out IS NULL
LIMIT 1;
