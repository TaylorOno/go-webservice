package web

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
