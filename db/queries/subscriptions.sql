-- subscriptions.sql

-- name: GetSubscription :one
-- http: GET /subscriptions/:id
SELECT * FROM subscriptions WHERE id = $1;

-- name: ListSubscriptions :many
-- http: GET /subscriptions
SELECT * FROM subscriptions;

-- name: CreateSubscription :one
-- http: POST /subscriptions
INSERT INTO subscriptions (
    user_id, service_name, price, start_date, end_date
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateSubscription :exec
-- http: PATCH /subscriptions/:id
UPDATE subscriptions
SET user_id = $2, service_name = $3, price = $4, start_date = $5, end_date = $6
WHERE id = $1;

-- name: DeleteSubscription :exec
-- http: DELETE /subscriptions/:id
DELETE FROM subscriptions WHERE id = $1;


-- name: GetTotalValueSubscription :one
-- http: GET /subscriptions/total-value
SELECT SUM(price) FROM subscriptions
WHERE (user_id = $1 OR NOT @filter_by_user_id::bool) AND (service_name = $2 OR NOT @filter_by_service_name::bool) AND start_date BETWEEN $3 AND $4
GROUP BY service_name, user_id;
