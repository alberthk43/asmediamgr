package parser

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/common"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/disk"
)

// TmdbService is a service that can search tmdb
type TmdbService interface {
	GetSearchMovies(query string, urlOptions map[string]string) (*tmdb.SearchMovies, error)
	GetMovieDetails(id int, urlOptions map[string]string) (*tmdb.MovieDetails, error)
	GetSearchTVShow(query string, urlOptions map[string]string) (*tmdb.SearchTVShows, error)
	GetTVDetails(id int, urlOptions map[string]string) (*tmdb.TVDetails, error)
}

// DiskService is a service that can do real disk operations, such as rename files, etc
type DiskService interface {
	RenameTvEpisode(task *disk.TvEpisodeRenameTask) error
	RenameTvSubtitle(task *disk.TvSubtitleRenameTask) error
	RenameMovie(task *disk.MovieRenameTask) error
	RenameMovieSubtitle(task *disk.MovieSubtitleRenameTask) error
	MoveToTrash(task *disk.MoveToTrashTask) error
}

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
// Note: this function is concurrent safe
func RegisterTmdbService(s TmdbService) {
	tmdbServiceMu.Lock()
	defer tmdbServiceMu.Unlock()
	tmdbServce = s
}

// RegisterTmdbService registers a tmdb service
// Note: this function is concurrent safe
func GetDefaultTmdbService() TmdbService {
	tmdbServiceMu.RLock()
	defer tmdbServiceMu.RUnlock()
	return tmdbServce
}

var (
	diskServiceMu sync.RWMutex
	diskService   DiskService
)

// RegisterDiskService registers a disk service
// Note: this function is concurrent safe
func RegisterDiskService(s DiskService) {
	diskServiceMu.Lock()
	defer diskServiceMu.Unlock()
	diskService = s
}

// GetDefaultDiskService returns the disk service
// Note: this function is concurrent safe
func GetDefaultDiskService() DiskService {
	diskServiceMu.RLock()
	defer diskServiceMu.RUnlock()
	return diskService
}

// Parserable is an interface for parsers
type Parserable interface {
	IsDefaultEnable() bool
	Init(cfgPath string, logger log.Logger) (priority float32, err error)
	Parse(entry *dirinfo.Entry, opts *ParserMgrRunOpts) (ok bool, err error)
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
		_, err := os.Stat(scanDir)
		if err != nil {
			return fmt.Errorf("failed to stat scanDir: %v", err)
		}
		go pm.runParsersWithDir(&wg, scanDir, opts)
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
// TODO need unittest for this function
func (pm *ParserMgr) runParsersWithDir(wg *sync.WaitGroup, scanDir string, opts *ParserMgrRunOpts) {
	defer wg.Done()
	doNextTime := make(map[string]*failNextTime)
	for {
		now := time.Now()
		entries, err := dirinfo.ScanMotherDir(scanDir)
		if err != nil {
			level.Error(pm.logger).Log("msg", fmt.Sprintf("failed to scan motherDir: %v", err))
			time.Sleep(pm.sleepDurScan)
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
				nextTime = &failNextTime{validTime: now, failCnt: 0}
				doNextTime[entry.Name()] = nextTime
			}
			if nextTime.validTime.After(now) {
				continue
			}
			parserName, err := pm.runEntry(entry, opts)
			if err != nil {
				level.Error(pm.logger).Log("msg", "run entry err", "entry", entry.Name(), "err", err)
				time.Sleep(pm.sleepDurScan)
				return
			}
			nextTime.failCnt++
			nextTime.validTime = now.Add(punishAddTime(nextTime.failCnt))
			if parserName != "" {
				level.Info(pm.logger).Log("msg", "entry parser succ", "entry", entry.Name(), "parser", parserName)
			} else {
				level.Warn(pm.logger).Log("msg", "entry parser fail", "entry", entry.Name(), "nextValidTime", nextTime.validTime, "failCnt", nextTime.failCnt)
			}
		}
		time.Sleep(pm.sleepDurScan)
	}
}

func punishAddTime(failCnt int32) time.Duration {
	if failCnt <= 0 {
		return 0
	}
	if failCnt >= 17 {
		return time.Duration(65536) * time.Minute
	}
	return time.Duration(math.Pow(2, float64(failCnt-1))) * time.Minute
}

func (pm *ParserMgr) runEntry(entry *dirinfo.Entry, opts *ParserMgrRunOpts) (okParserName string, err error) {
	// TODO if entry is NOT existed any more, should return "", nil
	for _, parserInfo := range pm.parsers {
		ok, err := pm.runParser(entry, parserInfo, opts)
		if err != nil {
			level.Error(pm.logger).Log("msg", "run parser err", "parser", parserInfo.name, "err", err)
			time.Sleep(pm.sleepDurParse)
			return "", nil
		}
		if ok {
			okParserName = parserInfo.name
			time.Sleep(pm.sleepDurParse)
			break
		}
		time.Sleep(pm.sleepDurParse)
	}
	return okParserName, nil
}

// runParser runs the parser, and will recover all parser logic level panic
func (pm *ParserMgr) runParser(entry *dirinfo.Entry, parserInfo parserInfo, opts *ParserMgrRunOpts) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic: %v\n", r)
			fmt.Printf("stack: %s\n", string(debug.Stack()))
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return parserInfo.parser.Parse(entry, opts)
}
