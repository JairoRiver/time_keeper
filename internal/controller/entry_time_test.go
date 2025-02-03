package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Create entry time function for test proposes on controller package
func createRandomEntryTime(t *testing.T, userId uuid.UUID) EntryTimeResponse {
	//testing nil user id
	errorEntryTimeParams := CreateEntryTimeParams{
		UserID: uuid.Nil,
	}
	errorEntryTime, err := testControl.CreateEntryTime(context.Background(), errorEntryTimeParams)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))
	assert.Zero(t, errorEntryTime)

	createEntryTimeParams := CreateEntryTimeParams{
		UserID:    userId,
		Tag:       util.RandomString(5),
		TimeStart: time.Now().UTC(),
	}

	entryTime, err := testControl.CreateEntryTime(context.Background(), createEntryTimeParams)
	assert.NoError(t, err)
	assert.NotZero(t, entryTime)
	assert.NotZero(t, entryTime.ID)
	assert.Equal(t, userId, entryTime.UserID)
	assert.Equal(t, createEntryTimeParams.Tag, entryTime.Tag)
	assert.Equal(t, createEntryTimeParams.TimeStart.Round(time.Second), entryTime.TimeStart.Round(time.Second))
	assert.True(t, entryTime.TimeEnd.IsZero())

	return entryTime
}

func TestCreateEntryTime(t *testing.T) {
	user := createRandomUser(t, "")
	createRandomEntryTime(t, user.UserId)
}

func TestGetEntryTime(t *testing.T) {
	//error entry time Id is zero
	errorEntryTime, err := testControl.GetEntryTime(context.Background(), uuid.Nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))
	assert.Zero(t, errorEntryTime)

	//test get entry time
	user := createRandomUser(t, "")
	entryTime := createRandomEntryTime(t, user.UserId)
	getEntryTime, err := testControl.GetEntryTime(context.Background(), entryTime.ID)
	assert.NoError(t, err)
	assert.NotZero(t, getEntryTime)
	assert.Equal(t, entryTime.ID, getEntryTime.ID)
	assert.Equal(t, entryTime.UserID, getEntryTime.UserID)
	assert.Equal(t, entryTime.Tag, getEntryTime.Tag)
	assert.Equal(t, entryTime.TimeStart, getEntryTime.TimeStart)
	assert.Equal(t, entryTime.TimeEnd, getEntryTime.TimeEnd)
}

func TestListEntryTime(t *testing.T) {
	//test error user id is zero value
	errorEntriesTimeParams := ListEntryTimeParams{UserId: uuid.Nil}
	errorListEntriesTime, err := testControl.ListEntryTime(context.Background(), errorEntriesTimeParams)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))
	assert.Nil(t, errorListEntriesTime)

	//test get entries times
	n := 20
	user := createRandomUser(t, "")
	for i := 0; i < n; i++ {
		createRandomEntryTime(t, user.UserId)
	}
	entriesTimeParams := ListEntryTimeParams{
		UserId:     user.UserId,
		PageNumber: 1,
	}
	listEntriesTime, err := testControl.ListEntryTime(context.Background(), entriesTimeParams)
	assert.NoError(t, err)
	assert.NotNil(t, listEntriesTime)
	assert.Len(t, listEntriesTime, n)
}

func TestUpdateEntryTime(t *testing.T) {
	//check if the entry time id is zero value
	errorUpdateParam := UpdateEntryTimeParams{Id: uuid.Nil}
	errorUpdateEntryTime, err := testControl.UpdateEntryTime(context.Background(), errorUpdateParam)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))
	assert.Zero(t, errorUpdateEntryTime)

	//check update entry time
	user := createRandomUser(t, "")
	entryTime := createRandomEntryTime(t, user.UserId)

	newTag := util.RandomString(6)
	newTimeStart := time.Now().UTC().Add(time.Duration(-1 * util.RandomInt(1, 30) * int64(time.Minute)))
	newTimeEnd := time.Now().UTC().Add(time.Duration(util.RandomInt(1, 30) * int64(time.Minute)))
	updateParams := UpdateEntryTimeParams{
		Id:        entryTime.ID,
		Tag:       newTag,
		TimeStart: newTimeStart,
		TimeEnd:   newTimeEnd,
	}
	updateEntryTime, err := testControl.UpdateEntryTime(context.Background(), updateParams)
	assert.NoError(t, err)
	assert.NotZero(t, updateEntryTime)
	assert.Equal(t, updateParams.Id, updateEntryTime.ID)
	assert.Equal(t, updateParams.Tag, updateEntryTime.Tag)
	assert.Equal(t, updateParams.TimeStart.Round(time.Second), updateEntryTime.TimeStart.Round(time.Second))
	assert.Equal(t, updateParams.TimeEnd.Round(time.Second), updateEntryTime.TimeEnd.Round(time.Second))
}

func TestGetEntryTimeOwner(t *testing.T) {
	//check if the entryTimeId are zero
	erroGetOwner, err := testControl.GetEntryTimeOwner(context.Background(), uuid.Nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))
	assert.Zero(t, erroGetOwner)

	//check get owner
	user := createRandomUser(t, "")
	entryTime := createRandomEntryTime(t, user.UserId)
	getOwner, err := testControl.GetEntryTimeOwner(context.Background(), entryTime.ID)
	assert.NoError(t, err)
	assert.NotZero(t, getOwner)
	assert.Equal(t, entryTime.ID, getOwner.EntryTimeId)
	assert.Equal(t, user.UserId, getOwner.EntryTimeUserId)
}

func TestDeleteEntryTime(t *testing.T) {
	user := createRandomUser(t, "")
	entryTime := createRandomEntryTime(t, user.UserId)
	deleteEntryTime, err := testControl.DeleteEntryTime(context.Background(), entryTime.ID)
	assert.NoError(t, err)
	assert.NotZero(t, deleteEntryTime)
	assert.Equal(t, entryTime.ID, deleteEntryTime.ID)
	assert.Equal(t, entryTime.UserID, deleteEntryTime.UserID)
	assert.Equal(t, entryTime.Tag, deleteEntryTime.Tag)
	assert.Equal(t, entryTime.TimeStart, deleteEntryTime.TimeStart)
	assert.Equal(t, entryTime.TimeEnd, deleteEntryTime.TimeEnd)
}
