package server

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"asmediamgr/pkg/config"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/prometric"
	"asmediamgr/pkg/tmdb"
)

func PrintAndDie(msg string) {
	slog.Error(msg)
	os.Exit(1)
}

type namedServices struct {
	tmdb   *tmdb.TmdbService
	diskOp *diskop.DiskOpService
}

type ParserServer struct {
	conf             *config.Configuration
	doneCh           <-chan struct{}
	shutdownComplete chan struct{}
	wg               sync.WaitGroup
	parsersInfo      []parserInfo
}

func NewParserServer(conf *config.Configuration) (*ParserServer, error) {
	return &ParserServer{
		conf:             conf,
		doneCh:           make(chan struct{}),
		shutdownComplete: make(chan struct{}),
	}, nil
}

func Run(s *ParserServer) error {
	if len(s.conf.MotherDirs) == 0 {
		return fmt.Errorf("no mother dirs found")
	}
	namedServices, err := s.initServices()
	if err != nil {
		return fmt.Errorf("failed to initialize services: %v", err)
	}
	parsersInfoSlice, err := s.initParsers(namedServices)
	if err != nil {
		return fmt.Errorf("failed to initialize parsers: %v", err)
	}
	if len(parsersInfoSlice) == 0 {
		return fmt.Errorf("no parsers found")
	}
	sorted := sortParserInfo(parsersInfoSlice)
	s.parsersInfo = sorted
	s.runProMetrics()
	s.runMotherDirs()
	return nil
}

func (s *ParserServer) runProMetrics() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9999", nil)
	}()
}

func (s *ParserServer) runMotherDirs() {
	for _, motherDir := range s.conf.MotherDirs {
		go s.runMotherDir(motherDir)
	}
}

func (s *ParserServer) runMotherDir(motherDir config.MontherDir) {
	s.wg.Add(1)
	defer s.wg.Done()
	prometric.CurMontherDirAdd()
	defer prometric.CurMontherDirDec()
	defer slog.Info("end mother dir loop", slog.String("dir_path", motherDir.DirPath))
	slog.Info("start mother dir loop", slog.String("dir_path", motherDir.DirPath))
	retryConMap := make(map[string]*retryControl)
	ticker := time.NewTicker(motherDir.SleepInterval)
	defer ticker.Stop()
	s.runWithMotherDir(motherDir, retryConMap)
	for {
		select {
		case <-s.doneCh:
			return
		case <-ticker.C:
			s.runWithMotherDir(motherDir, retryConMap)
		}
	}
}

type retryControl struct {
	visited  bool
	n        int
	nextTime time.Time
}

// runWithMotherDir impls with repeated error protection, 2**n try interval and retry
func (s *ParserServer) runWithMotherDir(motherDir config.MontherDir, retryConMap map[string]*retryControl) {
	slog.Info("mother dir run", slog.String("dir_path", motherDir.DirPath))
	prometric.LoopMontherDirInc()
	entries, err := dirinfo.ScanMotherDir(motherDir.DirPath)
	if err != nil {
		slog.Error("failed to scan mother dir", slog.String("dir_path", motherDir.DirPath), slog.String("err", err.Error()))
		return
	}
	for _, retryCon := range retryConMap {
		retryCon.visited = false
	}
	now := time.Now()
	for _, entry := range entries {
		name := getEntrySpecificName(entry)
		if retryCon, ok := retryConMap[name]; !ok {
			retryConMap[name] = &retryControl{
				visited:  true,
				n:        6, // 64 seconds later
				nextTime: now,
			}
		} else {
			retryCon.visited = true
		}
	}
	for k, retryCon := range retryConMap {
		if !retryCon.visited {
			delete(retryConMap, k)
		}
	}
	for _, entry := range entries {
		name := getEntrySpecificName(entry)
		retryCon, ok := retryConMap[name]
		if !ok {
			slog.Error("entry map not found", slog.String("dir_path", motherDir.DirPath))
			return
		}
		var err error
		if now.Sub(retryCon.nextTime) >= 0 {
			err = s.runWithEntry(entry)
			time.Sleep(time.Second)
		}
		retryCon.n++
		nextTime := nextRetryTime(retryCon.n, now)
		retryCon.nextTime = nextTime
		if err != nil {
			slog.Info("failed to parse entry", slog.String("entry", name),
				slog.String("err", err.Error()),
				slog.String("next_time", nextTime.String()),
			)
		}
	}
}

