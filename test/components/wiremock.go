package components

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wiremock/go-wiremock"
)

const (
	imageName = "wiremock/wiremock"
	imageTag  = "3.13.2-2"
)

var (
	// Record toggles WireMock proxy recording mode via `-record` test flag.
	//
	// When enabled, scenarios run against upstream services through WireMock and
	// captured mappings are saved under `test/features/recordings`.
	Record bool

	// WiremockURL is the reachable base URL for the running WireMock container.
	// It is initialized once in `BeforeSuite` and reused across scenarios.
	WiremockURL string

	// wiremockContainer stores the singleton test container instance for the
	// whole BDD suite lifecycle.
	wiremockContainer testcontainers.Container
)

func init() {
	// Register CLI flag once package is initialized so BDD runs can opt in to
	// proxy recording without changing code.
	flag.BoolVar(&Record, "record", false, "update Wiremock responses")
}

// init sets default environment variables for Podman compatibility on Windows and disables Ryuk for stability.
func init() {
	// Podman on Windows support
	if runtime.GOOS == "windows" {
		// Default to the standard Podman named pipe on Windows if DOCKER_HOST is not set
		os.Setenv("DOCKER_HOST", "npipe:////./pipe/podman-machine-default")

		// Disable Ryuk (reaper) as it often has issues with Podman on Windows
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

// Wiremock sets up the WireMock container to start once per suite and exit cleanly.
func Wiremock(t *testing.T, suiteCtx *godog.TestSuiteContext) {
	suiteCtx.BeforeSuite(func() {
		// Defensive guard in case setup is invoked more than once.
		if wiremockContainer != nil {
			return
		}

		// Start WireMock as an isolated test dependency.
		req := testcontainers.ContainerRequest{
			Image:        fmt.Sprintf("%s:%s", imageName, imageTag),
			Cmd:          wiremockArgs(),
			ExposedPorts: []string{"8080/tcp"}, //TODO use random port
			WaitingFor:   wait.ForHTTP("/__admin").WithPort("8080/tcp"),
			Name:         "wiremock-test-container",
		}

		var err error
		wiremockContainer, err = testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(t, err)

		// Resolve host/port after container startup to build the admin/runtime URL.
		WiremockURL = wiremockURL(t)
		slog.Info("WireMock started", slog.String("url", WiremockURL))
	})

	suiteCtx.AfterSuite(func() {
		if wiremockContainer != nil {
			// Use bounded shutdown to avoid hanging test runs on container teardown.
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			err := wiremockContainer.Terminate(ctx)
			if err != nil {
				slog.Error("failed to terminate WireMock container", slog.String("error", err.Error()))
				return
			}

			slog.Info("WireMock container terminated")
		}
	})

	suiteCtx.ScenarioContext().Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if wiremockContainer != nil {
			// Admin client controls WireMock state between scenarios.
			wiremockClient := wiremock.NewClient(WiremockURL)

			// Reset WireMock to clear previous scenario mappings
			if err := wiremockClient.Reset(); err != nil {
				slog.Error("failed to reset WireMock", slog.String("error", err.Error()))
				return ctx, err
			}

			// Delete all requests to clear WireMock request journal
			if err := wiremockClient.DeleteAllRequests(); err != nil {
				slog.Error("failed to clear WireMock request journal", slog.String("error", err.Error()))
				return ctx, err
			}

			// Add a delay to ensure WireMock is ready after reset
			// TODO: be smarter about this
			time.Sleep(500 * time.Millisecond)

			// In recording mode, allow requests to proxy through and capture traffic
			// after scenario completion.
			if Record {
				slog.Info("wireMock proxy recording enabled", slog.String("scenario", sc.Name))
				return ctx, nil
			}

			// In playback mode, import previously recorded scenario mappings.
			if err := loadScenarioMock(ctx, sc.Name); err != nil {
				slog.Error("failed to load scenario mappings", slog.String("scenario", sc.Name), slog.String("error", err.Error()))
			}
		}

		return ctx, nil
	})

	suiteCtx.ScenarioContext().After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if wiremockContainer != nil && Record {
			// Persist recorded mappings for the scenario so subsequent runs can use deterministic playback.
			if err = writeMockData(t, WiremockURL, sc.Name); err != nil {
				slog.Error("failed to stop recording", slog.String("error", err.Error()))
				return ctx, nil
			}

			slog.Info("wiremock recording stopped and saved", slog.String("scenario", sc.Name))
		}

		return ctx, nil
	})
}

