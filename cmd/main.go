package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/taylorono/go-webservice/internal/api"
	"github.com/taylorono/go-webservice/internal/framework/config"
	"github.com/taylorono/go-webservice/internal/framework/logging"
	"github.com/taylorono/go-webservice/internal/framework/metrics"
	"github.com/taylorono/go-webservice/internal/framework/web"
	"github.com/taylorono/go-webservice/internal/service"
)

var (
	setup   []func(ctx context.Context)
	cleanup []func(ctx context.Context)
)

func run(ctx context.Context, w io.Writer, args []string) error {
	// listen for SIGINT and SIGTERM
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load Configuration
	config.InitConfig(ctx)

	// apply any setup functions
	startup(ctx)

	// defer any cleanup functions
	defer shutdown()

	// Create Metric Reporter
	prometheusReporter := metrics.NewPrometheusReporter()

	// Register debug logging middleware
	var middleware []web.Middleware
	if logging.Level() <= slog.LevelDebug {
		middleware = append(middleware, logging.HttpLoggingMiddleware)
	}

	// Create business logic services
	greeter := service.NewService()

	// Create a new web server
	webServer := web.NewServer(
		web.WithPort(config.Registry.GetString("PORT")),
		web.WithDebugPort(config.Registry.GetString("DEBUG_PORT")),
		web.WithMiddleware(logging.HttpLoggingMiddleware),
		web.WithMetricRegistry(prometheusReporter),
	)

	// Register route handlers
	api.NewGreeterHandler(greeter).Routes(webServer)

	wg := sync.WaitGroup{}

	// Launch the web server in a goroutine
	wg.Go(func() {
		if err := webServer.Start(ctx); err != nil {
			slog.Info("web server stopped", "error", err)
		}
	})

	// Start the web server
	wg.Wait()
	return nil
}

func startup(ctx context.Context) {
	wg := sync.WaitGroup{}
	for _, setupFunc := range setup {
		wg.Go(func() { setupFunc(ctx) })
	}
	wg.Wait()
}

func shutdown() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	wg := sync.WaitGroup{}
	for _, cleanupFunc := range cleanup {
		wg.Go(func() { cleanupFunc(shutdownCtx) })
	}
	wg.Wait()
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
