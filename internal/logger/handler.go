package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type slogHandler struct {
	handler slog.Handler
}

func newSlogHandler(h slog.Handler) slog.Handler {
	return &slogHandler{handler: h}
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		record.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
	}

	if attrs := attrsFromContext(ctx); len(attrs) > 0 {
		record.AddAttrs(attrs...)
	}

	return h.handler.Handle(ctx, record)
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newSlogHandler(h.handler.WithAttrs(attrs))
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return newSlogHandler(h.handler.WithGroup(name))
}
