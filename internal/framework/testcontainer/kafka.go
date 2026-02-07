package testcontainer

import (
	"context"
	"log/slog"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

// StartKafkaContainer starts a Kafka container and returns its bootstrap servers.
// It also returns a cleanup function to stop the container.
func StartKafkaContainer(ctx context.Context) *kafka.KafkaContainer {
	slog.Info("Starting Kafka container...")

	// Podman on Windows support
	if os.Getenv("DOCKER_HOST") == "" {
		// Default to the standard Podman named pipe on Windows if DOCKER_HOST is not set
		os.Setenv("DOCKER_HOST", "npipe:////./pipe/podman-machine-default")
	}

	// Disable Ryuk (reaper) as it often has issues with Podman on Windows
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/cp-kafka:7.4.0",
		testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			if hostConfig.NetworkMode == "bridge" {
				hostConfig.NetworkMode = ""
			}
		}),
		testcontainers.WithEnv(map[string]string{
			"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:9093,CONTROLLER://0.0.0.0:9094,BROKER://0.0.0.0:9092",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT,BROKER:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@localhost:9094",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_LOG_DIRS":                                 "/tmp/kraft-combined-logs",
			"CLUSTER_ID":                                     "MkU3OEVBNTcwNTJENDM2Qk",
		}),
	)
	if err != nil {
		return nil
	}

	bootstrapServers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		return nil
	}

	bootstrapServersStr := ""
	if len(bootstrapServers) > 0 {
		bootstrapServersStr = bootstrapServers[0]
		for i := 1; i < len(bootstrapServers); i++ {
			bootstrapServersStr += "," + bootstrapServers[i]
		}
	}

	slog.Info("Kafka container started", slog.String("bootstrapServers", bootstrapServersStr))
	return kafkaContainer
}
