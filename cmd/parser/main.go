package main

import (
	"flag"
	"fmt"
	"os"

	asmediamgr "asmediamgr/internal"
	_ "asmediamgr/internal/builtin"
	"asmediamgr/pkg/common/aslog"
	"asmediamgr/pkg/config"
	"asmediamgr/pkg/server"

	"github.com/go-kit/log/level"
)

const (
	mainFilePath = "/cmd/parser/main.go"
)

// type flagConfig struct {
// 	aslogConfig *aslog.Config
// }

func main() {
	flag.Parse()

	// some preparation
	aslogConfig := &aslog.Config{
		Level: &aslog.AllowedLevel{
			LevelOpt: level.AllowInfo(),
		},
		Format: &aslog.AllowedFormat{},
	}
	logger := aslog.New(aslogConfig)

	// run program
	if err := asmediamgr.PrepareLog(mainFilePath); err != nil {
		fmt.Printf("Failed to prepare logging: %s\n", err)
		os.Exit(1)
	}
	c, err := config.LoadConfigurationFromFile(asmediamgr.Config)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed to load config: %v", err))
	}
	s, err := server.NewParserServer(c, logger)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed to initialize server: %v", err))
	}
	err = server.Run(s)
	if err != nil {
		server.PrintAndDie(fmt.Sprintf("Failed to run server: %v", err))
	}
	s.WaitForShutdown()
}
