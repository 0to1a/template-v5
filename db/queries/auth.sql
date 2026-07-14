-- name: GetActiveUserByEmail :one
SELECT public_uuid, email
FROM users
WHERE lower(email) = lower(sqlc.arg(email)) AND deleted_at IS NULL;
