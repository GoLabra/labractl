package log

import (
	"fmt"
	"log/slog"
	"os"
)

var Logger *slog.Logger

// Init configures the logger with the provided debug flag.
func Init(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

func Debugf(format string, args ...any) {
	if Logger == nil {
		Init(false)
	}
	Logger.Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	if Logger == nil {
		Init(false)
	}
	Logger.Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	if Logger == nil {
		Init(false)
	}
	Logger.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	if Logger == nil {
		Init(false)
	}
	Logger.Error(fmt.Sprintf(format, args...))
}
