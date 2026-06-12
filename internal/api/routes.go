package api

import (
	_ "github.com/JairoRiver/time_keeper/docs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (server *Server) setupRouter() {
	e := echo.New()

	// Serve generated Tailwind CSS and other static assets.
	e.Static("/static", "static")

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
	public := e.Group("api/v1")
	private := public.Group("")
	private.Use(server.handler.AuthMiddleware)
	cookie := public.Group("")
	cookie.Use(server.handler.CookieMiddleware)

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
	e.GET("/try", server.handler.Try)
	e.GET("/test", server.handler.HelloPage)

	// Pages (templ) — protected (redirect to / if no session)
	page := e.Group("")
	page.Use(server.handler.PageAuthMiddleware)
	page.GET("/registro", server.handler.RegistroPage)
	page.POST("/registro/start", server.handler.TimerStart)
	page.POST("/registro/stop", server.handler.TimerStop)
	page.GET("/resumen", server.handler.ResumenPage)

	public.GET("/swagger/*", echoSwagger.WrapHandler)

	// Auth (Logto OIDC) — web routes, no api/v1 prefix
	e.GET("/auth/login", server.handler.Login)
	e.GET("/auth/callback", server.handler.Callback)
	e.GET("/auth/logout", server.handler.Logout)
	authCookie := e.Group("")
	authCookie.Use(server.handler.CookieMiddleware)
	authCookie.GET("/auth/link", server.handler.LinkAccount)

	// User
	public.POST("/user", server.handler.CreateUser)
	cookie.POST("/refresh", server.handler.RefreshToken)

	//Entry Time routers
	private.POST("/entry-time", server.handler.CreateEntryTime)
	private.PUT("/entry-time", server.handler.UpdateEntryTime)
	private.GET("/entry-time/:id", server.handler.GetEntryTime)
	private.GET("/entries-time", server.handler.ListEntryTime)
	private.DELETE("/entry-time/:id", server.handler.DeleteEntryTime)

	server.router = e
}
