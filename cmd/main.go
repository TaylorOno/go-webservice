package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/taylorono/go-webservice/internal/server"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Configure Server
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", "42069"),
		Handler: server.NewServer(),
	}

	// Server loop
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
			ctx.Done()
		}
	}()

	// Allow for graceful shutdown
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-ctx.Done()
		// Wait for 10 seconds before forcing a shutdown.
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}

		cancel()
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
