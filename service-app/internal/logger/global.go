package logger

import (
	"sync"

	"github.com/ministryofjustice/opg-scanning/config"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// returns the singleton global logger.
// It ensures that NewLogger is only called once.
func GetLogger(appConfig *config.Config) *Logger {
	once.Do(func() {
		globalLogger = NewLogger(appConfig)
	})
	return globalLogger
}
