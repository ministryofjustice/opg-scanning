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

func NewLogger() *Logger {
	slogLogger := telemetry.NewLogger("opg-data-lpa-store/getlist")
	return &Logger{
		cfg:        config.NewConfig(),
		SlogLogger: slogLogger,
	}
}

func (l *Logger) Info(message string, args ...interface{}) {
	logMessage := fmt.Sprintf("INFO: "+message, args...)
	l.SlogLogger.Info(logMessage)
}

func (l *Logger) Error(message string, args ...interface{}) {
	logMessage := fmt.Sprintf("ERROR: "+message, args...)
	l.SlogLogger.Error(logMessage)
}
