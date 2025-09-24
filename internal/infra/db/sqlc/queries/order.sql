-- name: CreateOrder :exec
INSERT INTO orders (id, user_id, amount_jpy, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetOrder :one
SELECT id, amount_jpy, status, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderForUser :one
SELECT id, user_id, amount_jpy, status, created_at, updated_at
FROM orders
WHERE id = $1 AND user_id = $2;

-- name: UpdateOrder :exec
UPDATE orders
SET amount_jpy = $2, status = $3, updated_at = $4
WHERE id = $1;

-- name: UpdateOrderStatusIfPending :execrows
UPDATE orders
SET status = $2, updated_at = $3
WHERE id = $1 AND status = 'PENDING';

-- name: UpdateOrderStatusIfPendingForUser :execrows
UPDATE orders
SET status = $3, updated_at = $4
WHERE id = $1 AND user_id = $2 AND status = 'PENDING';
