package logger

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Logger struct {
	SlogLogger *slog.Logger
}

type loggerContextKey struct{}

func newLogger(environment string) *Logger {
	// Create the base logger using telemetry.NewLogger.
	baseLogger := telemetry.NewLogger("opg-scanning-service")

	slogLogger := baseLogger.With(
		slog.String("environment", environment),
	)
	return &Logger{
		SlogLogger: slogLogger,
	}
}

// Wraps the opg-go-common/telemetry packages StartTracerProvider.
func StartTracerProvider(ctx context.Context, logger *slog.Logger, exportTraces bool) (func(), error) {
	return telemetry.StartTracerProvider(ctx, logger, exportTraces)
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := oteltrace.SpanFromContext(r.Context())

			span.SetAttributes(
				attribute.String("http.target", r.URL.Path),
			)

			loggerWithRequest := logger.With(
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.Any("request", r),
			)

			r = r.WithContext(ContextWithLogger(r.Context(), loggerWithRequest))

			next.ServeHTTP(w, r)
		})
	}
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	// We don't always have the context so we need to check
	// if it exists, otherwise we'll get a panic.
	if val := ctx.Value(loggerContextKey{}); val != nil {
		if logger, ok := val.(*slog.Logger); ok {
			return logger
		}
	}
	return nil
}

func (l *Logger) Info(message string, fields map[string]interface{}, args ...any) {
	if fields != nil {
		l.SlogLogger.Info(message, anyFromAttrs(attrsFromMap(fields))...)
	} else {
		l.SlogLogger.Info(fmt.Sprintf(message, args...))
	}
}

func (l *Logger) InfoWithContext(ctx context.Context, message string, fields map[string]interface{}, args ...any) {
	if ctxLogger := LoggerFromContext(ctx); ctxLogger != nil {
		ctxLogger.Info(message, anyFromAttrs(attrsFromMap(fields))...)
	} else {
		l.Info(message, fields, args...)
	}
}

func (l *Logger) Error(message string, fields map[string]interface{}, args ...any) {
	if fields != nil {
		l.SlogLogger.Error(message, anyFromAttrs(attrsFromMap(fields))...)
	} else {
		l.SlogLogger.Error(fmt.Sprintf(message, args...))
	}
}

func (l *Logger) ErrorWithContext(ctx context.Context, message string, fields map[string]interface{}, args ...any) {
	if ctxLogger := LoggerFromContext(ctx); ctxLogger != nil {
		ctxLogger.Error(message, anyFromAttrs(attrsFromMap(fields))...)
	} else {
		l.Error(message, fields, args...)
	}
}

// converts a map[string]interface{} to a slice of slog.Attr.
func attrsFromMap(fields map[string]interface{}) []slog.Attr {
	if fields == nil {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(fields))
	for key, value := range fields {
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

// converts a slice of slog.Attr to a slice of any.
func anyFromAttrs(attrs []slog.Attr) []any {
	anys := make([]any, len(attrs))
	for i, a := range attrs {
		anys[i] = a
	}
	return anys
}
