package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/taylorono/go-webservice/internal/framework/config"
	"github.com/taylorono/go-webservice/internal/framework/web"
	"github.com/taylorono/go-webservice/internal/service"
)

func run(ctx context.Context, w io.Writer, args []string) error {
	// listen for SIGINT and SIGTERM
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Load Configuration
	config.InitConfig(ctx)

	// Create a new web server
	webServer := web.NewServer(
		web.WithPort(config.Registry.GetString("PORT")),
	)

	// Create a new service and register routes
	greeter := service.NewService()
	greeter.AddRoutes(webServer)

	// Start the web server
	return webServer.Start(ctx)
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
