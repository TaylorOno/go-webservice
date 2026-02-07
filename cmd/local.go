//go:build local

package main

import (
	"context"
	"log/slog"

	"github.com/taylorono/go-webservice/internal/framework/testcontainer"
)

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
