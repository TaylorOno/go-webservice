package logging

import (
	"flag"
	"log/slog"
	"os"
	"testing"
)

var (
	lvl          slog.Level
	enableJSON   bool
	enableSource bool
)

func init() {
	flag.TextVar(&lvl, "log-level", slog.LevelInfo, "log level: debug info warn error")
	flag.BoolVar(&enableJSON, "log-json", false, "enable structured logging")
	flag.BoolVar(&enableSource, "log-source", false, "enable logging of source file and line")

}

func init() {
	if !testing.Testing() {
		flag.Parse()
		if !enableJSON {
			slog.SetLogLoggerLevel(lvl)
			return
		}

		opts := &slog.HandlerOptions{Level: lvl, AddSource: enableSource}
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
	}
}

func Level() slog.Level {
	return lvl
}
