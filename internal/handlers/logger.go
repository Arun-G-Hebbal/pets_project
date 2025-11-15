package handlers

import (
	"log"
	"os"
)

// Custom logger with levels
var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

// InitLogger initializes loggers with different prefixes and outputs
func InitLogger() {
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info logs general informational messages
func Info(message string, args ...interface{}) {
	if infoLogger == nil {
		InitLogger()
	}
	infoLogger.Printf(message, args...)
}

// Warn logs warning messages
func Warn(message string, args ...interface{}) {
	if warnLogger == nil {
		InitLogger()
	}
	warnLogger.Printf(message, args...)
}

// Error logs error messages
func Error(message string, args ...interface{}) {
	if errorLogger == nil {
		InitLogger()
	}
	errorLogger.Printf(message, args...)
}
