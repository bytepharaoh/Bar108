-- name: CreateUser :one
INSERT INTO users (
    name,
    phone,
    email,
    password_hash,
    bonus_points
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetAllUsers :many
SELECT *
FROM users
ORDER BY created_at DESC;

-- name: GetActiveUsers :many
SELECT *
FROM users
WHERE is_active = true
ORDER BY created_at DESC;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserByPhone :one
SELECT *
FROM users
WHERE phone = $1;

-- name: UpdateUser :one
UPDATE users
SET
    name = $2,
    phone = $3,
    email = $4,
    password_hash = $5,
    bonus_points = $6,
    is_active = $7
WHERE id = $1
RETURNING *;

-- name: UpdateUserBonusPoints :one
UPDATE users
SET
    bonus_points = $2
WHERE id = $1
RETURNING *;

-- name: HasActiveOrdersByUserID :one
SELECT EXISTS (
    SELECT 1
    FROM orders
    WHERE user_id = $1
      AND status IN ('pending', 'confirmed', 'preparing', 'ready', 'out_for_delivery')
) AS has_active_orders;

-- name: DeactivateUser :one
UPDATE users
SET is_active = false
WHERE id = $1
RETURNING *;

-- name: ActivateUser :one
UPDATE users
SET is_active = true
WHERE id = $1
RETURNING *;