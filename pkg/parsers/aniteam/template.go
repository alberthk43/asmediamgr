package tvepfile

import (
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/service"
)

const (
	templateName = "tvepfile"
)

func gen(configPath string, namedServices *parser.CommonServices, services service.ServiceMap) (parser.Parser, error) {
	parser := &TvEpParser{
		tmdbService:   namedServices.Tmdb,
		distOpService: namedServices.DiskOp,
	}
	return parser, nil
}

func init() {
	parser.RegisterParserFn(templateName, gen)
}
