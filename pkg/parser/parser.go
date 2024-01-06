package parser

import (
	"sync"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
	"asmediamgr/pkg/servicemgr"
)

var (
	parserGenFnMapMu sync.Mutex
	ParserGenFnMap   = make(map[string]ParserGenFn)
)

// TmdbService is a service that can search tmdb
type TmdbService interface {
	GetMovieDetails(id int, urlOptions map[string]string) (*tmdb.MovieDetails, error)
	GetSearchMovies(query string, urlOptions map[string]string) (*tmdb.SearchMovies, error)
	GetSearchTVShow(query string, urlOptions map[string]string) (*tmdb.SearchTVShows, error)
	GetTVDetails(id int, urlOptions map[string]string) (*tmdb.TVDetails, error)
}

// DiskOpService is a service that can rename files
type DiskOpService interface {
	RenameSingleTvEpFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails, season int, episode int, destType diskop.DestType) error
	RenameSingleMovieFile(entry *dirinfo.Entry, old *dirinfo.File, movieDetail *tmdb.MovieDetails, destType diskop.DestType) error
	RenameMovieSubtiles(entry *dirinfo.Entry, filesMap map[string][]*dirinfo.File, movieDetail *tmdb.MovieDetails, destType diskop.DestType) error
	DelDirEntry(entry *dirinfo.Entry) error
	RenameTvMusicFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails, name string, destType diskop.DestType) error
}

type CommonServices struct {
	Tmdb   TmdbService
	DiskOp DiskOpService
}

type ParserGenFn func(configPath string, namedServices *CommonServices, services servicemgr.ServiceMap) (Parser, error)

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

type DefaultPriority struct{}

func (*DefaultPriority) Priority() int { return 0 }
