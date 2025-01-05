package controller

import (
	"context"
	"fmt"
	"time"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/google/uuid"
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
	entryTimeResponse := EntryTimeResponse{
		ID:        entryTime.ID,
		UserID:    entryTime.UserID,
		Tag:       entryTime.Tag,
		TimeStart: entryTime.TimeStart.Time,
		TimeEnd:   entryTime.TimeEnd.Time,
	}
	return entryTimeResponse
}

// create entry time controll method
type CreateEntryTimeParams struct {
	UserID    uuid.UUID
	Tag       string
	TimeStart time.Time
}

func (c *Control) CreateEntryTime(ctx context.Context, params CreateEntryTimeParams) (EntryTimeResponse, error) {
	//check if the UserId are nor zero
	if params.UserID == uuid.Nil {
		return EntryTimeResponse{}, fmt.Errorf("control CreateEntryTime UserId are empty, error: %w", ErrEmptyId)
	}

	entryTimeParam := db.CreateTimeEntryParams{
		UserID:    params.UserID,
		Tag:       params.Tag,
		TimeStart: pgtype.Timestamp{Time: params.TimeStart, Valid: true},
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
type ListEntryTimeParams struct {
	UserId    uuid.UUID
	DateStart time.Time
	DateEnd   time.Time
}

func (c *Control) ListEntryTime(ctx context.Context, params ListEntryTimeParams) ([]EntryTimeResponse, error) {
	//check if the UserId are empty
	if params.UserId == uuid.Nil {
		return nil, fmt.Errorf("control ListEntryTime UserId are empty, error: %w", ErrEmptyId)
	}

	listParams := db.ListTimeEntryParams{
		UserID:      params.UserId,
		TimeStart:   pgtype.Timestamp{Time: params.DateStart, Valid: true},
		TimeStart_2: pgtype.Timestamp{Time: params.DateEnd, Valid: true},
	}
	timeEntries, err := c.repo.ListTimeEntry(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("control ListEntryTime repo ListTimeEntry error: %w", err)
	}

	timeEntriesResponse := make([]EntryTimeResponse, len(timeEntries))
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
		updateParams.TimeStart = pgtype.Timestamp{Valid: false}
	} else {
		updateParams.TimeStart = pgtype.Timestamp{Time: params.TimeStart, Valid: true}
	}

	// chechk if the TimeEnd are empty
	if params.TimeEnd.IsZero() {
		updateParams.TimeEnd = pgtype.Timestamp{Valid: false}
	} else {
		updateParams.TimeEnd = pgtype.Timestamp{Time: params.TimeEnd, Valid: true}
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
