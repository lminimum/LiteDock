package httpserver

import (
	"net"
	"time"
)

// Option configures the HTTP server with custom settings.
type Option func(*Server)

// Port sets the port for the HTTP server to listen on.
func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}

// Prefork enables or disables prefork mode for the HTTP server.
func Prefork(prefork bool) Option {
	return func(s *Server) {
		s.prefork = prefork
	}
}

// ReadTimeout sets the maximum duration for reading the entire request.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WriteTimeout sets the maximum duration before timing out the write of the response.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// ShutdownTimeout sets the maximum time to wait for active connections to terminate.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
