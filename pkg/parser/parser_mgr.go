package parser

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/common"
	"asmediamgr/pkg/dirinfo"
)

var (
	// RegisteredParsers is a map of registered parsers
	RegisteredParsers = make(map[string]Parserable)
)

// RegisterParser registers a parser, panic if parser already registered
// Note: this function is not thread safe
func RegisterParser(name string, p Parserable) {
	if _, ok := RegisteredParsers[name]; ok {
		panic("Parser already registered")
	}
	RegisteredParsers[name] = p
}

var (
	tmdbServiceMu sync.RWMutex
	tmdbServce    TmdbService
)

// RegisterTmdbService registers a tmdb service
func RegisterTmdbService(s TmdbService) {
	tmdbServiceMu.Lock()
	defer tmdbServiceMu.Unlock()
	tmdbServce = s
}

// RegisterTmdbService registers a tmdb service
// Note: this function is thread safe
func GetDefaultTmdbService() TmdbService {
	tmdbServiceMu.RLock()
	defer tmdbServiceMu.RUnlock()
	return tmdbServce
}

// Parserable is an interface for parsers
type Parserable interface {
	IsDefaultEnable() bool
	Init(cfgPath string, logger log.Logger) (priority float32, err error)
	Parse(entry *dirinfo.Entry) (ok bool, err error)
}

// ParserMgrOpts is the options for the parser
type ParserMgrOpts struct {
	Logger         log.Logger // logger
	ConfigDir      string     // config tomls dir
	EnableParsers  []string   // enable parsers
	DisableParsers []string   // disable parsers, higher priority than enable parsers
}

// NewParserMgr creates a new parser
// disable parsers have higher priority than enable parsers
// if a parser is not in disable or enable list, it will be enabled if it is default enable
// parse will call Init() func, cfgPath can be optional needed for individual parser
func NewParserMgr(opts *ParserMgrOpts) (*ParserMgr, error) {
	disableMap := make(map[string]struct{})
	enabeMap := make(map[string]struct{})
	enableParsers := make(map[string]Parserable)
	for _, name := range opts.DisableParsers {
		disableMap[name] = struct{}{}
	}
	for _, name := range opts.EnableParsers {
		if _, ok := RegisteredParsers[name]; !ok {
			return nil, fmt.Errorf("parser %s not registered", name)
		}
		enabeMap[name] = struct{}{}
	}
	for name, parser := range RegisteredParsers {
		if _, ok := disableMap[name]; ok {
			continue
		}
		isDefulEnable := parser.IsDefaultEnable()
		_, ok := enabeMap[name]
		if isDefulEnable || ok {
			enableParsers[name] = parser
		}
	}
	pm := &ParserMgr{
		logger: opts.Logger,
	}
	for name, parser := range enableParsers {
		cfgPath := filepath.Join(opts.ConfigDir, name+".toml")
		priority, err := parser.Init(cfgPath, log.With(pm.logger, "parser", name))
		if err != nil {
			return nil, fmt.Errorf("parser %s init error: %v", name, err)
		}
		parserInfo := parserInfo{
			name:     name,
			priority: priority,
			parser:   parser,
		}
		pm.parsers = append(pm.parsers, parserInfo)
	}
	if len(pm.parsers) == 0 {
		return nil, fmt.Errorf("no parser enabled")
	}
	sort.Sort(byPriority(pm.parsers))
	return pm, nil
}

type byPriority []parserInfo

func (a byPriority) Len() int { return len(a) }

func (a byPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byPriority) Less(i, j int) bool { return a[i].priority < a[j].priority }

// parserInfo is a struct that holds the parser info
type parserInfo struct {
	name     string
	priority float32
	parser   Parserable
}

// ParserMgr is a struct that holds the parser
type ParserMgr struct {
	logger        log.Logger
	parsers       []parserInfo
	sleepDurScan  time.Duration
	sleepDurParse time.Duration
}

// ParserMgrRunOpts is the runtime options for the parser
type ParserMgrRunOpts struct {
	ScanDirs      []string
	MediaTypeDirs map[common.MediaType]string
	SleepDurScan  time.Duration
	SleepDurParse time.Duration
}

