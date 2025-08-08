package logger

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type requestCtx struct{}

// NewContextFromOld creates a new context.Background with the trace span and
// request values to be logged from oldCtx.
func NewContextFromOld(oldCtx context.Context) context.Context {
	newCtx := context.Background()

	if group := requestFromContext(oldCtx); group.Key != "" {
		newCtx = context.WithValue(newCtx, requestCtx{}, group)
	}

	span := trace.SpanFromContext(oldCtx)
	return trace.ContextWithSpan(newCtx, span)
}

func contextWithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestCtx{}, slog.Group("request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.String()),
	))
}

func requestFromContext(ctx context.Context) slog.Attr {
	val, _ := ctx.Value(requestCtx{}).(slog.Attr)
	return val
}
