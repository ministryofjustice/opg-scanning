package logger

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
)

type logAttrKey struct{}

func ContextWithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if v := attrsFromContext(ctx); len(attrs) > 0 {
		return context.WithValue(ctx, logAttrKey{}, slices.Concat(v, attrs))
	}

	return context.WithValue(ctx, logAttrKey{}, attrs)
}

func contextWithRequest(ctx context.Context, r *http.Request) context.Context {
	return ContextWithAttrs(ctx, slog.Group("request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.String()),
	))
}

func attrsFromContext(ctx context.Context) []slog.Attr {
	v, _ := ctx.Value(logAttrKey{}).([]slog.Attr)
	return v
}
