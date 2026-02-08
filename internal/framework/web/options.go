package web

import (
	"github.com/taylorono/go-webservice/internal/framework/metrics"
)

type OptionFunc func(*Server)

func WithPort(port string) OptionFunc {
	return func(o *Server) {
		o.port = port
	}
}

func WithDebugPort(port string) OptionFunc {
	return func(o *Server) {
		o.debugPort = port
	}
}

func WithMetricRegistry(registry metrics.Reporter) OptionFunc {
	return func(o *Server) {
		// Register metrics routes before middleware to avoid instrumentation.
		registry.Routes(o.mux)

		// Add default instrumentation middleware
		o.middleware = append(o.middleware, metrics.HttpMiddleware(registry))
	}
}
