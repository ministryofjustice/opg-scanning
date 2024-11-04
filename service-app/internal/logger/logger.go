package logger

import (
	"log"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Logger struct {
	cfg         *config.Config
	WriteToFile func(fileName string, message string, path string)
}

func NewLogger() *Logger {
	return &Logger{
		cfg:         config.NewConfig(),
		WriteToFile: util.WriteToFile,
	}
}

func (l *Logger) Logger(fileName, message string) {
	if l.cfg.Log.Level != "debug" {
		return
	}
	projectRoot, err := util.GetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	l.WriteToFile(fileName, message, projectRoot+"/"+l.cfg.ProjectPath+"/logs/")
}

func (l *Logger) Info(message string, args ...interface{}) {
	log.Printf("INFO: "+message, args...)
	l.Logger("InfoLog", message)
}

func (l *Logger) Error(message string, args ...interface{}) {
	log.Printf("ERROR: "+message, args...)
	l.Logger("ErrorLog", message)
}
