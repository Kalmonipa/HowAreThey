package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Define log levels as constants
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// A variable to hold the minimum log level
var minLogLevel = LogLevelInfo

func SetupLogger() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	if os.Getenv("TEST_ENV") == "true" {
		minLogLevel = LogLevelFatal
		return
	}

	envLogLevel := os.Getenv("LOG_LEVEL")

	// Map string values to log level constants
	switch envLogLevel {
	case "DEBUG":
		minLogLevel = LogLevelDebug
	case "INFO":
		minLogLevel = LogLevelInfo
	case "WARN":
		minLogLevel = LogLevelWarn
	case "ERROR":
		minLogLevel = LogLevelError
	case "FATAL":
		minLogLevel = LogLevelFatal
	default:
		minLogLevel = LogLevelInfo
	}
}

// logMessage function that supports formatted strings
func LogMessage(level int, format string, args ...interface{}) {
	ts := time.Now().Format("02/01/2006 15:04:05")
	msg := fmt.Sprintf(format, args...)

	if level < minLogLevel {
		return
	}

	levelStr := ""
	switch level {
	case LogLevelDebug:
		levelStr = "DEBUG"
	case LogLevelInfo:
		levelStr = "INFO"
	case LogLevelWarn:
		levelStr = "WARN"
	case LogLevelError:
		levelStr = "ERROR"
	case LogLevelFatal:
		levelStr = "FATAL"
	}

	log.Printf("%s\t%s\t%s\n", ts, levelStr, msg)
}
