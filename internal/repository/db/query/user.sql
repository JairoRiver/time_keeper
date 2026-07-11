-- name: CreateUser :one
INSERT INTO users (
  email,
  "role",
  secret_token_key,
  password_hash
) VALUES (
  $1, $2, $3, $4
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
  secret_token_key = COALESCE(sqlc.narg(secret_token_key), secret_token_key),
  password_hash = COALESCE(sqlc.narg(password_hash), password_hash),
  updated_at = NOW()
WHERE
  id = sqlc.arg(id)
RETURNING *;

-- name: GetUserSecretById :one
SELECT id, secret_token_key
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserCredentialsByEmail :one
SELECT id, password_hash, "role", is_active
FROM users
WHERE email IS NOT NULL AND lower(email) = lower(sqlc.arg(email))
LIMIT 1;