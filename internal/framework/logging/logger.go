package logging

import (
	"flag"
	"log/slog"
	"os"
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
	flag.Parse()
}

func init() {
	if !enableJSON {
		slog.SetLogLoggerLevel(lvl)
		return
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: enableSource,
	})))
}

func Level() slog.Level {
	return lvl
}
