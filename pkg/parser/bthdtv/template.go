package bthdtv

import (
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/servicemgr"
)

const (
	templateName = "bthdtv"
)

func gen(configPath string, namedServices *parser.CommonServices, services servicemgr.ServiceMap) (parser.Parser, error) {
	parser := &BtHdtvParser{
		tmdbService:   namedServices.Tmdb,
		distOpService: namedServices.DiskOp,
	}
	return parser, nil
}

func init() {
	parser.RegisterParserFn(templateName, gen)
}
