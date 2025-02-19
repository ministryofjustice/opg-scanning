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
	if len(fields) == 0 {
		return l
	}
	newSlogLogger := l.SlogLogger.With(convertFieldsToAny(fields)...)
	return &Logger{
		cfg:        l.cfg,
		SlogLogger: newSlogLogger,
	}
}

func (l *Logger) Info(message string, fields map[string]interface{}, args ...any) {
	if fields != nil {
		l.WithFields(fields).SlogLogger.Info(message)
	} else {
		logMessage := fmt.Sprintf(message, args...)
		l.SlogLogger.Info(logMessage)
	}
}

func (l *Logger) Error(message string, fields map[string]interface{}, args ...any) {
	if fields != nil {
		l.WithFields(fields).SlogLogger.Error(message)
	} else {
		logMessage := fmt.Sprintf(message, args...)
		l.SlogLogger.Error(logMessage)
	}
}

func convertFieldsToAny(fields map[string]interface{}) []any {
	anySlice := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		anySlice = append(anySlice, key, value)
	}
	return anySlice
}
