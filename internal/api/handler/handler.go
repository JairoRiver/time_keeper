package handler

import (
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
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
