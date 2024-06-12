package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"

	"github.com/alberthk43/asmediamgr/pkg/common"
	"github.com/alberthk43/asmediamgr/pkg/common/aslog"
	"github.com/alberthk43/asmediamgr/pkg/disk"
	"github.com/alberthk43/asmediamgr/pkg/parser"
	"github.com/alberthk43/asmediamgr/pkg/stat"
	"github.com/alberthk43/asmediamgr/pkg/tmdb"
	"github.com/alberthk43/asmediamgr/pkg/utils"

	_ "github.com/alberthk43/asmediamgr/pkg/parser/moviedir"
	_ "github.com/alberthk43/asmediamgr/pkg/parser/moviefile"
	_ "github.com/alberthk43/asmediamgr/pkg/parser/tvdir"
	_ "github.com/alberthk43/asmediamgr/pkg/parser/tvepfile"
)

var configFile string

// FlagConfig is the configuration for the program
type FlagConfig struct {
	ParserConfigDir             string          `yaml:"parserConfigDir"`
	Loglv                       string          `yaml:"loglv"`
	aslogConfig                 aslog.Config    `yaml:"aslogConfig"`
	EnableParsers               flagStringSlice `yaml:"enableParsers"`
	DisableParsers              flagStringSlice `yaml:"disableParsers"`
	ScanDirs                    flagStringSlice `yaml:"scanDirs"`
	ParserTargetMovieDir        string          `yaml:"parserTargetMovieDir"`
	ParserTargetTvDir           string          `yaml:"parserTargetTvDir"`
	ParserTargetTrash           string          `yaml:"parserTargetTrash"`
	ParserScanDur               time.Duration   `yaml:"parserScanDur"`
	ParserParseDur              time.Duration   `yaml:"parserParseDur"`
	TmdbProxy                   string          `yaml:"tmdbProxy"`
	TmdbCacheDur                time.Duration   `yaml:"tmdbCacheDur"`
	DryRun                      bool            `yaml:"dryRun"`
	StatInterval                time.Duration   `yaml:"statInterval"`
	StatInitWait                time.Duration   `yaml:"statInitWait"`
	StatMovieDirs               flagStringSlice `yaml:"statMovieDirs"`
	StatTvDirs                  flagStringSlice `yaml:"statTvDirs"`
	StatLargeMovieSize          string          `yaml:"statLargeMovieSize"`
	statLargeMovieSizeBytes     int64           `yaml:"statLargeMovieSizeBytes"`
	StatLargeTvEpisodeSize      string          `yaml:"statLargeTvEpisodeSize"`
	statLargeTvEpisodeSizeBytes int64           `yaml:"statLargeTvEpisodeSizeBytes"`
	EnableStat                  bool            `yaml:"enableStat"`
	EnablePrometheusHTTP        bool            `yaml:"enablePrometheusHTTP"`
	PrometheusPort              int             `yaml:"prometheusPort"`
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
	var err error
	cfg := FlagConfig{
		aslogConfig: aslog.Config{},
	}
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.StringVar(&cfg.ParserConfigDir, "parsercfg", "parsercfg", "parser dir")
	flag.StringVar(&cfg.Loglv, "loglv", "info", "log level")
	flag.Var(&cfg.EnableParsers, "enable", "enable parsers")
	flag.Var(&cfg.DisableParsers, "disable", "disable parsers")
	flag.Var(&cfg.ScanDirs, "scandir", "parser dirs")
	flag.StringVar(&cfg.ParserTargetMovieDir, "movietarget", "movie", "target movie dir")
	flag.StringVar(&cfg.ParserTargetTvDir, "tvtarget", "tv", "target tv dir")
	flag.StringVar(&cfg.ParserTargetTrash, "trash", "trash", "trash dir")
	flag.DurationVar(&cfg.ParserScanDur, "scandur", 5*time.Minute, "scan duration")
	flag.DurationVar(&cfg.ParserParseDur, "parsedur", 1*time.Second, "parse duration")
	flag.StringVar(&cfg.TmdbProxy, "tmdbproxy", "", "tmdb proxy")
	flag.DurationVar(&cfg.TmdbCacheDur, "tmdbcachedur", 6*time.Hour, "tmdb cache duration")
	flag.BoolVar(&cfg.DryRun, "dryrun", false, "dry run")
	flag.DurationVar(&cfg.StatInterval, "statinterval", 6*time.Hour, "stat interval")
	flag.DurationVar(&cfg.StatInitWait, "statinitwait", 10*time.Second, "stat init wait")
	flag.Var(&cfg.StatMovieDirs, "statmoviedir", "stat movie dirs")
	flag.Var(&cfg.StatTvDirs, "stattvdir", "stat tv dirs")
	flag.StringVar(&cfg.StatLargeMovieSize, "statlargemoviesize", "10G", "stat large movie size")
	flag.StringVar(&cfg.StatLargeTvEpisodeSize, "statlargeepisodesize", "5G", "stat large tv episode size")
	flag.BoolVar(&cfg.EnableStat, "stat", true, "enable stat")
	flag.BoolVar(&cfg.EnablePrometheusHTTP, "prometheus", true, "enable prometheus http")
	flag.IntVar(&cfg.PrometheusPort, "prometheusport", 12200, "prometheus port")
	data, err := os.ReadFile(configFile)
	if err != nil {
		printAndDie("failed to read config file: %v\n", err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		printAndDie("failed to decode config file: %v\n", err)
	}
	flag.Parse()

	cfg.StatMovieDirs = append(cfg.StatMovieDirs, cfg.ParserTargetMovieDir)
	cfg.StatTvDirs = append(cfg.StatTvDirs, cfg.ParserTargetTvDir)
	if n, err := utils.SizeStringToBytesNum(cfg.StatLargeMovieSize); err != nil {
		printAndDie("failed to parse large movie size: %v\n", err)
	} else {
		cfg.statLargeMovieSizeBytes = n
	}
	if n, err := utils.SizeStringToBytesNum(cfg.StatLargeTvEpisodeSize); err != nil {
		printAndDie("failed to parse large episode size: %v\n", err)
	} else {
		cfg.statLargeTvEpisodeSizeBytes = n
	}

	loglvVal, err := level.Parse(cfg.Loglv) // TODO test log level is working OR not
	if err != nil {
		printAndDie("failed to parse log level: %v\n", err)
	}
	cfg.aslogConfig.Level = &aslog.AllowedLevel{
		LevelOpt: level.Allow(loglvVal),
	}
	logger := aslog.New(&cfg.aslogConfig)

	parserMgr, err := parser.NewParserMgr(&parser.ParserMgrOpts{
		Logger:         log.With(logger, "component", "parsermgr"),
		ConfigDir:      cfg.ParserConfigDir,
		EnableParsers:  cfg.EnableParsers,
		DisableParsers: cfg.DisableParsers,
	})
	if err != nil {
		printAndDie("failed to create parser manager: %v\n", err)
	}

	if tmdbService, err := tmdb.NewTmdbService(&tmdb.Configuration{
		Logger:        logger,
		Sock5Proxy:    cfg.TmdbProxy,
		ValidCacheDur: cfg.TmdbCacheDur,
	}); err != nil {
		printAndDie("failed to create tmdb service: %v\n", err)
	} else {
		parser.RegisterTmdbService(tmdbService)
	}

	if diskService, err := disk.NewDiskService(&disk.DiskServiceOpts{
		Logger:         log.With(logger, "component", "disk"),
		DryRunModeOpen: cfg.DryRun,
	}); err != nil {
		printAndDie("failed to create disk service: %v\n", err)
	} else {
		parser.RegisterDiskService(diskService)
	}

	var wg sync.WaitGroup
	if cfg.EnableStat {
		statOpts := &stat.StatOpts{
			Logger:             log.With(logger, "component", "stat"),
			Interval:           cfg.StatInterval,
			InitWait:           cfg.StatInitWait,
			MovieDirs:          cfg.StatMovieDirs,
			TvDirs:             cfg.StatTvDirs,
			LargeMovieSize:     cfg.statLargeMovieSizeBytes,
			LargeTvEpisodeSize: cfg.statLargeTvEpisodeSizeBytes,
		}
		stat, err := stat.NewStat(statOpts)
		if err != nil {
			printAndDie("failed to create stat: %v\n", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = stat.Run()
			if err != nil {
				printAndDie("failed to run stat: %v\n", err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = parserMgr.RunParsers(&parser.ParserMgrRunOpts{
			ScanDirs: cfg.ScanDirs,
			MediaTypeDirs: map[common.MediaType]string{
				common.MediaTypeMovie: cfg.ParserTargetMovieDir,
				common.MediaTypeTv:    cfg.ParserTargetTvDir,
				common.MediaTypeTrash: cfg.ParserTargetTrash,
			},
			SleepDurScan:  cfg.ParserScanDur,
			SleepDurParse: cfg.ParserParseDur,
		})
		if err != nil {
			printAndDie("failed to run parsers: %v\n", err)
		}
	}()
	if cfg.EnablePrometheusHTTP {
		go func() {
			initPrometheusHTTP()
			err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.PrometheusPort), nil)
			if err != nil {
				printAndDie("failed to run prometheus http: %v\n", err)
			}
		}()
	}
	wg.Wait()
}

func initPrometheusHTTP() {
	http.Handle("/metrics", promhttp.Handler())
}

func printAndDie(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
