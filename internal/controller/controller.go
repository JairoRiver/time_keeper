package controller

import (
	"context"
	"errors"
	"time"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/google/uuid"
)

// Control defines a Entry Time service controller.
type Control struct {
	repo db.Querier
}

// New creates a short link service controller.
func New(repo db.Querier) *Control {
	return &Control{repo}
}

var ErrInvalidRoleValue = errors.New("error invalid role value")
var ErrInvalidIdType = errors.New("error id must be an UUID")
var ErrInvalidEmailType = errors.New("error email must be a string")
var ErrInvalidGetParamType = errors.New("error get param type are invalid")
var ErrEmptyId = errors.New("error id are empty")
var ErrEmptyEmail = errors.New("error email are empty")

type Controller interface {
	CreateEntryTime(ctx context.Context, params CreateEntryTimeParams) (EntryTimeResponse, error)
	CreateUser(ctx context.Context, params CreateUserParam) (UserResponse, error)
	DeleteEntryTime(ctx context.Context, id uuid.UUID) (EntryTimeResponse, error)
	GetEntryTime(ctx context.Context, id uuid.UUID) (EntryTimeResponse, error)
	GetEntryTimeOwner(ctx context.Context, entryTimeId uuid.UUID) (EntryTimeOwnerResponse, error)
	GetUser(ctx context.Context, params GetUserParams) (UserResponse, error)
	GetUserSecretKey(ctx context.Context, userId uuid.UUID) (UserKeyResponse, error)
	ListEntryTime(ctx context.Context, params ListEntryTimeParams) ([]EntryTimeResponse, error)
	UpdateEntryTime(ctx context.Context, params UpdateEntryTimeParams) (EntryTimeResponse, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (UserResponse, error)
}

var _ Controller = (*Control)(nil)

func getWeekRange(weekOffset int) (time.Time, time.Time) {
	today := time.Now()
	weekDay := int(today.Weekday())
	if weekDay == 0 {
		weekDay = 7 //Sunday day is 7 not zero
	}

	offsetDays := (weekOffset - 1) * 7
	// First day of the week (monday)
	auxStartOfWeek := today.AddDate(0, 0, 1-weekDay-offsetDays)
	startOfWeek := time.Date(auxStartOfWeek.Year(), auxStartOfWeek.Month(), auxStartOfWeek.Day(), 0, 0, 0, 0, time.UTC)
	// Last day of the week (Sunday)
	auxEndOfWeek := startOfWeek.AddDate(0, 0, 6)
	endOfWeek := time.Date(auxEndOfWeek.Year(), auxEndOfWeek.Month(), auxEndOfWeek.Day(), 23, 59, 59, int(time.Nanosecond*999999999), time.UTC)
	return startOfWeek, endOfWeek
}
