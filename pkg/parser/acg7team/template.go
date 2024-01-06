package acg7team

import (
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/servicemgr"
)

const (
	templateName = "acg7team"
)

func gen(configPath string, namedServices *parser.CommonServices, services servicemgr.ServiceMap) (parser.Parser, error) {
	parser := &Acg7TeamParser{
		tmdbService:   namedServices.Tmdb,
		distOpService: namedServices.DiskOp,
	}
	return parser, nil
}

func init() {
	parser.RegisterParserFn(templateName, gen)
}
