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

	"github.com/taylorono/go-lib/config"
	"github.com/taylorono/go-lib/headers"
	"github.com/taylorono/go-lib/logging"
	"github.com/taylorono/go-lib/metrics"
	"github.com/taylorono/go-lib/rest"
	"github.com/taylorono/go-lib/web"
	"github.com/taylorono/go-webservice/internal/api"
	"github.com/taylorono/go-webservice/internal/joker"
	"github.com/taylorono/go-webservice/internal/service"
	"github.com/taylorono/go-webservice/ui"
)

var (
	// init functions can register setup functions that will be run before the application starts
	setup []func(ctx context.Context)

	// init functions can register cleanup functions that will be run after the application stops
	cleanup []func(ctx context.Context)
)

func main() {
	// Create a context that will be canceled when the application is shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	slog.Info("shutdown successful")
}

func run(ctx context.Context, w io.Writer, args []string) error {
	setup = append(setup, config.InitConfig, logging.InitLogger, logging.WithEnabledFunction(config.GetLogLevel))

	// apply any setup functions
	startup(ctx)

	// defer any cleanup functions
	defer shutdown()

	// Create Metric Reporter
	prometheusReporter := metrics.NewPrometheusReporter()

	// Enable Logger Metrics
	logging.WithMetricReporter(prometheusReporter)

	// Create business logic services
	greeter := initializeGreetingService(ctx, prometheusReporter)

	// Creates a web server for the transport layer
	webServer := createWebServer(prometheusReporter)

	// Register service route handlers
	api.NewGreeterHandler(greeter).Routes(webServer)

	// Register UI route handlers
	ui.NewUIHandler().Routes(webServer)

	// Launch the web server in a goroutine
	wg := sync.WaitGroup{}
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

func createWebServer(prometheusReporter *metrics.PrometheusReporter) *web.Server {
	// Register debug logging middleware
	var middleware []web.Middleware
	if logging.Level() <= slog.LevelDebug {
		middleware = append(middleware, logging.HttpLoggingMiddleware)
	}

	// Add Platform Headers middleware
	// This will copy the platform headers from requests into the context
	ph := headers.NewPlatformHeaders(headers.StartsWithHeaders([]string{"X-Platform"}))
	middleware = append(middleware, ph.ContextHeaders)

	// Create a new web server
	webServer := web.NewServer(
		web.WithPort(config.Registry.GetString("PORT")),
		web.WithDebugPort(config.Registry.GetString("DEBUG_PORT")),
		web.WithMiddleware(middleware...),
		web.WithMetricRegistry(prometheusReporter),
	)

	return webServer
}

func initializeGreetingService(_ context.Context, reporter *metrics.PrometheusReporter) *service.Greeter {

	// Sample RestClient Dependency
	jokeClient := rest.NewClientBuilder("jokes").
		WithMetricRegistry(reporter).
		WithMiddleware(headers.Middleware).
		WithMiddleware(
			rest.Verbose().WithHandler(logging.ComponentLoggerFor("jokes")).RequestLogger(),
		).
		Build()

	// Create downstream dependency wrapper
	jokeService := joker.NewJokeProvider(jokeClient, config.Registry.GetString("JOKES.URL"))

	// Create business logic services
	return service.NewGreater(jokeService)
}
