package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Handler holds the controller dependency and request-scoped config.
type Handler struct {
	ctrl          controller.Controller
	log           zerolog.Logger
	secureCookies bool
}

func New(ctrl controller.Controller, log zerolog.Logger, secureCookies bool) *Handler {
	return &Handler{ctrl: ctrl, log: log, secureCookies: secureCookies}
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

// errorResponse is the generic JSON body returned to clients on error. Internal
// error detail is logged, never sent to the client.
func errorResponse(msg string) map[string]string {
	return map[string]string{"error": msg}
}

// internalError logs the underlying error (already wrapped with context by the
// controller) and returns a generic 500 response so implementation details never
// reach the client.
func (h *Handler) internalError(c echo.Context, err error) error {
	h.log.Error().Err(err).Msg("request failed")
	return c.JSON(http.StatusInternalServerError, errorResponse("internal server error"))
}

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