// wiremockArgs enables proxying and trusts all proxy targets.
func wiremockArgs() []string {
	// Base arguments keep WireMock in browser proxy mode while allowing HTTPS
	// upstreams with self-signed certificates in test environments.
	args := []string{
		"--enable-browser-proxying",
		"--trust-all-proxy-targets",
	}

	if Record {
		// Recording mode: pass unmatched requests through to upstream service.
		return append(args, "--proxy-pass-through=true")
	}

	// Playback mode: fail unmatched requests to keep tests deterministic.
	return append(args, "--proxy-pass-through=false")
}

// writeMockData saves the recorded snapshot to a file for future reference.
func writeMockData(t *testing.T, wiremockURL, scenarioName string) error {
	// SnapshotRequest asks WireMock to return all generated mappings in one payload so we can write a single scenario-specific recording artifact.
	snapShotRequest := SnapshotRequest{
		OutputFormat:       "FULL",
		RepeatsAsScenarios: true, // captures repeat requests in the event responses differ
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(snapShotRequest)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, wiremockURL+"/__admin/recordings/snapshot", buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req.WithContext(t.Context()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Mappings []map[string]interface{} `json:"mappings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode recording result: %w", err)
	}

	// Nothing matched/proxied during this scenario.
	if len(result.Mappings) == 0 {
		return nil
	}

	// Persist recordings under versioned test fixtures.
	mappingsDir := filepath.Join("features", "recordings")
	if err := os.MkdirAll(mappingsDir, 0755); err != nil {
		return err
	}

	// File naming aligns with feature scenario names while remaining portable.
	normalizedName := strings.ReplaceAll(strings.ToLower(scenarioName), " ", "_")
	file := filepath.Join(mappingsDir, fmt.Sprintf("%s.json", normalizedName))
	openFile, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer openFile.Close()

	encoder := json.NewEncoder(openFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(result)
	if err != nil {
		return err
	}

	return nil
}

func wiremockURL(t *testing.T) string {
	// testcontainers may map container ports dynamically; always resolve at runtime.
	host, err := wiremockContainer.Host(t.Context())
	require.NoError(t, err, "failed to get WireMock host")

	port, err := wiremockContainer.MappedPort(t.Context(), "8080/tcp")
	require.NoError(t, err, "failed to get WireMock port")

	return fmt.Sprintf("http://%s:%s", host, port.Port())
}

func loadScenarioMock(ctx context.Context, scenarioName string) error {
	// Scenario recordings are stored as WireMock import payloads.
	recordingDir := filepath.Join("features", "recordings")
	normalizedName := strings.ReplaceAll(strings.ToLower(scenarioName), " ", "_")
	recording := filepath.Join(recordingDir, fmt.Sprintf("%s.json", normalizedName))

	if _, err := os.Stat(recording); os.IsNotExist(err) {
		slog.Debug("no scenario specific mock found", slog.String("path", recording))
		return nil
	}

	data, err := os.ReadFile(recording)
	if err != nil {
		return err
	}

	// Import mappings into WireMock before scenario steps execute.
	resp, err := http.Post(WiremockURL+"/__admin/mappings/import", "application/json", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type SnapshotRequest struct {
	CaptureHeaders     CaptureHeaders `json:"captureHeaders"`
	OutputFormat       string         `json:"outputFormat"`
	Persist            bool           `json:"persist"`
	RepeatsAsScenarios bool           `json:"repeatsAsScenarios"`
}

type CaptureHeaders struct {
	Accept      struct{} `json:"Accept"`
	ContentType struct{} `json:"Content-Type"`
	Host        struct{} `json:"Host"`
}
