package aslog

import (
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var (
	timestampFormat = log.TimestampFormat(
		func() time.Time { return time.Now().UTC() },
		"2006-01-02T15:04:05.000Z07:00",
	)
)

type AllowedLevel struct {
	LevelOpt level.Option
}

type AllowedFormat struct {
}

type Config struct {
	Level  *AllowedLevel
	Format *AllowedFormat
}

func New(config *Config) log.Logger {
	return NewWithLogger(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), config)
}

func NewWithLogger(l log.Logger, config *Config) log.Logger {
	if config.Level != nil {
		l = log.With(l, "ts", timestampFormat, "caller", log.Caller(5))
		l = level.NewFilter(l, config.Level.LevelOpt)
	} else {
		l = log.With(l, "ts", timestampFormat, "caller", log.DefaultCaller)
	}
	return l
}
