package api

import (
	_ "github.com/JairoRiver/time_keeper/docs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (server *Server) setupRouter() {
	e := echo.New()
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

	public.GET("/swagger/*", echoSwagger.WrapHandler)

	//User routers
	public.POST("/user", server.handler.CreateUser)

	//Entry Time routers
	private.POST("/entry-time", server.handler.CreateEntryTime)
	private.PUT("/entry-time", server.handler.UpdateEntryTime)
	private.GET("/entry-time/:id", server.handler.GetEntryTime)
	private.GET("/entries-time", server.handler.ListEntryTime)
	private.DELETE("/entry-time/:id", server.handler.DeleteEntryTime)

	server.router = e
}
