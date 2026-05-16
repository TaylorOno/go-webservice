package components

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pathToServerBinary string
)

type (
	portKey    struct{}
	sessionKey struct{}
)

type service struct {
	session *gexec.Session
	output  io.Reader
}

func (s *service) Stop(t *testing.T) error {
	if runtime.GOOS == "windows" {
		return s.stopWindows(t)
	}

	return s.stop(t)
}

func (s *service) stopWindows(t *testing.T) error {
	s.session.Kill()

	ok := assert.Eventually(t, func() bool { return s.session.ExitCode() == 1 }, 10*time.Second, 100*time.Millisecond)
	if !ok {
		return fmt.Errorf("server failed to exit; %d != 1", s.session.ExitCode())
	}

	return s.log()
}

func (s *service) stop(t *testing.T) error {
	s.session.Interrupt()

	ok := assert.Eventually(t, func() bool { return s.session.ExitCode() == 0 }, 10*time.Second, 100*time.Millisecond)
	if !ok {
		return fmt.Errorf("server failed to exit; %d != 0", s.session.ExitCode())
	}

	return s.log()
}

func (s *service) log() error {
	fmt.Println("=== Logs")
	_, err := io.Copy(os.Stdout, s.output)
	if err != nil {
		return fmt.Errorf("failed to copy logs: %w", err)
	}
	fmt.Println("")

	return nil
}

// Build sets up pre-suite and post-suite actions, initializes the server binary, and defines scenario lifecycle hooks.
func Build(t *testing.T, suiteCtx *godog.TestSuiteContext) {
	// Compile the server binary before each suite
	suiteCtx.BeforeSuite(func() {
		defer func(t time.Time) {
			slog.Info("build successful", slog.String("path", pathToServerBinary), slog.Duration("time", time.Since(t)))
		}(time.Now())

		var err error
		pathToServerBinary, err = gexec.Build("github.com/taylorono/go-webservice/cmd")
		require.NoError(t, err)
	})

	suiteCtx.AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
		slog.Info("cleanup build artifacts completed")
	})

	scenarioCtx := suiteCtx.ScenarioContext()

	scenarioCtx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		server, ok := ctx.Value(sessionKey{}).(*service)
		if !ok {
			return ctx, nil
		}

		return ctx, server.Stop(t)
	})

	// Define the steps
	scenarioCtx.Step(`^the server started(?: with profile (.*))?`, func(ctx context.Context, profile string) (context.Context, error) {
		defer func(t time.Time) {
			slog.Info("server started", slog.String("path", pathToServerBinary), slog.Duration("time", time.Since(t)))
		}(time.Now())

		// start the application on a random port
		port, err := findAvailablePort()
		if err != nil {
			return ctx, fmt.Errorf("failed to get port: %w", err)
		}

		// start the application on available port with the provided profile
		t.Setenv("PROFILE", profile)
		t.Setenv("PORT", fmt.Sprintf("%d", port))
		cmd := exec.CommandContext(ctx, pathToServerBinary)
		slog.Info("starting server", "cmd", cmd.String())

		out := &bytes.Buffer{}
		session, err := gexec.Start(cmd, out, out)
		if err != nil {
			slog.Error("failed to start server", slog.String("error", err.Error()))
			return ctx, err
		}

		ctx = context.WithValue(ctx, portKey{}, port)
		ctx = context.WithValue(ctx, sessionKey{}, &service{
			session: session,
			output:  out,
		})

		return ctx, nil
	})

	scenarioCtx.Step(`^the server is running$`, func(ctx context.Context) error {
		port := GetServicePort(ctx)

		ok := assert.Eventually(t, func() bool { _, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port)); return err == nil }, 5*time.Second, 100*time.Millisecond)
		if !ok {
			return fmt.Errorf("the server is not listening on port %d", port)
		}

		return nil
	})
}

// GetServicePort retrieves the service port from the provided context, assuming it has been stored with a custom key.
func GetServicePort(ctx context.Context) int {
	return ctx.Value(portKey{}).(int)
}

func findAvailablePort() (port int, err error) {
	l, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", "0"))
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
