package db

import (
	"context"
	"testing"
	"time"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

// Create Time Entry function for test proposes
func createRandomTimeEntry(t *testing.T, userId pgtype.UUID) TimeEntry {
	var newUserId uuid.UUID
	if !userId.Valid {
		user := createRandomUser(t)
		newUserId = user.ID
	} else {
		newUserId = userId.Bytes
	}

	tag := util.RandomString(9)
	timeEntryParams := CreateTimeEntryParams{
		UserID:    newUserId,
		Tag:       tag,
		TimeStart: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	timeEntry, err := testQueries.CreateTimeEntry(context.Background(), timeEntryParams)
	assert.NoError(t, err)
	assert.Equal(t, newUserId, timeEntry.UserID)
	assert.Equal(t, tag, timeEntry.Tag)
	assert.NotEmpty(t, timeEntry.TimeStart)

	return timeEntry
}

func TestCreateTimeEntry(t *testing.T) {
	_ = createRandomTimeEntry(t, pgtype.UUID{})
}

func TestGetTimeEntryById(t *testing.T) {
	entryTime := createRandomTimeEntry(t, pgtype.UUID{})
	getEntryTime, err := testQueries.GetTimeEntryById(context.Background(), entryTime.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, getEntryTime.ID)
	assert.NotEmpty(t, getEntryTime.UserID)
	assert.NotEmpty(t, getEntryTime.Tag)
	assert.NotEmpty(t, getEntryTime.TimeStart)
	assert.NotEmpty(t, getEntryTime.CreatedAt)
}

func TestListTimeEntry(t *testing.T) {
	//Create random entry tyme for test
	user := createRandomUser(t)
	timeEntryLen := 20
	for i := 0; i < timeEntryLen; i++ {
		_ = createRandomTimeEntry(t, pgtype.UUID{Bytes: user.ID, Valid: true})
	}

	listParams := ListTimeEntryParams{
		UserID:      user.ID,
		TimeStart:   pgtype.Timestamp{Time: time.Now().AddDate(0, 0, -1), Valid: true},
		TimeStart_2: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	listEntries, err := testQueries.ListTimeEntry(context.Background(), listParams)
	assert.NoError(t, err)
	assert.Len(t, listEntries, timeEntryLen)
	for _, entryTime := range listEntries {
		assert.NotEmpty(t, entryTime)
		assert.Equal(t, entryTime.UserID, user.ID)
		assert.NotEmpty(t, entryTime.ID)
		assert.NotEmpty(t, entryTime.Tag)
		assert.NotZero(t, entryTime.TimeStart.Time)
		assert.Zero(t, entryTime.TimeEnd)
	}
}

func TestDeleteTimeEntry(t *testing.T) {
	entryTime := createRandomTimeEntry(t, pgtype.UUID{})
	deletedEntryTime, err := testQueries.DeleteTimeEntry(context.Background(), entryTime.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, deletedEntryTime)
	assert.Equal(t, entryTime.ID, deletedEntryTime.ID)
	assert.Equal(t, entryTime.UserID, deletedEntryTime.UserID)
	assert.Equal(t, entryTime.TimeStart, deletedEntryTime.TimeStart)
	assert.Equal(t, entryTime.TimeEnd, deletedEntryTime.TimeEnd)
	assert.Equal(t, entryTime.Tag, deletedEntryTime.Tag)
	assert.Equal(t, entryTime.CreatedAt, deletedEntryTime.CreatedAt)
	assert.Equal(t, entryTime.UpdatedAt, deletedEntryTime.UpdatedAt)
}
