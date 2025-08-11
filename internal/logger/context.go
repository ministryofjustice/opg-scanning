package logger

import (
	"context"
	"log/slog"
	"net/http"
)

type requestCtx struct{}

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
