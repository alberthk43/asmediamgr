package main

import (
	asmediamgr "asmediamgr/internal"
	"asmediamgr/pkg/server"
	"time"

	"log/slog"
	"os"
)

func main() {
	if err := asmediamgr.PrepareLog(); err != nil {
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
	asmediamgr.WaitLogFileAllWritten()
	time.Sleep(10 * time.Second)
}
