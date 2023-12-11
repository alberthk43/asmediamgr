package server

import (
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/services/diskop"
	"asmediamgr/pkg/services/tmdb"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slog"
)

type Server struct {
	conf             *Configuration
	shutdownComplete chan struct{}
}

func NewServer(conf *Configuration) (*Server, error) {
	return &Server{
		conf: conf,
	}, nil
}

func Run(s *Server) error {
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
	return nil
}

func PrintAndDie(msg string) {
	slog.Error(msg)
	os.Exit(1)
}

type namedServices struct {
	tmdb   *tmdb.TmdbService
	diskOp *diskop.DiskOpService
}

func (s *Server) initServices() (*namedServices, error) {
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

func (s *Server) initParsers(namedServices *namedServices) ([]parserInfo, error) {
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
		parser, err := genFn(parserConfPath, &parser.NamedServices{
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
	}
	return parsersInfo, nil
}

func (s *Server) WaitForShutdown() {
	<-s.shutdownComplete
}
