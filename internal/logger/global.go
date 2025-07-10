package logger

import (
	"sync"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// returns the singleton global logger.
// It ensures that NewLogger is only called once.
func GetLogger(environment string) *Logger {
	once.Do(func() {
		globalLogger = newLogger(environment)
	})
	return globalLogger
}
