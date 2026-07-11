package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EntryTimeResponse struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Tag       string
	TimeStart time.Time
	TimeEnd   time.Time
}

func formatEntryTimeResponse(entryTime db.TimeEntry) EntryTimeResponse {
	// timestamptz values come back from pgx in the server's local zone; normalize
	// to UTC so the whole app has a single, consistent time contract.
	entryTimeResponse := EntryTimeResponse{
		ID:        entryTime.ID,
		UserID:    entryTime.UserID,
		Tag:       entryTime.Tag,
		TimeStart: entryTime.TimeStart.Time.UTC(),
		TimeEnd:   entryTime.TimeEnd.Time.UTC(),
	}
	return entryTimeResponse
}

// create entry time controll method
type CreateEntryTimeParams struct {
	UserID    uuid.UUID
	Tag       string
	TimeStart time.Time
	TimeEnd   time.Time
}

func (c *Control) CreateEntryTime(ctx context.Context, params CreateEntryTimeParams) (EntryTimeResponse, error) {
	//check if the UserId are nor zero
	if params.UserID == uuid.Nil {
		return EntryTimeResponse{}, fmt.Errorf("control CreateEntryTime UserId are empty, error: %w", ErrEmptyId)
	}

	//check if TimeEnd is are a zero value
	var timeEnd pgtype.Timestamptz
	if !params.TimeEnd.IsZero() {
		timeEnd = pgtype.Timestamptz{Time: params.TimeEnd, Valid: true}
	}

	entryTimeParam := db.CreateTimeEntryParams{
		UserID:    params.UserID,
		Tag:       params.Tag,
		TimeStart: pgtype.Timestamptz{Time: params.TimeStart, Valid: true},
		TimeEnd:   timeEnd,
	}
	entryTime, err := c.repo.CreateTimeEntry(ctx, entryTimeParam)
	if err != nil {
		return EntryTimeResponse{}, fmt.Errorf("control CreateEntryTime repo CreateTimeEntry error: %w", err)
	}

	entryTimeResponse := formatEntryTimeResponse(entryTime)
	return entryTimeResponse, nil
}

// get entry time controll method
func (c *Control) GetEntryTime(ctx context.Context, id uuid.UUID) (EntryTimeResponse, error) {
	//check if the id are zero
	if id == uuid.Nil {
		return EntryTimeResponse{}, fmt.Errorf("control GetEntryTime Id are empty, error: %w", ErrEmptyId)
	}

	entryTime, err := c.repo.GetTimeEntryById(ctx, id)
	if err != nil {
		return EntryTimeResponse{}, fmt.Errorf("control GetEntryTime repo GetTimeEntryById error: %w", err)
	}

	entryTimeResponse := formatEntryTimeResponse(entryTime)
	return entryTimeResponse, nil
}

// list entry time control method
// Page 1 have the current week days, page 2 have the las week days etc
type ListEntryTimeParams struct {
	UserId     uuid.UUID
	PageNumber int
}

func (c *Control) ListEntryTime(ctx context.Context, params ListEntryTimeParams) ([]EntryTimeResponse, error) {
	//check if the UserId are empty
	if params.UserId == uuid.Nil {
		return nil, fmt.Errorf("control ListEntryTime UserId are empty, error: %w", ErrEmptyId)
	}

	//check the page number if the page number is 0 change to 1 the defauld value
	var pageNumber int
	if params.PageNumber == 0 {
		pageNumber = 1
	} else {
		pageNumber = params.PageNumber
	}
	dateStart, dateEnd := getWeekRange(pageNumber)

	listParams := db.ListTimeEntryParams{
		UserID:      params.UserId,
		TimeStart:   pgtype.Timestamptz{Time: dateStart, Valid: true},
		TimeStart_2: pgtype.Timestamptz{Time: dateEnd, Valid: true},
	}
	timeEntries, err := c.repo.ListTimeEntry(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("control ListEntryTime repo ListTimeEntry error: %w", err)
	}

	timeEntriesResponse := make([]EntryTimeResponse, 0, len(timeEntries))
	for _, entryTime := range timeEntries {
		timeEntriesResponse = append(timeEntriesResponse, formatEntryTimeResponse(entryTime))
	}

	return timeEntriesResponse, nil
}

// update entry time control method
type UpdateEntryTimeParams struct {
	Id        uuid.UUID
	Tag       string
	TimeStart time.Time
	TimeEnd   time.Time
}

