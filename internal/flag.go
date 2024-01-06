package asmediamgr

import "flag"

var (
	Config  string
	LogPath string
)

func init() {
	flag.StringVar(&Config, "config", "config.toml", "config file path")
	flag.StringVar(&LogPath, "logpath", "asmediamgr.log", "log file path")
}