const (
	defaultScanSleepDur  = time.Duration(5) * time.Minute // default sleep duration for scanning
	defaultParseSleepDur = time.Duration(1) * time.Second // default sleep duration for parsing
)

// RunParsers runs the parsers with the options, maybe in multiple dirs, with multiple goroutines
func (pm *ParserMgr) RunParsers(opts *ParserMgrRunOpts) error {
	if opts.SleepDurScan == 0 {
		pm.sleepDurScan = defaultScanSleepDur
	} else {
		pm.sleepDurScan = opts.SleepDurScan
	}
	if opts.SleepDurParse == 0 {
		pm.sleepDurParse = defaultParseSleepDur
	} else {
		pm.sleepDurParse = opts.SleepDurParse
	}
	if len(opts.ScanDirs) == 0 {
		return fmt.Errorf("no scan dirs")
	}
	var wg sync.WaitGroup
	for _, scanDir := range opts.ScanDirs {
		wg.Add(1)
		err := pm.runParsersWithDir(&wg, scanDir)
		if err != nil {
			return fmt.Errorf("runParsersWithDir() error: %v", err)
		}
	}
	wg.Wait()
	return nil
}

// failNextTime is a struct that holds the next time to run the parser
// prevent the parser from running too frequently
type failNextTime struct {
	validTime time.Time // valid to run time for entry
	failCnt   int32     // fail count
}

// runParsersWithDir runs the parsers with the dir
func (pm *ParserMgr) runParsersWithDir(wg *sync.WaitGroup, scanDir string) error {
	_, err := os.Stat(scanDir)
	if err != nil {
		return fmt.Errorf("failed to stat scanDir: %v", err)
	}
	go func() {
		defer wg.Done()
		doNextTime := make(map[string]*failNextTime)
		now := time.Now()
		for {
			defer time.Sleep(pm.sleepDurScan)
			entries, err := dirinfo.ScanMotherDir(scanDir)
			if err != nil {
				level.Error(pm.logger).Log("msg", fmt.Sprintf("failed to scan motherDir: %v", err))
				break
			}
			entriesMap := make(map[string]struct{})
			for _, entry := range entries {
				entriesMap[entry.Name()] = struct{}{}
			}
			for entryName := range doNextTime {
				if _, ok := entriesMap[entryName]; !ok {
					delete(doNextTime, entryName)
				}
			}
			for _, entry := range entries {
				nextTime, ok := doNextTime[entry.Name()]
				if !ok {
					nextTime := &failNextTime{
						validTime: now,
						failCnt:   0,
					}
					doNextTime[entry.Name()] = nextTime
				} else {
					nextTime.failCnt++
					nextTime.validTime = nextTime.validTime.Add(
						time.Duration(math.Pow(2, float64(nextTime.failCnt))) * time.Second)
				}
				for _, parserInfo := range pm.parsers {
					defer time.Sleep(pm.sleepDurParse)
					ok, err := pm.runParser(entry, parserInfo)
					if err != nil {
						level.Error(pm.logger).Log("msg", fmt.Sprintf("parser %s runParser() error: %v", parserInfo.name, err))
						break
					}
					if ok {
						level.Info(pm.logger).Log("msg", fmt.Sprintf("parser %s runParser() success", parserInfo.name))
						break
					}
				}
				level.Info(pm.logger).Log("msg", fmt.Sprintf("entry %s all parsers failed, will retry after %v", entry.Name(),
					nextTime.validTime.Add(time.Duration(math.Pow(2, float64(nextTime.failCnt)))*time.Second)))
			}
		}
	}()
	return nil
}

// runParser runs the parser, and will recover all parser logic level panic
func (pm *ParserMgr) runParser(entry *dirinfo.Entry, parserInfo parserInfo) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parser %s panic: %v", parserInfo.name, r)
		}
	}()
	ok, err = parserInfo.parser.Parse(entry)
	if err != nil {
		return false, fmt.Errorf("parser %s Parse() error: %v", parserInfo.name, err)
	}
	return ok, nil
}
