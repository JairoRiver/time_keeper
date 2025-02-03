// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: time_entry.sql

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createTimeEntry = `-- name: CreateTimeEntry :one
INSERT INTO time_entries (
  user_id,
  tag,
  time_start
) VALUES (
  $1, $2, $3
) RETURNING id, user_id, tag, time_start, time_end, created_at, updated_at
`

type CreateTimeEntryParams struct {
	UserID    uuid.UUID        `json:"user_id"`
	Tag       string           `json:"tag"`
	TimeStart pgtype.Timestamp `json:"time_start"`
}

func (q *Queries) CreateTimeEntry(ctx context.Context, arg CreateTimeEntryParams) (TimeEntry, error) {
	row := q.db.QueryRow(ctx, createTimeEntry, arg.UserID, arg.Tag, arg.TimeStart)
	var i TimeEntry
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Tag,
		&i.TimeStart,
		&i.TimeEnd,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteTimeEntry = `-- name: DeleteTimeEntry :one
DELETE FROM time_entries 
WHERE id = $1
RETURNING id, user_id, tag, time_start, time_end, created_at, updated_at
`

func (q *Queries) DeleteTimeEntry(ctx context.Context, id uuid.UUID) (TimeEntry, error) {
	row := q.db.QueryRow(ctx, deleteTimeEntry, id)
	var i TimeEntry
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Tag,
		&i.TimeStart,
		&i.TimeEnd,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getTimeEntryById = `-- name: GetTimeEntryById :one
SELECT id, user_id, tag, time_start, time_end, created_at, updated_at FROM time_entries
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetTimeEntryById(ctx context.Context, id uuid.UUID) (TimeEntry, error) {
	row := q.db.QueryRow(ctx, getTimeEntryById, id)
	var i TimeEntry
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Tag,
		&i.TimeStart,
		&i.TimeEnd,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listTimeEntry = `-- name: ListTimeEntry :many
SELECT id, user_id, tag, time_start, time_end, created_at, updated_at
FROM time_entries
WHERE user_id = $1
  AND time_start >= $2
  AND time_start <= $3
ORDER BY DATE_TRUNC('day',time_start) DESC, tag ASC
`

type ListTimeEntryParams struct {
	UserID      uuid.UUID        `json:"user_id"`
	TimeStart   pgtype.Timestamp `json:"time_start"`
	TimeStart_2 pgtype.Timestamp `json:"time_start_2"`
}

func (q *Queries) ListTimeEntry(ctx context.Context, arg ListTimeEntryParams) ([]TimeEntry, error) {
	rows, err := q.db.Query(ctx, listTimeEntry, arg.UserID, arg.TimeStart, arg.TimeStart_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TimeEntry{}
	for rows.Next() {
		var i TimeEntry
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Tag,
			&i.TimeStart,
			&i.TimeEnd,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTimeEntry = `-- name: UpdateTimeEntry :one
UPDATE time_entries
SET
  tag = COALESCE($1, tag),
  time_start = COALESCE($2, time_start),
  time_end = COALESCE($3, time_end),
  updated_at = NOW()
WHERE
  id = $4
RETURNING id, user_id, tag, time_start, time_end, created_at, updated_at
`

type UpdateTimeEntryParams struct {
	Tag       pgtype.Text      `json:"tag"`
	TimeStart pgtype.Timestamp `json:"time_start"`
	TimeEnd   pgtype.Timestamp `json:"time_end"`
	ID        uuid.UUID        `json:"id"`
}

func (q *Queries) UpdateTimeEntry(ctx context.Context, arg UpdateTimeEntryParams) (TimeEntry, error) {
	row := q.db.QueryRow(ctx, updateTimeEntry,
		arg.Tag,
		arg.TimeStart,
		arg.TimeEnd,
		arg.ID,
	)
	var i TimeEntry
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Tag,
		&i.TimeStart,
		&i.TimeEnd,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
