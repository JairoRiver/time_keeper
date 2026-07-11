package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestAuthRateLimiter verifies that the shared auth limiter lets a burst of 5
// requests through and throttles the next one from the same IP with a 429.
func TestAuthRateLimiter(t *testing.T) {
	e := echo.New()
	e.GET("/try", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, newAuthRateLimiter())

	var lastCode int
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodGet, "/try", nil)
		req.RemoteAddr = "203.0.113.5:1234" // same IP for every request
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if i < 5 {
			assert.Equal(t, http.StatusOK, rec.Code, "request %d should pass the burst", i+1)
		}
		lastCode = rec.Code
	}
	assert.Equal(t, http.StatusTooManyRequests, lastCode, "6th request should be rate limited")
}

// TestAuthRateLimiterPerIP verifies the limit is tracked per IP: a second IP is
// not affected by the first one exhausting its burst.
func TestAuthRateLimiterPerIP(t *testing.T) {
	e := echo.New()
	e.GET("/try", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, newAuthRateLimiter())

	exhaust := func(ip string) int {
		var code int
		for i := 0; i < 6; i++ {
			req := httptest.NewRequest(http.MethodGet, "/try", nil)
			req.RemoteAddr = ip + ":1234"
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			code = rec.Code
		}
		return code
	}

	assert.Equal(t, http.StatusTooManyRequests, exhaust("203.0.113.5"))
	// A different IP still gets its own fresh burst.
	req := httptest.NewRequest(http.MethodGet, "/try", nil)
	req.RemoteAddr = "203.0.113.9:1234"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
