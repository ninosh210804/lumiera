-- name: GetProduct :one
SELECT p.*, c.name AS category_name, r.id AS recipe_id
FROM products p
JOIN categories c ON c.id = p.category_id
LEFT JOIN recipes r ON r.id = p.recipe_id
WHERE p.id = $1 AND p.deleted_at IS NULL;

-- name: ListProducts :many
SELECT p.*, c.name AS category_name
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE p.location_id = $1
  AND p.deleted_at IS NULL
ORDER BY c.sort_order, p.sort_order, p.name;

-- name: ListActiveProducts :many
SELECT p.*, c.name AS category_name
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE p.location_id = $1
  AND p.deleted_at IS NULL
  AND p.is_active = TRUE
ORDER BY c.sort_order, p.sort_order, p.name;

-- name: ListActiveMenuProducts :many
SELECT
    p.*,
    c.name AS category_name,
    c.sort_order AS category_sort_order
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE p.location_id = $1
  AND p.is_active = TRUE
  AND p.is_stop_listed = FALSE
  AND p.deleted_at IS NULL
  AND c.is_active = TRUE
  AND c.deleted_at IS NULL
ORDER BY c.sort_order, p.sort_order;

-- name: CreateProduct :one
INSERT INTO products (
    location_id, category_id, name, description, sku,
    base_price, is_active, image_url, sort_order, created_by
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name        = $2,
    description = $3,
    category_id = $4,
    base_price  = $5,
    is_active   = $6,
    image_url   = $7,
    sort_order  = $8,
    sku         = $9
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SetProductStopList :exec
UPDATE products SET is_stop_listed = $2 WHERE id = $1;

-- name: SoftDeleteProduct :exec
UPDATE products SET deleted_at = NOW() WHERE id = $1;

-- name: ListCategories :many
SELECT * FROM categories
WHERE location_id = $1 AND deleted_at IS NULL
ORDER BY sort_order, name;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateCategory :one
INSERT INTO categories (location_id, name, icon, sort_order, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, icon = $3, sort_order = $4, is_active = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :exec
UPDATE categories SET deleted_at = NOW() WHERE id = $1;

-- name: GetModifierGroups :many
SELECT
    g.*,
    json_agg(
        json_build_object(
            'id',          o.id,
            'name',        o.name,
            'price_delta', o.price_delta,
            'is_active',   o.is_active,
            'sort_order',  o.sort_order,
            'linked_ingredient_id', o.linked_ingredient_id,
            'ingredient_qty_delta', o.ingredient_qty_delta
        ) ORDER BY o.sort_order
    ) AS options
FROM product_modifier_groups g
JOIN modifier_options o ON o.group_id = g.id AND o.is_active = TRUE
WHERE g.product_id = $1
GROUP BY g.id
ORDER BY g.sort_order;

-- name: CreateModifierGroup :one
INSERT INTO product_modifier_groups (product_id, name, selection_type, required, min_select, max_select, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeleteModifierGroup :exec
DELETE FROM product_modifier_groups WHERE id = $1;

-- name: CreateModifierOption :one
INSERT INTO modifier_options (group_id, name, price_delta, linked_ingredient_id, ingredient_qty_delta, sort_order)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateModifierOption :one
UPDATE modifier_options
SET name = $2, price_delta = $3, is_active = $4, sort_order = $5
WHERE id = $1
RETURNING *;

-- name: GetPriceHistory :many
SELECT * FROM price_history
WHERE product_id = $1
ORDER BY effective_from DESC
LIMIT $2;

-- name: GetModifierOption :one
SELECT id, group_id, name, price_delta, is_active
FROM modifier_options
WHERE id = $1;
