-- name: GetUserByID :one
SELECT u.*, r.code AS role_code, r.name AS role_name
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT u.*, r.code AS role_code
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.email = $1 AND u.deleted_at IS NULL AND u.is_active = TRUE;

-- name: ListBaristasByLocation :many
SELECT u.id, u.full_name, u.pin_hash, u.is_active
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.default_location_id = $1
  AND r.code = 'barista'
  AND u.deleted_at IS NULL
  AND u.is_active = TRUE
ORDER BY u.full_name;

-- name: ListUsersByLocation :many
SELECT u.*, r.code AS role_code, r.name AS role_name
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.default_location_id = $1
  AND u.deleted_at IS NULL
ORDER BY r.code, u.full_name;

-- name: CreateUser :one
INSERT INTO users (default_location_id, role_id, full_name, email, pin_hash, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateUserPIN :exec
UPDATE users SET pin_hash = $2 WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users SET is_active = FALSE WHERE id = $1;

-- name: GetUserPermissions :many
SELECT p.code
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN users u ON u.role_id = rp.role_id
WHERE u.id = $1;

-- name: GetLocation :one
SELECT * FROM locations WHERE id = $1 AND deleted_at IS NULL;

-- name: ListLocations :many
SELECT * FROM locations WHERE deleted_at IS NULL ORDER BY name;

-- name: CreateLocation :one
INSERT INTO locations (name, address, city, timezone, phone)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListRoles :many
SELECT id, code, name FROM roles ORDER BY
  CASE code WHEN 'admin' THEN 1 WHEN 'manager' THEN 2 ELSE 3 END;
