-- name: RegisterDevice :one
INSERT INTO device_registry (location_id, user_id, device_name, platform, app_version)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateDeviceHeartbeat :exec
UPDATE device_registry
SET last_seen_at = NOW(), app_version = $2
WHERE id = $1;

-- name: InsertSyncEvent :one
INSERT INTO sync_events (client_uuid, device_id, sequence, event_type, payload, device_ts, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (client_uuid) DO NOTHING
RETURNING *;

-- name: GetSyncEventByUUID :one
SELECT * FROM sync_events WHERE client_uuid = $1;

-- name: ListSyncEventsAfterCursor :many
SELECT * FROM sync_events
WHERE device_id != $1
  AND (server_ts, id) > ($2::timestamptz, $3::uuid)
ORDER BY server_ts, id
LIMIT $4;

-- name: CreateSyncConflict :one
INSERT INTO sync_conflicts (sync_event_id, kind, details, needs_review)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListUnresolvedConflicts :many
SELECT sc.*, se.event_type, se.payload, se.device_ts
FROM sync_conflicts sc
JOIN sync_events se ON se.id = sc.sync_event_id
WHERE sc.needs_review = TRUE
ORDER BY sc.created_at DESC;

-- name: ResolveConflict :one
UPDATE sync_conflicts
SET needs_review = FALSE,
    resolved_by  = $2,
    resolved_at  = NOW(),
    resolution   = $3
WHERE id = $1
RETURNING *;

-- name: InsertAuditLog :exec
INSERT INTO audit_log (user_id, action, entity, entity_id, old_value, new_value, device_id, ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetOrderByClientUUID :one
SELECT * FROM orders WHERE client_uuid = $1;

-- name: GetMenuSnapshot :many
SELECT
    p.*,
    c.name AS category_name,
    c.sort_order AS category_sort_order,
    c.icon AS category_icon
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE p.location_id = $1
  AND p.is_active = TRUE
  AND p.deleted_at IS NULL
  AND c.deleted_at IS NULL
ORDER BY c.sort_order, p.sort_order;
