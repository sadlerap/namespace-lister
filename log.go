package main

import (
	"log/slog"
	"os"
	"strconv"
)

// buildLogger constructs a new instance of the logger
func buildLogger() *slog.Logger {
	logLevel := getLogLevel()

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}

// getLogLevel fetches the log level from the appropriate environment variable
func getLogLevel() slog.Level {
	env := os.Getenv(EnvLogLevel)
	level, err := strconv.Atoi(env)
	if err != nil {
		return slog.LevelError
	}
	return slog.Level(level)
}
