package logger

import (
	"log"
	"os"
)

type Logger struct {
	logger  *log.Logger
	logFile *os.File
}

func Init(path string) *Logger {
	var logger = log.Default()
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err == nil {
		logger = log.New(logFile, "", log.Ldate|log.Ltime)
	} else {
		logger.Printf("ERROR: can't open log file\n       %v", err)
	}
	return &Logger{logger: logger, logFile: logFile}
}

func (log *Logger) Close() {
	log.logFile.Close()
}

func (log *Logger) Info(msg string) {
	log.logger.Printf("INFO: " + msg)
}

func (log *Logger) Error(msg string, err error) {
	log.logger.Printf("ERROR: %s\n       %v", msg, err)
}
