package asmediamgr

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func PrepareLog(mainFilePath string) error {
	appName := "parser"
	logName := appName + ".log"
	logFile, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	mw := io.MultiWriter(os.Stderr, logFile)
	_, filename, _, _ := runtime.Caller(1)
	basePath := strings.TrimSuffix(filename, mainFilePath)
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source := a.Value.Any().(*slog.Source); source != nil {
				if relPath, err := filepath.Rel(basePath, source.File); err == nil {
					source.File = filepath.ToSlash(relPath)
				}
			}
		}
		return a
	}
	loggerMw := slog.NewJSONHandler(mw, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace})
	slog.SetDefault(slog.New(loggerMw))
	slog.Info("Logging initialized")
	return nil
}
