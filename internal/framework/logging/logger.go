package logging

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"testing"
)

var (
	lvl          slog.Level
	enableJSON   bool
	enableSource bool
	handler      slog.Handler
)

func init() {
	flag.TextVar(&lvl, "log-level", slog.LevelInfo, "log level: [debug info warn error]")
	flag.BoolVar(&enableJSON, "log-json", false, "enable structured logging")
	flag.BoolVar(&enableSource, "log-source", false, "enable logging of source file and line")
}

func Level() slog.Level {
	return lvl
}

func InitLogger(ctx context.Context) {
	if !testing.Testing() {
		flag.Parse()
		if !enableJSON {
			opts := &slog.HandlerOptions{Level: lvl, AddSource: enableSource}
			handler = NewHandler(slog.NewTextHandler(os.Stdout, opts))
		} else {
			opts := &slog.HandlerOptions{Level: lvl, AddSource: enableSource}
			handler = NewHandler(slog.NewJSONHandler(os.Stdout, opts))
		}
	}

	slog.SetDefault(slog.New(handler))
}

func GetHandler() slog.Handler {
	return handler
}
