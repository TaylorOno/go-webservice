package web

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/taylorono/go-webservice/internal/framework/profile"
)

func init() {
	flag.String("port", "8080", "port to listen on")
	flag.String("debug-port", "", "when set pprof will be enabled on this port")
}

type Middleware func(next http.HandlerFunc) http.HandlerFunc

// Server represents a web server suitable for kubernetes deployments.
type Server struct {
	port       string
	debugPort  string
	mux        *http.ServeMux
	middleware []Middleware
}

// NewServer Creates a new web server with the given options.
func NewServer(opts ...OptionFunc) *Server {
	// default server
	s := &Server{
		port:       "8080",
		mux:        http.NewServeMux(),
		middleware: []Middleware{},
	}

	// apply config overrides
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// HandleFunc registers a new route with the given pattern and handler function applying any global middleware.
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	// apply any configured middleware
	for _, m := range s.middleware {
		handler = m(handler)
	}

	s.mux.HandleFunc(pattern, handler)
}

// Start starts the web server with the given context and will block until the context has been canceled. A context cancellation will cause a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	var err error

	// Configure Server
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", s.port),
		Handler: s.mux,
	}

	// Server loop
	go func() {
		slog.Info(fmt.Sprintf("listening on %s\n", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
			ctx.Done()
		}
	}()

	// Launch pprof if the port has been specified
	if s.debugPort != "" {
		profile.ListenAndServe(ctx, s.debugPort)
	}

	// Allow for a graceful shutdown
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		// Wait for 10 seconds before forcing a shutdown.
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err = httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	})

	wg.Wait()
	return err
}
