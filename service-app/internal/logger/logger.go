package logger

import "log"

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(message string) {
	log.Println("INFO: " + message)
}

func (l *Logger) InfoFormated(message string, args ...interface{}) {
	log.Printf("INFO: "+message, args...)
}

func (l *Logger) Error(message string) {
	log.Println("ERROR: " + message)
}

func (l *Logger) ErrorFormated(message string, args ...interface{}) {
	log.Printf("ERROR: "+message, args...)
}
