package server

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// Handler is a interface for represents request handlers
type Handler interface {
	Handler() func(echo.Context) error
}

// Server is the struct represents the server
type Server struct {
	// Port is the port number of the server
	Port string

	// Handlers is the map of path to handler
	Handlers map[string]Handler

	// Echo is the echo instance
	Echo *echo.Echo
}

// New returns a new server
func New(port string, handlers map[string]Handler) *Server {
	// create a new echo instance
	e := echo.New()

	// set logger
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logrus.New().WithFields(logrus.Fields{
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")

			return nil
		},
	}))

	return &Server{
		Port:     port,
		Handlers: handlers,
		Echo:     e,
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Register handlers
	for path, handler := range s.Handlers {
		s.Echo.POST(path, handler.Handler())
	}

	return s.Echo.Start(s.Port)
}

// Stop stops the server
func (s *Server) Stop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.Echo.Shutdown(ctx)
}
