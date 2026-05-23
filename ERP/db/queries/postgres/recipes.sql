-- name: GetRecipe :one
SELECT * FROM recipes WHERE id = $1 AND deleted_at IS NULL;

-- name: ListRecipes :many
SELECT * FROM recipes
WHERE location_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: CreateRecipe :one
INSERT INTO recipes (location_id, name, recipe_type, yield_qty, yield_unit, notes, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateRecipe :one
UPDATE recipes
SET name = $2, recipe_type = $3, yield_qty = $4, yield_unit = $5, notes = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteRecipe :exec
UPDATE recipes SET deleted_at = NOW() WHERE id = $1;

-- name: GetRecipeItems :many
SELECT
    ri.id, ri.recipe_id, ri.ingredient_id, ri.sub_recipe_id,
    ri.qty, ri.unit, ri.sort_order,
    i.name  AS ingredient_name,
    i.unit  AS ingredient_unit,
    i.current_avg_cost,
    i.current_qty,
    sr.name AS sub_recipe_name
FROM recipe_items ri
LEFT JOIN ingredients i  ON i.id  = ri.ingredient_id
LEFT JOIN recipes     sr ON sr.id = ri.sub_recipe_id
WHERE ri.recipe_id = $1
ORDER BY ri.sort_order;

-- name: AddRecipeItem :one
INSERT INTO recipe_items (recipe_id, ingredient_id, sub_recipe_id, qty, unit, sort_order)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateRecipeItem :one
UPDATE recipe_items SET qty = $2, unit = $3, sort_order = $4
WHERE id = $1
RETURNING *;

-- name: RemoveRecipeItem :exec
DELETE FROM recipe_items WHERE id = $1;

-- name: LinkProductRecipe :exec
UPDATE products SET recipe_id = $2 WHERE id = $1;

-- name: UnlinkProductRecipe :exec
UPDATE products SET recipe_id = NULL WHERE id = $1;
