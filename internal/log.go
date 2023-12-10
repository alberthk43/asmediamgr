package asmediamgr

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func PrepareLog() error {
	loggerStd := slog.NewJSONHandler(os.Stderr, nil)
	slog.SetDefault(slog.New(loggerStd))
	logFile, err := os.Open(LogPath)
	if err != nil {
		logFile, err = os.Create(LogPath)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
	}
	mw := io.MultiWriter(os.Stderr, logFile)
	loggerMw := slog.NewJSONHandler(mw, &slog.HandlerOptions{AddSource: true})
	slog.SetDefault(slog.New(loggerMw))
	return nil
}
