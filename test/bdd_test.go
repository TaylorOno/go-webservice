//go:build integration

package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/taylorono/go-webservice/test/components"
)

type contextKey string

const (
	lastBodyKey contextKey = "lastBody"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		TestSuiteInitializer: InitializeSuite(t),
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeSuite(t *testing.T) func(*godog.TestSuiteContext) {
	return func(suiteContext *godog.TestSuiteContext) {
		// Build the application once per test suite
		components.Build(t, suiteContext)

		// only configure wiremock if we are not running against a real upstreams
		if true {
			// Start wiremock once per test suite
			components.Wiremock(t, suiteContext)

			// Route real upstream URLs through WireMock acting as a forward proxy.
			suiteContext.BeforeSuite(func() {
				t.Setenv("HTTP_PROXY", components.WiremockURL)
				t.Setenv("HTTPS_PROXY", components.WiremockURL) // TODO add wiremock certificate for https support
				t.Setenv("NO_PROXY", "localhost,127.0.0.1,::1")
			})
		}
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^I send a (GET|POST) request(?: with the body `+"`"+`([^`+"`"+`]*)`+"`"+`)? to the (.*) endpoint$`, iSendARequestToTheEndpoint)
	sc.Step(`^I see the message (.*)$`, iSeeTheMessage)
}

func iSendARequestToTheEndpoint(ctx context.Context, method, body, endpoint string) (context.Context, error) {
	port := components.GetServicePort(ctx)

	url := fmt.Sprintf("http://localhost:%d%s", port, endpoint)
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return ctx, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, lastBodyKey, string(b)), nil
}

func iSeeTheMessage(ctx context.Context, expected string) error {
	lastBody, ok := ctx.Value(lastBodyKey).(string)
	if !ok {
		return fmt.Errorf("last body not found in context")
	}

	expected = strings.Trim(expected, `"`)
	if !strings.Contains(lastBody, expected) {
		return fmt.Errorf("expected to see %q, but got %q", expected, lastBody)
	}

	return nil
}
