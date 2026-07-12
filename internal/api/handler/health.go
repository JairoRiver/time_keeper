package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Pinger is the subset of the pgx pool used by the health check: a liveness
// ping against the database. Keeping it an interface lets the handler be tested
// without a real database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// healthResponse is the public body of GET /health.
type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// healthPingTimeout bounds the database ping so a stuck connection cannot hang
// the liveness check used by the deploy runbook and reverse proxy.
const healthPingTimeout = 2 * time.Second

// Health reports whether the service can reach its database and which version is
// running. It returns 200 when the ping succeeds and 503 when it fails. The
// route is public (no auth) and excluded from rate limiting so uptime and
// deploy checks stay cheap and always reachable.
func (h *Handler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), healthPingTimeout)
	defer cancel()

	if h.db == nil {
		// Defensive: a Handler wired without a Pinger cannot verify the DB.
		return c.JSON(http.StatusServiceUnavailable, healthResponse{
			Status:  "unavailable",
			Version: h.version,
		})
	}

	if err := h.db.Ping(ctx); err != nil {
		h.log.Warn().Err(err).Msg("health check: database ping failed")
		return c.JSON(http.StatusServiceUnavailable, healthResponse{
			Status:  "unavailable",
			Version: h.version,
		})
	}

	return c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Version: h.version,
	})
}
