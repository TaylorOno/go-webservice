//go:build local

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/taylorono/go-webservice/internal/framework/testcontainer"
)

// init sets default environment variables for Podman compatibility on Windows and disables Ryuk for stability.
func init() {
	if os.Getenv("GOOS") == "windows" {
		// Podman on Windows support
		// Default to the standard Podman named pipe on Windows if DOCKER_HOST is not set
		os.Setenv("DOCKER_HOST", "npipe:////./pipe/podman-machine-default")

		// Disable Ryuk (reaper) as it often has issues with Podman on Windows
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

// init initializes the setup function by starting a Kafka test container and registering its cleanup procedure.
func init() {
	setup = append(setup, func(ctx context.Context) {
		// start the Kafka test container
		kafkaContainer := testcontainer.StartKafkaContainer(ctx)

		// TODO: create any topics you might need.

		// register the cleanup function
		cleanup = append(cleanup, func(ctx context.Context) {
			slog.Info("removing kafka container")
			err := kafkaContainer.Terminate(ctx)
			if err != nil {
				slog.Error("failed to terminate kafka container", slog.String("error", err.Error()))
			}
		})
	})
}