func nextRetryTime(n int, now time.Time) time.Time {
	const maxRetryInterval = 60 * 60 * time.Second // 1 hour
	if n <= 0 {
		return now.Add(maxRetryInterval)
	}
	if n > 18 {
		return now.Add(maxRetryInterval)
	}
	n = int(math.Exp2(float64(n)))
	return now.Add(time.Duration(n) * time.Second)
}

func getEntrySpecificName(entry *dirinfo.Entry) string {
	if len(entry.FileList) == 0 {
		return ""
	}
	return entry.FileList[0].Name
}

func (s *ParserServer) runWithEntry(entry *dirinfo.Entry) error {
	prometric.EntryInc()
	name := getEntrySpecificName(entry)
	for _, parserInfo := range s.parsersInfo {
		prometric.ParserInc()
		prometric.TemplateParserInc(parserInfo.template)
		if err := parserInfo.parser.Parse(entry); err != nil {
			slog.Info("failed to parse entry", slog.String("parser", parserInfo.info()), slog.String("entry", name), slog.String("err", err.Error()))
		} else {
			slog.Info("succ to parse entry", slog.String("parser", parserInfo.info()), slog.String("entry", name))
			prometric.ParserSuccInc()
			return nil
		}
	}
	return fmt.Errorf("no parser found for entry: %s", name)
}

func (s *ParserServer) initServices() (*namedServices, error) {
	tmdbConf := tmdb.Configuration{
		Sock5Proxy: s.conf.TmdbSock5Proxy,
	}
	tmdb, err := tmdb.NewTmdbService(&tmdbConf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tmdb service: %v", err)
	}
	destPathMap := map[diskop.DestType]string{
		diskop.OnAirTv:    s.conf.DestTvOnAirDir,
		diskop.OnAirMovie: s.conf.DestMovieOnAirDir,
	}
	diskop, err := diskop.NewDiskOpService(destPathMap)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize diskop service: %v", err)
	}
	namedServices := &namedServices{
		tmdb:   tmdb,
		diskOp: diskop,
	}
	return namedServices, nil
}

const (
	parserTemplateIdx = iota
	parserNameIdx
	groupLen
)

type parserInfo struct {
	name     string
	template string
	parser   parser.Parser
	priority int
}

func (s *parserInfo) info() string {
	return fmt.Sprintf(
		"%s,%s(p:%d)", s.name, s.template, s.priority,
	)
}

func (s *ParserServer) initParsers(namedServices *namedServices) ([]parserInfo, error) {
	dir, err := os.Open(s.conf.ParserConfDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open parser config directory: %v", err)
	}
	defer dir.Close()
	subs, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read parser config directory: %v", err)
	}
	parserGenFnMap := parser.ParserGenFnMap
	var parsersInfo []parserInfo
	for _, sub := range subs {
		if sub.IsDir() {
			continue
		}
		if filepath.Ext(sub.Name()) != ".toml" {
			continue
		}
		fileNameWithoutExt := strings.TrimSuffix(sub.Name(), filepath.Ext(sub.Name()))
		groups := strings.Split(fileNameWithoutExt, ",")
		if len(groups) != groupLen {
			return nil, fmt.Errorf("invalid parser config file name: %s", sub.Name())
		}
		templateName := groups[0]
		parserName := groups[1]
		genFn, ok := parserGenFnMap[templateName]
		if !ok {
			return nil, fmt.Errorf("invalid parser config file name: %s", sub.Name())
		}
		parserConfPath := filepath.Join(s.conf.ParserConfDir, sub.Name())
		parser, err := genFn(parserConfPath, &parser.CommonServices{
			Tmdb:   namedServices.tmdb,
			DiskOp: namedServices.diskOp,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize parser: %s, err: %v", sub.Name(), err)
		}
		parsersInfo = append(parsersInfo, parserInfo{
			name:     parserName,
			template: templateName,
			parser:   parser,
			priority: parser.Priority(),
		})
		slog.Info("add parsers", slog.String("name", parserName), slog.String("template", templateName))
	}
	return parsersInfo, nil
}

type parserInfoSlice []parserInfo

func (s parserInfoSlice) Len() int {
	return len(s)
}

func (s parserInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s parserInfoSlice) Less(i, j int) bool {
	return s[i].priority < s[j].priority
}

func sortParserInfo(parsersInfo parserInfoSlice) []parserInfo {
	sort.Sort(parsersInfo)
	return parsersInfo
}

func (s *ParserServer) WaitForShutdown() {
	<-s.shutdownComplete
}
