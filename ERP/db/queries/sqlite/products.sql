-- name: GetProduct :one
SELECT p.*, c.name AS category_name
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE p.id = ? AND p.deleted_at IS NULL;

-- name: ListActiveMenuProducts :many
SELECT
    p.*,
    c.name AS category_name,
    c.sort_order AS category_sort_order,
    c.icon AS category_icon
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE p.location_id = ?
  AND p.is_active = 1
  AND p.is_stop_listed = 0
  AND p.deleted_at IS NULL
  AND c.is_active = 1
  AND c.deleted_at IS NULL
ORDER BY c.sort_order, p.sort_order;

-- name: GetModifierGroups :many
SELECT * FROM product_modifier_groups WHERE product_id = ? ORDER BY sort_order;

-- name: GetModifierOptions :many
SELECT * FROM modifier_options
WHERE group_id = ? AND is_active = 1
ORDER BY sort_order;

-- name: SetProductStopList :exec
UPDATE products SET is_stop_listed = ?, updated_at = datetime('now') WHERE id = ?;

-- name: UpsertProduct :exec
INSERT INTO products (
    id, location_id, category_id, recipe_id, name, description, sku,
    base_price, is_active, is_stop_listed, image_url, sort_order,
    created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    name           = excluded.name,
    description    = excluded.description,
    base_price     = excluded.base_price,
    is_active      = excluded.is_active,
    is_stop_listed = excluded.is_stop_listed,
    image_url      = excluded.image_url,
    sort_order     = excluded.sort_order,
    recipe_id      = excluded.recipe_id,
    updated_at     = excluded.updated_at;

-- name: GetIngredient :one
SELECT * FROM ingredients WHERE id = ? AND deleted_at IS NULL;

-- name: ListIngredients :many
SELECT * FROM ingredients
WHERE location_id = ? AND deleted_at IS NULL ORDER BY name;

-- name: UpdateIngredientBalance :exec
UPDATE ingredients
SET current_qty      = current_qty + ?,
    current_avg_cost = CASE
        WHEN ? > 0 THEN
            (current_avg_cost * current_qty + ? * ?) / NULLIF(current_qty + ?, 0)
        ELSE current_avg_cost
    END,
    updated_at = datetime('now')
WHERE id = ?;

-- name: GetRecipeItems :many
SELECT ri.*, i.name AS ingredient_name, i.unit, i.current_avg_cost, i.current_qty
FROM recipe_items ri
LEFT JOIN ingredients i ON i.id = ri.ingredient_id
WHERE ri.recipe_id = ?;
