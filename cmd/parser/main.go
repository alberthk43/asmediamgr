package main

import (
	asmediamgr "asmediamgr/internal"
	_ "asmediamgr/internal/builtin"
	"asmediamgr/pkg/config"
	"asmediamgr/pkg/parsesvr"

	"flag"
	"fmt"
	"log/slog"
	"os"
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
		parsesvr.PrintAndDie(fmt.Sprintf("Failed to load config: %v", err))
	}
	s, err := parsesvr.NewParserServer(c)
	if err != nil {
		parsesvr.PrintAndDie(fmt.Sprintf("Failed to initialize server: %v", err))
	}
	err = parsesvr.Run(s)
	if err != nil {
		parsesvr.PrintAndDie(fmt.Sprintf("Failed to run server: %v", err))
	}
	s.WaitForShutdown()
}
