package server

import "time"

type MontherDir struct {
	DirPath       string
	SleepInterval time.Time
}

type Configuration struct {
	ServiceConfDir string
	ParserConfDir  string
	MotherDirs     []MontherDir
}
