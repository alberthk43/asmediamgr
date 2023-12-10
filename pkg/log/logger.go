package logger

import (
	"log/slog"
)

func GetLogger() *slog.Logger {
	return slog.Default()
}
