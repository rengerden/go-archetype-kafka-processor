package config

import (
	"fmt"
	"net/http"
)

// Handler is a HTTP endpoint handler
type Handler func(w http.ResponseWriter, r *http.Request)

// HTTPServer is an abstraction from a HTTP server
type HTTPServer interface {
	// Registers a HTTP Endpoint
	RegisterEndpoint(endpoint string, handler Handler)

	// Runs the HTTP server
	Run()
}

// DefaultHTTPServer is the default implementation of HTTPServer
type DefaultHTTPServer struct {
	// The struct logger
	logger Logger

	// The port to run on
	Port int
}

// RegisterEndpoint registers a Handler to a specific endpoint
func (dhs DefaultHTTPServer) RegisterEndpoint(endpoint string, handler Handler) {
	dhs.logger.Debug(fmt.Sprintf("Registering handler on endpoint '%s'", endpoint))
	http.HandleFunc(endpoint, handler)
}

// Run runs the HTTP server
func (dhs DefaultHTTPServer) Run() {
	dhs.logger.Info(fmt.Sprintf("Starting HTTP server on port '%d'...", dhs.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", dhs.Port), nil)
}

// NewDefaultHTTPServer creates a new DefaultHttpServer
func NewDefaultHTTPServer(port int) *DefaultHTTPServer {
	return &DefaultHTTPServer{
		logger: NewLogger(),
		Port:   port,
	}
}
