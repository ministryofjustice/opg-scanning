package logger

import "log"

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(message string) {
	log.Println("INFO: " + message)
}

func (l *Logger) Error(message string) {
	log.Println("ERROR: " + message)
}
