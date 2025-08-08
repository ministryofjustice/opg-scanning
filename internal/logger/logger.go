package logger

import (
	"context"
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

func New(environment string) *Logger {
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

func (l *Logger) Info(message string, args ...any) {
	l.SlogLogger.Info(message, args...)
}

func (l *Logger) InfoContext(ctx context.Context, message string, args ...any) {
	if ctxLogger := LoggerFromContext(ctx); ctxLogger != nil {
		ctxLogger.InfoContext(ctx, message, args...)
	} else {
		l.SlogLogger.InfoContext(ctx, message, args...)
	}
}

func (l *Logger) Error(message string, args ...any) {
	l.SlogLogger.Error(message, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, message string, args ...any) {
	if ctxLogger := LoggerFromContext(ctx); ctxLogger != nil {
		ctxLogger.ErrorContext(ctx, message, args...)
	} else {
		l.SlogLogger.ErrorContext(ctx, message, args...)
	}
}
