package web

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

func init() {
	flag.String("port", "8080", "port to listen on")
}

type Middleware func(next http.HandlerFunc) http.HandlerFunc

// Server represents a web server.
type Server struct {
	port       string
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

func (s *Server) AddRoute(pattern string, handler http.HandlerFunc) {
	// apply any configured middleware
	for _, m := range s.middleware {
		handler = m(handler)
	}

	s.mux.HandleFunc(pattern, handler)
}

// Start starts the web server with the given context and will block until the the context has been cancelled. A context cancellation will cause a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	var err error

	// Configure Server
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", s.port),
		Handler: s.mux,
	}

	// Server loop
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
			ctx.Done()
		}
	}()

	// Allow for a graceful shutdown
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-ctx.Done()
		// Wait for 10 seconds before forcing a shutdown.
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err = httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}

		cancel()
		wg.Done()
	}()

	wg.Wait()
	return err
}
