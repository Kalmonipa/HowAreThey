package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

var minLogLevel = LogLevelInfo
var fileLogger *log.Logger

func SetupLogger() {
	log.SetFlags(0)

	// Configure stdout logger
	log.SetOutput(os.Stdout)

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

	// Open log file
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	fileLogger = log.New(logFile, "", log.LstdFlags)
}

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

	// Log to stdout
	log.Printf("%s\t%s\t%s\n", ts, levelStr, msg)

	// Log to file
	fileLogger.Printf("%s\t%s\t%s\n", ts, levelStr, msg)
}
