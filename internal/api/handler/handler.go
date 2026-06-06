package handler

import (
	"context"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/pkg/identity"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Handler holds the controller and identity provider dependencies.
type Handler struct {
	ctrl     controller.Controller
	identity identity.Provider
	log      zerolog.Logger
}

func New(ctrl controller.Controller, idp identity.Provider, log zerolog.Logger) *Handler {
	return &Handler{ctrl: ctrl, identity: idp, log: log}
}

const (
	accessTokenDuration  = 24 * time.Hour
	refreshTokenDuration = 720 * time.Hour
)

func validateEntryTimeOwnership(h *Handler, userId, entryTimeId uuid.UUID) (bool, error) {
	entryTime, err := h.ctrl.GetEntryTime(context.Background(), entryTimeId)
	if err != nil {
		return false, err
	}
	if entryTime.UserID != userId {
		return false, nil
	}
	return true, nil
}
