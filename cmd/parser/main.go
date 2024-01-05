package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	asmediamgr "asmediamgr/internal"
	_ "asmediamgr/internal/builtin"
	"asmediamgr/pkg/config"
	"asmediamgr/pkg/server"
)

const (
	mainFilePath = "/cmd/parser/main.go"
)

func main() {
	flag.Parse()
	if err := asmediamgr.PrepareLog(mainFilePath); err != nil {
		slog.Error("Failed to prepare logging: %s", err)
		os.Exit(1)
	}
	c, err := config.LoadConfigurationFromFile(asmediamgr.Config)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed to load config: %v", err))
	}
	s, err := server.NewParserServer(c)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed to initialize server: %v", err))
	}
	err = server.Run(s)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed to run server: %v", err))
	}
	s.WaitForShutdown()
}
