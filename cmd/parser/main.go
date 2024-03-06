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
)

const (
	mainFilePath = "/cmd/parser/main.go"
)

type flagConfig struct {
	aslogConfig *aslog.Config
}

func main() {

	// cfg := flagConfig{
	// 	aslogConfig: &aslog.Config{},
	// }

	// logger := aslog.New(cfg.aslogConfig)

	flag.Parse()
	if err := asmediamgr.PrepareLog(mainFilePath); err != nil {
		fmt.Printf("Failed to prepare logging: %s\n", err)
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
