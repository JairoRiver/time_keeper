// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CreateTimeEntry(ctx context.Context, arg CreateTimeEntryParams) (TimeEntry, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetTimeEntryById(ctx context.Context, id uuid.UUID) (TimeEntry, error)
	GetUserByEmail(ctx context.Context, email pgtype.Text) (User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByIdentityId(ctx context.Context, userIdentityID pgtype.UUID) (User, error)
	GetUserSecretById(ctx context.Context, id uuid.UUID) (GetUserSecretByIdRow, error)
	ListTimeEntry(ctx context.Context, arg ListTimeEntryParams) ([]TimeEntry, error)
	UpdateTimeEntry(ctx context.Context, arg UpdateTimeEntryParams) (TimeEntry, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}

var _ Querier = (*Queries)(nil)
