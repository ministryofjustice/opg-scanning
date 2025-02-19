package logger

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/config"
)

type Logger struct {
	cfg        *config.Config
	SlogLogger *slog.Logger
}

func NewLogger(cfg *config.Config) *Logger {
	// Create the base logger using telemetry.NewLogger.
	baseLogger := telemetry.NewLogger("opg-scanning-service")
	
	slogLogger := baseLogger.With(
		slog.String("environment", cfg.App.Environment),
	)
	return &Logger{
		cfg:        cfg,
		SlogLogger: slogLogger,
	}
}

// Wraps the opg-go-common/telemetry packages StartTracerProvider.
func StartTracerProvider(ctx context.Context, logger *slog.Logger, exportTraces bool) (func(), error) {
	return telemetry.StartTracerProvider(ctx, logger, exportTraces)
}

// Returns the opg-go-common/telemetry packages HTTP middleware.
func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return telemetry.Middleware(logger)
}

// Retrieves the logger from the context using the opg-go-common/telemetry packages helper.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	return telemetry.LoggerFromContext(ctx)
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{
		cfg:        l.cfg,
		SlogLogger: l.SlogLogger.With(anyFromAttrs(attrsFromMap(fields))...),
	}
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