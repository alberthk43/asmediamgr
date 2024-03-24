package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/common"
	"asmediamgr/pkg/common/aslog"
	"asmediamgr/pkg/disk"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/tmdb"

	_ "asmediamgr/pkg/parser/tvepfile"
)

// flagConfig is the configuration for the program
type flagConfig struct {
	configFile           string
	parserConfigDir      string
	loglv                string
	aslogConfig          aslog.Config
	enableParsers        flagStringSlice
	disableParsers       flagStringSlice
	parserDirs           flagStringSlice
	parserTargetMovieDir string
	parserTargetTvDir    string
	parserScanDur        time.Duration
	parserParseDur       time.Duration
	tmdbProxy            string
	dryRun               bool
}

type flagStringSlice []string

func (f *flagStringSlice) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *flagStringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}
	cfg := flagConfig{
		aslogConfig: aslog.Config{},
	}
	flag.StringVar(&cfg.configFile, "config", "config.yaml", "config file")
	flag.StringVar(&cfg.parserConfigDir, "parsercfg", "parsercfg", "parser dir")
	flag.StringVar(&cfg.loglv, "loglv", "info", "log level")
	flag.Var(&cfg.enableParsers, "enable", "enable parsers")
	flag.Var(&cfg.disableParsers, "disable", "disable parsers")
	flag.Var(&cfg.parserDirs, "parserscan", "parser dirs")
	flag.StringVar(&cfg.parserTargetMovieDir, "moviedir", "movies", "target movie dir")
	flag.StringVar(&cfg.parserTargetTvDir, "tvdir", "tv", "target tv dir")
	flag.DurationVar(&cfg.parserScanDur, "scandur", 5*time.Minute, "scan duration")
	flag.DurationVar(&cfg.parserParseDur, "parsedur", 1*time.Second, "parse duration")
	flag.StringVar(&cfg.tmdbProxy, "tmdbproxy", "", "tmdb proxy")
	flag.BoolVar(&cfg.dryRun, "dryrun", false, "dry run")
	flag.Parse()

	loglvVal, err := level.Parse(cfg.loglv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level: %v\n", err)
		os.Exit(1)
	}
	cfg.aslogConfig.Level = &aslog.AllowedLevel{
		LevelOpt: level.Allow(loglvVal),
	}
	logger := aslog.New(&cfg.aslogConfig)

	parserMgrOpts := &parser.ParserMgrOpts{
		Logger:         log.With(logger, "component", "parsermgr"),
		ConfigDir:      cfg.parserConfigDir,
		EnableParsers:  cfg.enableParsers,
		DisableParsers: cfg.disableParsers,
	}
	parserMgr, err := parser.NewParserMgr(parserMgrOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create parser manager: %v\n", err)
		os.Exit(1)
	}

	tmdbService, err := tmdb.NewTmdbService(&tmdb.Configuration{
		Sock5Proxy: cfg.tmdbProxy,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create tmdb service: %v\n", err)
		os.Exit(1)
	}
	parser.RegisterTmdbService(tmdbService)

	diskService, err := disk.NewDiskService(&disk.DiskServiceOpts{
		Logger:         log.With(logger, "component", "disk"),
		DryRunModeOpen: cfg.dryRun,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create disk service: %v\n", err)
		os.Exit(1)
	}
	parser.RegisterDiskService(diskService)

	parserMgrRunOpts := &parser.ParserMgrRunOpts{
		ScanDirs: cfg.parserDirs,
		MediaTypeDirs: map[common.MediaType]string{
			common.MediaTypeMovie: cfg.parserTargetMovieDir,
			common.MediaTypeTv:    cfg.parserTargetTvDir,
		},
		SleepDurScan:  cfg.parserScanDur,
		SleepDurParse: cfg.parserParseDur,
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = parserMgr.RunParsers(parserMgrRunOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run parsers: %v\n", err)
			os.Exit(1)
		}
	}()
	wg.Wait()
}
