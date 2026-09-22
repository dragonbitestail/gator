package ilogger

import (
	"log/slog"
	"os"
	"strings"
)


var LogLevel slog.Level =  getLogLevelFromEnv()
var handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: LogLevel,
		//AddSource: true,
})

//var logr = slog.New(handler)


func GetLogger() *slog.Logger {

	return slog.New(handler)
}

func getLogLevelFromEnv() slog.Level {
    levelStr := os.Getenv("LOG_LEVEL")

    switch strings.ToLower(levelStr) {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelWarn
    }
}
