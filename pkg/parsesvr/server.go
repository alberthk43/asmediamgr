package parsesvr

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/services/diskop"
	"asmediamgr/pkg/services/tmdb"

	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
	conf             *Configuration
	doneCh           <-chan struct{}
	shutdownComplete chan struct{}
	wg               sync.WaitGroup
	parsersInfo      []parserInfo
}

func NewParserServer(conf *Configuration) (*ParserServer, error) {
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
	parsersInfo, err := s.initParsers(namedServices)
	if err != nil {
		return fmt.Errorf("failed to initialize parsers: %v", err)
	}
	if len(parsersInfo) == 0 {
		return fmt.Errorf("no parsers found")
	}
	s.parsersInfo = parsersInfo
	s.runMotherDirs()
	return nil
}

func (s *ParserServer) runMotherDirs() {
	for _, motherDir := range s.conf.MotherDirs {
		go s.runMotherDir(motherDir)
	}
}

func (s *ParserServer) runMotherDir(motherDir MontherDir) {
	s.wg.Add(1)
	defer s.wg.Done()
	defer slog.Info("end mother dir loop", slog.String("dir_path", motherDir.DirPath))
	slog.Info("start mother dir loop", slog.String("dir_path", motherDir.DirPath))
	for {
		select {
		case <-s.doneCh:
			return
		default:
			slog.Info("mother dir run", slog.String("dir_path", motherDir.DirPath))
			s.runWithMotherDir(motherDir)
			time.Sleep(motherDir.SleepInterval)
		}
	}
}

func (s *ParserServer) runWithMotherDir(motherDir MontherDir) {
	entries, err := dirinfo.ScanMotherDir(motherDir.DirPath)
	if err != nil {
		slog.Error("failed to scan mother dir", slog.String("dir_path", motherDir.DirPath), slog.String("err", err.Error()))
		return
	}
	_ = entries
}

func (s *ParserServer) initServices() (*namedServices, error) {
	tmdb, err := tmdb.NewTmdbService(filepath.Join(s.conf.ServiceConfDir, "tmdb.toml"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tmdb service: %v", err)
	}
	diskop, err := diskop.NewDiskOpService()
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
		slog.Info("add parsers", slog.String("name", parserName))
	}
	return parsersInfo, nil
}

func (s *ParserServer) WaitForShutdown() {
	<-s.shutdownComplete
}
