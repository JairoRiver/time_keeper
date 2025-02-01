package api

import (
	"net/http"

	"github.com/JairoRiver/time_keeper/internal/api/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Server serve a HTTP request
type Server struct {
	handler *handler.Handler
	logger  *zerolog.Logger
}

func New(handler *handler.Handler, logger *zerolog.Logger) *Server {
	return &Server{handler, logger}
}

func (server *Server) Start(address string) error {
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			server.logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")
			return nil
		},
	}))

	if err := e.Start(address); err != http.ErrServerClosed {
		return err
	}
	return nil
}
