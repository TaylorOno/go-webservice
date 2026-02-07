package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taylorono/go-webservice/internal/framework/config"
	"github.com/taylorono/go-webservice/internal/framework/web"
	"github.com/taylorono/go-webservice/internal/service"
	"golang.org/x/sync/errgroup"
)

var (
	setup   []func(ctx context.Context)
	cleanup []func(ctx context.Context)
)

func run(ctx context.Context, w io.Writer, args []string) error {
	// listen for SIGINT and SIGTERM
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	// apply any setup functions
	for _, setupFunc := range setup {
		setupFunc(ctx)
	}

	// defer any cleanup functions
	defer func() {
		for _, cleanupFunc := range cleanup {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cleanupFunc(shutdownCtx)
		}
	}()

	// Load Configuration
	config.InitConfig(ctx)

	// Create a new web server
	webServer := web.NewServer(
		web.WithPort(config.Registry.GetString("PORT")),
		web.WithDebugPort(config.Registry.GetString("DEBUG_PORT")),
	)

	// Create a new service and register routes
	greeter := service.NewService()
	greeter.AddRoutes(webServer)

	// Launch the web server in a goroutine
	eg := &errgroup.Group{}
	eg.Go(func() error { return webServer.Start(ctx) })

	// Start the web server
	return eg.Wait()
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
