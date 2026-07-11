package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// noopNext is a downstream handler that must NOT be reached when auth fails.
func noopNext(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	e := echo.New()
	h := newTestHandler(new(MockController))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/entries-time", nil)
	req.Header.Set(authorizationHeaderKey, "Bearer not-a-jwt")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.AuthMiddleware(noopNext)(c)

	assert.NoError(t, err)
	// A malformed token is a client error (401), never a 500.
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.JSONEq(t, `{"error":"invalid or expired token"}`, rec.Body.String())
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	e := echo.New()
	h := newTestHandler(new(MockController))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/entries-time", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.AuthMiddleware(noopNext)(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.JSONEq(t, `{"error":"authorization header is not provided"}`, rec.Body.String())
}

func TestCookieMiddleware_MalformedToken(t *testing.T) {
	e := echo.New()
	h := newTestHandler(new(MockController))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	req.AddCookie(&http.Cookie{Name: util.RefreshTokenName, Value: "not-a-jwt"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CookieMiddleware(noopNext)(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.JSONEq(t, `{"error":"invalid or expired token"}`, rec.Body.String())
}
