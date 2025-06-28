package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateEntryTime_Success(t *testing.T) {
	e := echo.New()

	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	userID := uuid.New()
	start := time.Now()
	end := start.Add(1 * time.Hour)

	// Expected controller response
	expectedEntry := controller.EntryTimeResponse{
		ID:        uuid.New(),
		UserID:    userID,
		Tag:       "work",
		TimeStart: start,
		TimeEnd:   end,
	}

	// Mock expectation
	mockCtrl.On("CreateEntryTime", mock.Anything, mock.AnythingOfType("controller.CreateEntryTimeParams")).Return(expectedEntry, nil)

	// Build request body
	body, _ := json.Marshal(CreateEntryTimeParams{
		Tag:       "work",
		TimeStart: start,
		TimeEnd:   end,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/entry-time", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	addAuthPayload(c, userID)

	// Execute handler
	if assert.NoError(t, h.CreateEntryTime(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp EntryTimeResponse
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp)) {
			assert.Equal(t, expectedEntry.ID, resp.Id)
			assert.Equal(t, expectedEntry.Tag, resp.Tag)
			assert.WithinDuration(t, expectedEntry.TimeStart, resp.TimeStart, time.Second)
			assert.WithinDuration(t, expectedEntry.TimeEnd, resp.TimeEnd, time.Second)
		}
	}

	mockCtrl.AssertExpectations(t)
}

func TestGetEntryTime_Success(t *testing.T) {
	e := echo.New()

	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	userID := uuid.New()
	entryID := uuid.New()
	start := time.Now()
	end := start.Add(2 * time.Hour)

	expectedEntry := controller.EntryTimeResponse{
		ID:        entryID,
		UserID:    userID,
		Tag:       "gym",
		TimeStart: start,
		TimeEnd:   end,
	}

	// Mock expectations
	mockCtrl.On("GetEntryTime", mock.Anything, entryID).Return(expectedEntry, nil)

	// Build request with path param
	req := httptest.NewRequest(http.MethodGet, "/api/v1/entry-time/"+entryID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/entry-time/:id")
	c.SetParamNames("id")
	c.SetParamValues(entryID.String())

	addAuthPayload(c, userID)

	// Execute handler
	if assert.NoError(t, h.GetEntryTime(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp EntryTimeResponse
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp)) {
			assert.Equal(t, expectedEntry.ID, resp.Id)
			assert.Equal(t, expectedEntry.Tag, resp.Tag)
		}
	}

	mockCtrl.AssertExpectations(t)
}
