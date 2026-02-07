package web

import "net/http"

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

func WithRoutes(addRoutes func(mux *http.ServeMux)) OptionFunc {
	return func(o *Server) {
		addRoutes(o.mux)
	}
}
