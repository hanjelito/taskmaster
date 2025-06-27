package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	file *os.File
	log  *log.Logger
}

func New(filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := log.New(file, "", 0)

	return &Logger{
		file: file,
		log:  logger,
	}, nil
}

func (l *Logger) logf(level string, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.log.Printf("[%s] %s: %s", timestamp, level, message)

	// También mostrar en consola para debugging
	fmt.Printf("[%s] %s: %s\n", timestamp, level, message)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.logf("INFO", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.logf("ERROR", format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logf("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) Close() error {
	return l.file.Close()
}
