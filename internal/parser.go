package asmediamgr

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func PrepareLog() error {
	appName := "parser"
	logName := appName + ".log"
	f, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	mw := io.MultiWriter(os.Stderr, f)
	basePath := "C:/Users/caskeep/mycode/project/2023Q4/asmediamgr"
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			if source != nil {
				relPath, err := filepath.Rel(basePath, source.File)
				if err == nil {
					source.File = relPath
				}
			}
		}
		return a
	}
	loggerMw := slog.NewJSONHandler(mw, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace})
	slog.SetDefault(slog.New(loggerMw))
	slog.Info("logging initialized")
	return nil
}
