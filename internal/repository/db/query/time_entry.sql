-- name: CreateTimeEntry :one
INSERT INTO time_entries (
  user_id,
  tag,
  time_start,
  time_end
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetTimeEntryById :one
SELECT * FROM time_entries
WHERE id = $1 LIMIT 1;

-- name: ListTimeEntry :many
SELECT *
FROM time_entries
WHERE user_id = $1
  AND time_start >= $2
  AND time_start <= $3
ORDER BY DATE_TRUNC('day',time_start) DESC, tag ASC;

-- name: UpdateTimeEntry :one
UPDATE time_entries
SET
  tag = COALESCE(sqlc.narg(tag), tag),
  time_start = COALESCE(sqlc.narg(time_start), time_start),
  time_end = COALESCE(sqlc.narg(time_end), time_end),
  updated_at = NOW()
WHERE
  id = sqlc.arg(id)
RETURNING *;

-- name: DeleteTimeEntry :one
DELETE FROM time_entries 
WHERE id = $1
RETURNING *;