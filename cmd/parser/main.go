package main

import (
	asmediamgr "asmediamgr/internal"
	"asmediamgr/pkg/server"

	"log/slog"
	"os"
)

const (
	mainFilePath = "/cmd/parser/main.go"
)

func main() {
	if err := asmediamgr.PrepareLog(mainFilePath); err != nil {
		slog.Error("Failed to prepare logging: %s", err)
		os.Exit(1)
	}
	s, err := server.NewServer(&server.Configuration{})
	if err != nil {
		server.PrintAndDie(err.Error())
	}
	err = server.Run(s)
	if err != nil {
		server.PrintAndDie(err.Error())
	}
	s.WaitForShutdown()
}
