package api

import (
	"time"

	_ "github.com/JairoRiver/time_keeper/docs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/time/rate"
)

// newAuthRateLimiter returns an in-memory, per-IP rate limiter for the auth
// endpoints (anonymous user creation and login). A burst of 5 with a slow
// refill (5 req/min) stops user-creation spam and password brute-force while
// leaving normal navigation untouched.
func newAuthRateLimiter() echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(5.0 / 60.0), // ~0.083 req/s = 5 per minute
			Burst:     5,
			ExpiresIn: 3 * time.Minute,
		},
	)
	return middleware.RateLimiter(store)
}

func (server *Server) setupRouter() {
	e := echo.New()

	// Serve generated Tailwind CSS and other static assets.
	e.Static("/static", "static")

	// Default security headers (X-Frame-Options, X-Content-Type-Options, etc.).
	e.Use(middleware.Secure())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			server.logger.Info().
				Str("method", v.Method).
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")
			return nil
		},
	}))
	// api/v1 namespace. Auth middleware is applied per-route (not via an
	// empty-prefix subgroup with .Use()) so that unknown /api/v1/* paths return
	// a 404 instead of being swallowed by a group "/*" catch-all.
	api := e.Group("api/v1")
	auth := server.handler.AuthMiddleware
	cookieAuth := server.handler.CookieMiddleware

	// Shared per-IP limiter for auth endpoints (reused for login in TK-11).
	authRateLimiter := newAuthRateLimiter()

	// @title Short Link API
	// @version 1.0
	// @description Testing Swagger APIs.
	// @termsOfService http://swagger.io/terms/

	// @contact.name API Support
	// @contact.url http://www.swagger.io/support
	// @contact.email support@swagger.io

	// @license.name Apache 2.0
	// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

	// @host localhost:8081
	// @BasePath /api/v1/
	// @schemes http

	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization

	// Pages (templ) — public
	e.GET("/", server.handler.LandingPage)
	e.GET("/try", server.handler.Try, authRateLimiter)

	// Pages (templ) — protected (redirect to / if no session).
	// Middleware is applied per-route on purpose: an empty-prefix group with
	// .Use() would register a global "/*" catch-all (Echo behaviour), turning
	// every unknown URL into a 401/redirect instead of a 404.
	pageAuth := server.handler.PageAuthMiddleware
	e.GET("/registro", server.handler.RegistroPage, pageAuth)
	e.POST("/registro/start", server.handler.TimerStart, pageAuth)
	e.POST("/registro/stop", server.handler.TimerStop, pageAuth)
	e.GET("/resumen", server.handler.ResumenPage, pageAuth)

	// Swagger UI is dev-only; disabled in deployment config.
	if server.enableSwagger {
		api.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// Auth (Logto OIDC) — web routes, no api/v1 prefix
	e.GET("/auth/login", server.handler.Login)
	e.GET("/auth/callback", server.handler.Callback)
	e.GET("/auth/logout", server.handler.Logout)
	// Per-route middleware (not an empty-prefix group) to avoid a global "/*"
	// catch-all that would shadow the 404 handler for unknown URLs.
	e.GET("/auth/link", server.handler.LinkAccount, server.handler.CookieMiddleware)

	// User
	api.POST("/user", server.handler.CreateUser, authRateLimiter)
	api.POST("/refresh", server.handler.RefreshToken, cookieAuth)

	//Entry Time routers
	api.POST("/entry-time", server.handler.CreateEntryTime, auth)
	api.PUT("/entry-time", server.handler.UpdateEntryTime, auth)
	api.GET("/entry-time/:id", server.handler.GetEntryTime, auth)
	api.GET("/entries-time", server.handler.ListEntryTime, auth)
	api.DELETE("/entry-time/:id", server.handler.DeleteEntryTime, auth)

	server.router = e
}
