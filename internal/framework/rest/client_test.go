package rest

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taylorono/go-webservice/internal/framework/metrics"
)

func TestClientFunc_Do(t *testing.T) {
	level := slog.LevelDebug
	ctxKey := "test"

	header := http.Header{}
	header.Add("x-test", "test")

	ctx := context.WithValue(context.Background(), ctxKey, header)
	opts := &slog.HandlerOptions{Level: level}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))
	client := NewClientBuilder("test").
		WithMetricRegistry(metrics.NewPrometheusReporter()).
		WithMiddleware(
			Verbose(level).RequestLogger(),
			AddHeadersFromContext(ctxKey),
		).
		Build()

	req, err := http.NewRequest("GET", "https://official-joke-api.appspot.com/random_joke", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	do, err := client.Do(req.WithContext(ctx))
	assert.NoError(t, err)
	assert.Equal(t, 200, do.StatusCode)
}
