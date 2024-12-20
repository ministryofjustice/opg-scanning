package logger

import (
	"fmt"
	"log/slog"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/config"
)

type Logger struct {
	cfg        *config.Config
	SlogLogger *slog.Logger
}

func NewLogger(cfg *config.Config) *Logger {
	slogLogger := telemetry.NewLogger("opg-data-lpa-store/getlist").With(
		slog.String("environment", cfg.App.Environment),
		slog.String("application", "opg-scanning-service"),
	)
	return &Logger{
		cfg:        cfg,
		SlogLogger: slogLogger,
	}
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
