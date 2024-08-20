package logger

import (
	"fmt"
	"log/slog"
	"os"
	"playlist-manager/pkg/utils"
)

// formatErr Formats the error to a string for logging if it is not already a string
func formatErr(err any) string {
	switch e := err.(type) {
	case string:
		return e
	case error:
		return e.Error()
	default:
		return "Logger: Unknown error type passed to logger"
	}
}

// Init initializes the logger, needs to be called before any logging if you want to use the custom logger
func Init(level string) {
	//Log level parsing
	slogLevel := utils.LevelStringToSlog(level)

	//Open file for logging
	logFile, err := os.OpenFile("log.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Fatal("Logger: Error opening the log file -> " + fmt.Sprint(err))
	}

	//Set the default logger
	slog.SetDefault(slog.New(
		//Log to file in JSON format
		slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: slogLevel,
		}),
	))
}

// DebugSource logs a debug message with the source specified
func DebugSource(err any, source string) {
	slog.Debug(formatErr(err), "source", source)
}

// Debug logs a debug message
func Debug(err any, args ...any) {
	slog.Debug(formatErr(err), args...)
}

// InfoSource logs an info message with the source specified
func InfoSource(err any, source string) {
	slog.Info(formatErr(err), "source", source)
}

// Info logs an info message
func Info(err any, args ...any) {
	slog.Info(formatErr(err), args...)
}

// WarnSource logs a warning message with the source specified
func WarnSource(err any, source string) {
	slog.Warn(formatErr(err), "source", source)
}

// Warn logs a warning message
func Warn(err any, args ...any) {
	slog.Warn(formatErr(err), args...)
}

// ErrorSource logs an error message with the source specified
func ErrorSource(err any, source string) {
	slog.Error(formatErr(err), "source", source)
}

// Error logs an error message
func Error(err any, args ...any) {
	slog.Error(formatErr(err), args...)
}

// FatalSource logs an error message with the source specified and then exits the program using os.Exit(1)
func FatalSource(err any, source string) {
	slog.Error(formatErr(err), "source", source)
	os.Exit(1)
}

// Fatal logs an error message and then exits the program using os.Exit(1)
func Fatal(err any, args ...any) {
	slog.Error(formatErr(err), args...)
	os.Exit(1)
}

// PanicSource logs an error message with the source specified and then panics
func PanicSource(err any, source string) {
	slog.Error(formatErr(err), "source", source)
	panic(formatErr(err))
}

// Panic logs an error message and then panics
func Panic(err any, args ...any) {
	slog.Error(formatErr(err), args...)
	panic(formatErr(err))
}
