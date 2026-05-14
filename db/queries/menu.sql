-- name: GetAllMenuItems :many
SELECT 
    m.id,
    m.name,
    m.description,
    m.price,
    m.available,
    m.created_at,
    c.id AS category_id,
    c.name AS category_name
FROM menu_items m
JOIN categories c ON c.id = m.category_id
WHERE m.available = true
ORDER BY c.name, m.name;

-- name: GetMenuItemByID :one
SELECT
    m.id,
    m.name,
    m.description,
    m.price,
    m.available,
    m.created_at,
    c.id   AS category_id,
    c.name AS category_name
FROM menu_items m
JOIN categories c ON c.id = m.category_id
WHERE m.id = $1;

-- name: GetAllCategories :many
SELECT id, name, created_at
FROM categories
ORDER BY name ASC;

-- name: CreateMenuItem :one
INSERT INTO menu_items (
    category_id,
    name,
    description,
    price,
    available
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateMenuItem :one
UPDATE menu_items
SET
    category_id = $2,
    name        = $3,
    description = $4,
    price       = $5,
    available   = $6
WHERE id = $1
RETURNING *;

-- name: DeleteMenuItem :exec
DELETE FROM menu_items
WHERE id = $1;
