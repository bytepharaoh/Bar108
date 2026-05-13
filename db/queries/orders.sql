-- ORDERS

-- name: CreateOrder :one
-- Creates the main order row.
-- Returns the full order so we can use its ID immediately.
INSERT INTO orders (
    user_id,
    promotion_id,
    status,
    total_price,
    discount_amount,
    final_price,
    notes,
    delivery_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetOrderByID :one
-- Fetches a single order with its user info joined.
-- We join users so the handler can return customer name
-- without a second query.
SELECT
    o.*,
    u.name  AS customer_name,
    u.phone AS customer_phone
FROM orders o
JOIN users u ON u.id = o.user_id
WHERE o.id = $1;

-- name: GetOrdersByUserID :many
-- All orders for a specific user, newest first.
SELECT
    o.*,
    u.name  AS customer_name,
    u.phone AS customer_phone
FROM orders o
JOIN users u ON u.id = o.user_id
WHERE o.user_id = $1
ORDER BY o.created_at DESC;

-- name: GetAllOrders :many
-- Admin endpoint — all orders, newest first.
SELECT
    o.*,
    u.name  AS customer_name,
    u.phone AS customer_phone
FROM orders o
JOIN users u ON u.id = o.user_id
ORDER BY o.created_at DESC;

-- name: GetPendingOrders :many
-- Admin dashboard — only unprocessed orders.
SELECT
    o.*,
    u.name  AS customer_name,
    u.phone AS customer_phone
FROM orders o
JOIN users u ON u.id = o.user_id
WHERE o.status = 'pending'
ORDER BY o.created_at ASC;

-- name: UpdateOrderStatus :one
-- Changes the order status.
-- updated_at is refreshed so we know when the last change was.
UPDATE orders
SET
    status     = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AssignCourier :one
-- Assigns a courier to an order.
-- This also updates status to 'out_for_delivery' automatically
-- because you only assign a courier when the order is ready to go.
UPDATE orders
SET
    courier_id = $2,
    status     = 'out_for_delivery',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelOrder :one
UPDATE orders
SET
    status     = 'cancelled',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- ORDER ITEMS

-- name: CreateOrderItem :one
-- Inserts one item line into an order.
-- unit_price is passed explicitly — it's a snapshot
-- of the price at the time of ordering, NOT a live reference.
-- This is critical: menu prices can change, but order history must not.
INSERT INTO order_items (
    order_id,
    menu_item_id,
    quantity,
    unit_price
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetOrderItems :many
-- All items for a given order, with menu item name joined.
-- We join menu_items so we don't need a second query to get names.
SELECT
    oi.*,
    mi.name AS item_name
FROM order_items oi
JOIN menu_items mi ON mi.id = oi.menu_item_id
WHERE oi.order_id = $1;

-- ORDER STATUS HISTORY

-- name: CreateOrderStatusHistory :one
-- Called every time an order status changes.
-- Builds the full timeline a customer sees when tracking their order.
INSERT INTO order_status_history (
    order_id,
    status,
    note
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetOrderStatusHistory :many
-- Full status timeline for an order, oldest first.
-- This is what the customer sees: "Confirmed at 13:02, Preparing at 13:15..."
SELECT *
FROM order_status_history
WHERE order_id = $1
ORDER BY changed_at ASC;

-- PROMOTIONS

-- name: GetPromotionByCode :one
-- Looks up a promo code. Called before placing an order
-- to validate and calculate the discount.
SELECT *
FROM promotions
WHERE code = $1;

-- name: IncrementPromotionUsage :one
-- Called after an order is placed with a promo code.
-- Increments the used_count so we can track and enforce usage_limit.
UPDATE promotions
SET used_count = used_count + 1
WHERE id = $1
RETURNING *;

-- COURIERS

-- name: GetAllCouriers :many
SELECT * FROM couriers ORDER BY name;

-- name: GetCourierByID :one
SELECT * FROM couriers WHERE id = $1;

-- name: GetAvailableCouriers :many
-- Only couriers with status = 'available' can be assigned.
SELECT * FROM couriers
WHERE status = 'available'
ORDER BY name;

-- name: UpdateCourierStatus :one
UPDATE couriers
SET status = $2
WHERE id = $1
RETURNING *;