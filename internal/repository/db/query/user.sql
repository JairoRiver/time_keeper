-- name: CreateUser :one
INSERT INTO users (
  email,
  "role"
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByIdentityId :one
SELECT * FROM users
WHERE user_identity_id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  email = COALESCE(sqlc.narg(email), email),
  "role" = COALESCE(sqlc.narg(role), role),
  user_identity_id = COALESCE(sqlc.narg(user_identity_id), user_identity_id),
  email_validated = COALESCE(sqlc.narg(email_validated), email_validated),
  is_active = COALESCE(sqlc.narg(is_active), is_active),
  updated_at = NOW()
WHERE
  id = sqlc.arg(id)
RETURNING *;