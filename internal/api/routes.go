package api

import (
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
	router := e.Group("api/v1")

	// @title Short Link API
	// @version 1.0
	// @description Testing Swagger APIs.
	// @termsOfService http://swagger.io/terms/// @contact.name API Support
	// @contact.url http://www.swagger.io/support
	// @contact.email support@swagger.io// @securityDefinitions.apiKey JWT
	// @in header
	// @name token// @license.name Apache 2.0
	// @license.url http://www.apache.org/licenses/LICENSE-2.0.html// @host localhost:8081
	// @BasePath /v1// @schemes http
	// Swagger documentation
	router.GET("/swagger/*", echoSwagger.WrapHandler)

	//User routers
	router.POST("/user", server.handler.CreateUser)

	//Entry Time routers
	router.POST("/entry-time", server.handler.CreateEntryTime)
	router.PUT("/entry-time", server.handler.UpdateEntryTime)
	router.GET("/entry-time/:id", server.handler.GetEntryTime)
	router.GET("/entries-time", server.handler.ListEntryTime)
	router.DELETE("/entry-time/:id", server.handler.DeleteEntryTime)

	server.router = e
}
