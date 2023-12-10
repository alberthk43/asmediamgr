package main

import (
	asmediamgr "asmediamgr/internal"
	"log/slog"
	"os"
)

func main() {
	err := asmediamgr.PrepareLog()
	if err != nil {
		slog.Error("Failed to open log file: %v", err)
		os.Exit(1)
	}
	slog.Info("Asmediamgr started!")
}
