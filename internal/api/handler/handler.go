package handler

import (
	"context"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/google/uuid"
)

// Handler defines a HTTP handler struct
type Handler struct {
	ctrl controller.Controller
}

func New(ctrl controller.Controller) *Handler {
	return &Handler{ctrl}
}

const (
	accessTokenDuration  = 24 * time.Hour
	refreshTokenDuration = 720 * time.Hour
)

func validateEntryTimeOwnership(h *Handler, userId, entryTimeId uuid.UUID) (bool, error) {
	//get entry time
	entryTime, err := h.ctrl.GetEntryTime(context.Background(), entryTimeId)
	if err != nil {
		return false, err
	}

	if entryTime.UserID != userId {
		return false, nil
	}
	return true, nil
}
