package parser

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/service"

	"sync"
)

var (
	parserGenFnMapMu sync.Mutex
	ParserGenFnMap   = map[string]ParserGenFn{}
)

type TmdbService interface {
}

type DiskOpService interface {
}

type NamedServices struct {
	Tmdb   TmdbService
	DiskOp DiskOpService
}

func init() {
	ParserGenFnMap = make(map[string]ParserGenFn)
}

type ParserGenFn func(configPath string, namedServices *NamedServices, services service.ServiceMap) (Parser, error)

func RegisterParserFn(name string, genFn ParserGenFn) {
	if genFn == nil {
		panic("ParserGenFn is nil")
	}
	parserGenFnMapMu.Lock()
	defer parserGenFnMapMu.Unlock()
	if _, ok := ParserGenFnMap[name]; ok {
		panic("ParserGenFn already registered")
	}
	ParserGenFnMap[name] = genFn
}

type Parser interface {
	Parse(entry *dirinfo.Entry) error
	Priority() int
}
