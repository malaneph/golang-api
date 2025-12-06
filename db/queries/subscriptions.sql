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
SET 
    user_id = CASE WHEN $2::text IS NOT NULL THEN $2 ELSE user_id END,
    service_name = CASE WHEN $3::text IS NOT NULL THEN $3 ELSE service_name END,
    price = CASE WHEN $4::integer IS NOT NULL THEN $4 ELSE price END,
    start_date = CASE WHEN $5::timestamp IS NOT NULL THEN $5 ELSE start_date END,
    end_date = CASE WHEN $6::timestamp IS NOT NULL THEN $6 ELSE end_date END
WHERE id = $1;

-- name: DeleteSubscription :exec
-- http: DELETE /subscriptions/:id
DELETE FROM subscriptions WHERE id = $1;


-- name: GetTotalValueSubscription :one
-- http: GET /subscriptions/total-value
SELECT COALESCE(SUM(price), 0) FROM subscriptions
WHERE (user_id = $1 OR NOT $5::bool) AND (service_name = $2 OR NOT $6::bool) AND start_date BETWEEN $3 AND $4;


-- name: GetEarliestStartDateForUser :one
SELECT MIN(start_date) FROM subscriptions WHERE user_id = $1;