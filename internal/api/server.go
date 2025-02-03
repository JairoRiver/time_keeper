package api

import (
	"net/http"

	"github.com/JairoRiver/time_keeper/internal/api/handler"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Server serve a HTTP request
type Server struct {
	handler *handler.Handler
	logger  *zerolog.Logger
	router  *echo.Echo
}

func New(handler *handler.Handler, logger *zerolog.Logger) *Server {
	server := Server{
		handler: handler,
		logger:  logger,
	}
	server.setupRouter()
	return &server
}

func (server *Server) Start(address string) error {
	if err := server.router.Start(address); err != http.ErrServerClosed {
		return err
	}
	return nil
}
