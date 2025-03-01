package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create Entry Time
type CreateEntryTimeParams struct {
	UserId    uuid.UUID `json:"user_id" binding:"required"`
	Tag       string    `json:"tag"`
	TimeStart time.Time `json:"time_start" binding:"required"`
	TimeEnd   time.Time `json:"time_end"`
}
type EntryTimeResponse struct {
	Id        uuid.UUID
	Tag       string
	TimeStart time.Time
	TimeEnd   time.Time
}

func parseEntryTimeResponse(entryTime controller.EntryTimeResponse) EntryTimeResponse {
	return EntryTimeResponse{
		Id:        entryTime.ID,
		Tag:       entryTime.Tag,
		TimeStart: entryTime.TimeStart,
		TimeEnd:   entryTime.TimeEnd,
	}
}

// @Summary Create an entry time
// @Description Create a new time entry for a user
// @ID post-create-entry-time
// @Accept  json
// @Produce  json
// @Param request body CreateEntryTimeParams true "Entry Time Data"
// @Success 201 {object} EntryTimeResponse
// @Router /api/v1/entry-time [post]
func (h *Handler) CreateEntryTime(c echo.Context) error {
	var entryTimeParams CreateEntryTimeParams
	err := c.Bind(&entryTimeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	entryTimeControllerParams := controller.CreateEntryTimeParams{
		UserID:    entryTimeParams.UserId,
		Tag:       entryTimeParams.Tag,
		TimeStart: entryTimeParams.TimeStart,
		TimeEnd:   entryTimeParams.TimeEnd,
	}
	ctx := context.Background()
	entryTime, err := h.ctrl.CreateEntryTime(ctx, entryTimeControllerParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	response := parseEntryTimeResponse(entryTime)
	return c.JSON(http.StatusCreated, response)
}

// Get Entry Time By :id
type GetEntryTimeParam struct {
	Id uuid.UUID `param:"id" binding:"required"`
}

// @Summary Get an entry time
// @Description Retrieve an entry time by its ID
// @ID get-entry-time
// @Produce json
// @Param id path string true "Entry Time ID"
// @Success 200 {object} EntryTimeResponse
// @Router /api/v1/entry-time/{id} [get]
func (h *Handler) GetEntryTime(c echo.Context) error {
	var entryTimeParams GetEntryTimeParam
	err := c.Bind(&entryTimeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	ctx := context.Background()
	entryTime, err := h.ctrl.GetEntryTime(ctx, entryTimeParams.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	response := parseEntryTimeResponse(entryTime)
	return c.JSON(http.StatusOK, response)
}

// list entry Time
type ListEntryTimeParams struct {
	UserId     uuid.UUID `query:"user_id" binding:"required"`
	PageNumber int       `query:"page_number" binding:"required,gte=1"`
}

// @Summary List entry times
// @Description Retrieve a paginated list of entry times for a user
// @ID get-list-entry-time
// @Produce json
// @Param user_id query string true "User ID"
// @Param page_number query int true "Page number (must be >= 1)"
// @Success 200 {array} EntryTimeResponse
// @Router /api/v1/entries-time [get]
func (h *Handler) ListEntryTime(c echo.Context) error {
	var listEntryTimeParams ListEntryTimeParams
	err := c.Bind(&listEntryTimeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	params := controller.ListEntryTimeParams{
		UserId:     listEntryTimeParams.UserId,
		PageNumber: listEntryTimeParams.PageNumber,
	}
	ctx := context.Background()
	listEntries, err := h.ctrl.ListEntryTime(ctx, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	var response []EntryTimeResponse
	for _, entryTime := range listEntries {
		response = append(response, parseEntryTimeResponse(entryTime))
	}

	return c.JSON(http.StatusOK, response)
}

// Update Entry Time
type UpdateEntryTimeParams struct {
	Id        uuid.UUID `json:"id" binding:"required"`
	Tag       string    `json:"tag"`
	TimeStart time.Time `json:"time_start"`
	TimeEnd   time.Time `json:"time_end"`
}

// @Summary Update an entry time
// @Description Update an existing entry time by its ID
// @ID put-update-entry-time
// @Accept json
// @Produce json
// @Param request body UpdateEntryTimeParams true "Entry Time Data"
// @Success 200 {object} EntryTimeResponse
// @Router /api/v1/entry-time [put]
func (h *Handler) UpdateEntryTime(c echo.Context) error {
	var updateEntryTimeParams UpdateEntryTimeParams
	err := c.Bind(&updateEntryTimeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	params := controller.UpdateEntryTimeParams{
		Id:        updateEntryTimeParams.Id,
		Tag:       updateEntryTimeParams.Tag,
		TimeStart: updateEntryTimeParams.TimeStart,
		TimeEnd:   updateEntryTimeParams.TimeEnd,
	}
	ctx := context.Background()
	entryTime, err := h.ctrl.UpdateEntryTime(ctx, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	response := parseEntryTimeResponse(entryTime)
	return c.JSON(http.StatusOK, response)
}

// Delete Entry Time
type DeleteEntryTimeParams struct {
	Id uuid.UUID `param:"id" binding:"required"`
}

// @Summary Delete an entry time
// @Description Delete an entry time by its ID
// @ID delete-entry-time
// @Produce json
// @Param id path string true "Entry Time ID"
// @Success 202 "No Content"
// @Router /api/v1/entry-time/{id} [delete]
func (h *Handler) DeleteEntryTime(c echo.Context) error {
	var deleteEntryTimeParams DeleteEntryTimeParams
	err := c.Bind(&deleteEntryTimeParams)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	ctx := context.Background()
	entryTime, err := h.ctrl.DeleteEntryTime(ctx, deleteEntryTimeParams.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	response := parseEntryTimeResponse(entryTime)
	return c.JSON(http.StatusAccepted, response)
}
