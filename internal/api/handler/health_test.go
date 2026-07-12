package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// stubPinger is a Pinger whose Ping returns the configured error.
type stubPinger struct {
	err error
}

func (p stubPinger) Ping(ctx context.Context) error { return p.err }

func TestHealth_DatabaseUp(t *testing.T) {
	e := echo.New()
	h := &Handler{db: stubPinger{err: nil}, version: "v1.2.3"}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.Health(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var body healthResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "v1.2.3", body.Version)
}

func TestHealth_DatabaseDown(t *testing.T) {
	e := echo.New()
	h := &Handler{db: stubPinger{err: errors.New("connection refused")}, version: "v1.2.3"}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.Health(c))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body healthResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "unavailable", body.Status)
	assert.Equal(t, "v1.2.3", body.Version)
}

func TestHealth_NoPinger(t *testing.T) {
	e := echo.New()
	h := &Handler{version: "v1.2.3"} // db left nil

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.Health(c))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}