func (c *Control) UpdateEntryTime(ctx context.Context, params UpdateEntryTimeParams) (EntryTimeResponse, error) {
	// check if the Id are zero value
	if params.Id == uuid.Nil {
		return EntryTimeResponse{}, fmt.Errorf("control UpdateEntryTime Id are empty, error: %w", ErrEmptyId)
	}

	updateParams := db.UpdateTimeEntryParams{
		ID: params.Id,
	}
	// check if the Tag are empty
	if len(params.Tag) == 0 {
		updateParams.Tag = pgtype.Text{Valid: false}
	} else {
		updateParams.Tag = pgtype.Text{String: params.Tag, Valid: true}
	}

	// chechk if the TimeStart are empty
	if params.TimeStart.IsZero() {
		updateParams.TimeStart = pgtype.Timestamptz{Valid: false}
	} else {
		updateParams.TimeStart = pgtype.Timestamptz{Time: params.TimeStart, Valid: true}
	}

	// chechk if the TimeEnd are empty
	if params.TimeEnd.IsZero() {
		updateParams.TimeEnd = pgtype.Timestamptz{Valid: false}
	} else {
		updateParams.TimeEnd = pgtype.Timestamptz{Time: params.TimeEnd, Valid: true}
	}

	updateEntryTime, err := c.repo.UpdateTimeEntry(ctx, updateParams)
	if err != nil {
		return EntryTimeResponse{}, fmt.Errorf("control UpdateEntryTime repo UpdateTimeEntry error: %w", err)
	}

	entryTimeResponse := formatEntryTimeResponse(updateEntryTime)
	return entryTimeResponse, nil
}

// get entry time owner method
type EntryTimeOwnerResponse struct {
	EntryTimeId     uuid.UUID
	EntryTimeUserId uuid.UUID
}

func (c *Control) GetEntryTimeOwner(ctx context.Context, entryTimeId uuid.UUID) (EntryTimeOwnerResponse, error) {
	//check if the entryTimeId are zero
	if entryTimeId == uuid.Nil {
		return EntryTimeOwnerResponse{}, fmt.Errorf("control GetEntryTimeOwner entryTimeId are empty, error: %w", ErrEmptyId)
	}

	entryTime, err := c.repo.GetTimeEntryById(ctx, entryTimeId)
	if err != nil {
		return EntryTimeOwnerResponse{}, fmt.Errorf("control GetEntryTimeOwner repo GetTimeEntryById error: %w", err)
	}

	entryTimeOwnerResponse := EntryTimeOwnerResponse{EntryTimeId: entryTime.ID, EntryTimeUserId: entryTime.UserID}
	return entryTimeOwnerResponse, nil
}

// GetActiveTimer returns the open entry (no TimeEnd) for a user, if any.
func (c *Control) GetActiveTimer(ctx context.Context, userId uuid.UUID) (EntryTimeResponse, bool, error) {
	if userId == uuid.Nil {
		return EntryTimeResponse{}, false, fmt.Errorf("control GetActiveTimer userId is empty, error: %w", ErrEmptyId)
	}
	entry, err := c.repo.GetActiveTimerByUser(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EntryTimeResponse{}, false, nil
		}
		return EntryTimeResponse{}, false, fmt.Errorf("control GetActiveTimer repo error: %w", err)
	}
	return formatEntryTimeResponse(entry), true, nil
}

// StopTimer finds the active entry for a user and sets its TimeEnd to now.
func (c *Control) StopTimer(ctx context.Context, userId uuid.UUID) (EntryTimeResponse, error) {
	active, isActive, err := c.GetActiveTimer(ctx, userId)
	if err != nil {
		return EntryTimeResponse{}, err
	}
	if !isActive {
		return EntryTimeResponse{}, nil
	}
	return c.UpdateEntryTime(ctx, UpdateEntryTimeParams{
		Id:      active.ID,
		TimeEnd: time.Now().UTC(),
	})
}

// ListEntryTimeByDateRangeParams holds parameters for a custom date-range listing.
type ListEntryTimeByDateRangeParams struct {
	UserId    uuid.UUID
	DateStart time.Time
	DateEnd   time.Time
}

// ListEntryTimeByDateRange returns all entries for a user within an explicit date range.
func (c *Control) ListEntryTimeByDateRange(ctx context.Context, params ListEntryTimeByDateRangeParams) ([]EntryTimeResponse, error) {
	if params.UserId == uuid.Nil {
		return nil, fmt.Errorf("control ListEntryTimeByDateRange userId is empty, error: %w", ErrEmptyId)
	}
	listParams := db.ListTimeEntryParams{
		UserID:      params.UserId,
		TimeStart:   pgtype.Timestamptz{Time: params.DateStart, Valid: true},
		TimeStart_2: pgtype.Timestamptz{Time: params.DateEnd, Valid: true},
	}
	entries, err := c.repo.ListTimeEntry(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("control ListEntryTimeByDateRange repo error: %w", err)
	}
	out := make([]EntryTimeResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, formatEntryTimeResponse(e))
	}
	return out, nil
}

// delete entry time method
func (c *Control) DeleteEntryTime(ctx context.Context, id uuid.UUID) (EntryTimeResponse, error) {
	entryTime, err := c.repo.DeleteTimeEntry(ctx, id)
	if err != nil {
		return EntryTimeResponse{}, fmt.Errorf("control DeleteEntryTime repo DeleteTimeEntry error: %w", err)
	}

	entryTimeOwnerResponse := formatEntryTimeResponse(entryTime)
	return entryTimeOwnerResponse, nil
}
