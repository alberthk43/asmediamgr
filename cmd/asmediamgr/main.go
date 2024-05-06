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
	"asmediamgr/pkg/stat"
	"asmediamgr/pkg/tmdb"
	"asmediamgr/pkg/utils"

	_ "asmediamgr/pkg/parser/moviedir"
	_ "asmediamgr/pkg/parser/moviefile"
	_ "asmediamgr/pkg/parser/tvdir"
	_ "asmediamgr/pkg/parser/tvepfile"
)

// flagConfig is the configuration for the program
type flagConfig struct {
	configFile                  string
	parserConfigDir             string
	loglv                       string
	aslogConfig                 aslog.Config
	enableParsers               flagStringSlice
	disableParsers              flagStringSlice
	parserDirs                  flagStringSlice
	parserTargetMovieDir        string
	parserTargetTvDir           string
	parserTargetTrash           string
	parserScanDur               time.Duration
	parserParseDur              time.Duration
	tmdbProxy                   string
	tmdbCacheDur                time.Duration
	dryRun                      bool
	statInterval                time.Duration
	statInitWait                time.Duration
	statMovieDirs               flagStringSlice
	statTvDirs                  flagStringSlice
	statLargeMovieSize          string
	statLargeMovieSizeBytes     int64
	statLargeTvEpisodeSize      string
	statLargeTvEpisodeSizeBytes int64
	enableStat                  bool
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
	flag.Var(&cfg.parserDirs, "scandir", "parser dirs")
	flag.StringVar(&cfg.parserTargetMovieDir, "movietarget", "movies", "target movie dir")
	flag.StringVar(&cfg.parserTargetTvDir, "tvtarget", "tv", "target tv dir")
	flag.StringVar(&cfg.parserTargetTrash, "trash", "trash", "trash dir")
	flag.DurationVar(&cfg.parserScanDur, "scandur", 5*time.Minute, "scan duration")
	flag.DurationVar(&cfg.parserParseDur, "parsedur", 1*time.Second, "parse duration")
	flag.StringVar(&cfg.tmdbProxy, "tmdbproxy", "", "tmdb proxy")
	flag.DurationVar(&cfg.tmdbCacheDur, "tmdbcachedur", 6*time.Hour, "tmdb cache duration")
	flag.BoolVar(&cfg.dryRun, "dryrun", false, "dry run")
	flag.DurationVar(&cfg.statInterval, "statinterval", 6*time.Hour, "stat interval")
	flag.DurationVar(&cfg.statInitWait, "statinitwait", 10*time.Second, "stat init wait")
	flag.Var(&cfg.statMovieDirs, "statmoviedir", "stat movie dirs")
	flag.Var(&cfg.statTvDirs, "stattvdir", "stat tv dirs")
	flag.StringVar(&cfg.statLargeMovieSize, "statlargemoviesize", "10G", "stat large movie size")
	flag.StringVar(&cfg.statLargeTvEpisodeSize, "statlargeepisodesize", "5G", "stat large tv episode size")
	flag.BoolVar(&cfg.enableStat, "stat", true, "enable stat")
	flag.Parse()

	cfg.statMovieDirs = append(cfg.statMovieDirs, cfg.parserTargetMovieDir)
	cfg.statTvDirs = append(cfg.statTvDirs, cfg.parserTargetTvDir)
	if n, err := utils.SizeStringToBytesNum(cfg.statLargeMovieSize); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse large movie size: %v\n", err)
		os.Exit(1)
	} else {
		cfg.statLargeMovieSizeBytes = n
	}
	if n, err := utils.SizeStringToBytesNum(cfg.statLargeTvEpisodeSize); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse large episode size: %v\n", err)
		os.Exit(1)
	} else {
		cfg.statLargeTvEpisodeSizeBytes = n
	}

	loglvVal, err := level.Parse(cfg.loglv) // TODO test log level is working OR not
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level: %v\n", err)
		os.Exit(1)
	}
	cfg.aslogConfig.Level = &aslog.AllowedLevel{
		LevelOpt: level.Allow(loglvVal),
	}
	logger := aslog.New(&cfg.aslogConfig)

	parserMgr, err := parser.NewParserMgr(&parser.ParserMgrOpts{
		Logger:         log.With(logger, "component", "parsermgr"),
		ConfigDir:      cfg.parserConfigDir,
		EnableParsers:  cfg.enableParsers,
		DisableParsers: cfg.disableParsers,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create parser manager: %v\n", err)
		os.Exit(1)
	}

	if tmdbService, err := tmdb.NewTmdbService(&tmdb.Configuration{
		Logger:        logger,
		Sock5Proxy:    cfg.tmdbProxy,
		ValidCacheDur: cfg.tmdbCacheDur,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create tmdb service: %v\n", err)
		os.Exit(1)
	} else {
		parser.RegisterTmdbService(tmdbService)
	}

	if diskService, err := disk.NewDiskService(&disk.DiskServiceOpts{
		Logger:         log.With(logger, "component", "disk"),
		DryRunModeOpen: cfg.dryRun,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create disk service: %v\n", err)
		os.Exit(1)
	} else {
		parser.RegisterDiskService(diskService)
	}

	parserMgrRunOpts := &parser.ParserMgrRunOpts{
		ScanDirs: cfg.parserDirs,
		MediaTypeDirs: map[common.MediaType]string{
			common.MediaTypeMovie: cfg.parserTargetMovieDir,
			common.MediaTypeTv:    cfg.parserTargetTvDir,
			common.MediaTypeTrash: cfg.parserTargetTrash,
		},
		SleepDurScan:  cfg.parserScanDur,
		SleepDurParse: cfg.parserParseDur,
	}

	var wg sync.WaitGroup
	if cfg.enableStat {
		statOpts := &stat.StatOpts{
			Logger:             log.With(logger, "component", "stat"),
			Interval:           cfg.statInterval,
			InitWait:           cfg.statInitWait,
			MovieDirs:          cfg.statMovieDirs,
			TvDirs:             cfg.statTvDirs,
			LargeMovieSize:     cfg.statLargeMovieSizeBytes,
			LargeTvEpisodeSize: cfg.statLargeTvEpisodeSizeBytes,
		}
		stat, err := stat.NewStat(statOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create stat: %v\n", err)
			os.Exit(1)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = stat.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to run stat: %v\n", err)
				os.Exit(1)
			}
		}()
	}
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
