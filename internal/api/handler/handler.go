package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/pkg/identity"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Handler holds the controller and identity provider dependencies.
type Handler struct {
	ctrl          controller.Controller
	identity      identity.Provider
	log           zerolog.Logger
	secureCookies bool
}

func New(ctrl controller.Controller, idp identity.Provider, log zerolog.Logger, secureCookies bool) *Handler {
	return &Handler{ctrl: ctrl, identity: idp, log: log, secureCookies: secureCookies}
}

// sessionCookie builds the refresh-token session cookie with the configured
// security attributes. Secure is toggled via config so local HTTP dev works
// while production over HTTPS gets the flag set.
func (h *Handler) sessionCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     util.RefreshTokenName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().UTC().Add(refreshTokenDuration),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
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
