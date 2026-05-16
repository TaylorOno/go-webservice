package logging

import (
	"context"
	"log/slog"
)

type Handler struct {
	slog.Handler
}

func NewHandler(handler slog.Handler) *Handler {
	return &Handler{handler}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}
