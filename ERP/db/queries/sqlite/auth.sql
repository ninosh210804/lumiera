-- name: GetUserByID :one
SELECT u.*, r.code AS role_code
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = ? AND u.deleted_at IS NULL;

-- name: ListActiveBaristasByLocation :many
SELECT u.id, u.full_name, u.pin_hash
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.default_location_id = ?
  AND r.code = 'barista'
  AND u.is_active = 1
  AND u.deleted_at IS NULL
ORDER BY u.full_name;

-- name: GetUserPermissions :many
SELECT p.code
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN users u ON u.role_id = rp.role_id
WHERE u.id = ?;

-- name: UpsertUser :exec
INSERT INTO users (id, default_location_id, role_id, full_name, email, pin_hash, is_active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    full_name   = excluded.full_name,
    pin_hash    = excluded.pin_hash,
    is_active   = excluded.is_active,
    updated_at  = excluded.updated_at;

-- name: UpsertLocation :exec
INSERT INTO locations (id, name, address, city, timezone, phone, is_active, settings, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    name       = excluded.name,
    address    = excluded.address,
    is_active  = excluded.is_active,
    settings   = excluded.settings,
    updated_at = excluded.updated_at;

-- name: GetActiveShiftForUser :one
SELECT * FROM shifts
WHERE user_id = ? AND location_id = ? AND closed_at IS NULL
LIMIT 1;

-- name: OpenShift :one
INSERT INTO shifts (id, location_id, user_id, opening_cash, client_uuid)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: CloseShift :one
UPDATE shifts
SET closed_at             = datetime('now'),
    closing_cash_expected = ?,
    closing_cash_actual   = ?,
    updated_at            = datetime('now')
WHERE id = ?
RETURNING *;
