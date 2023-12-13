package tvepfile

import (
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/service"
)

const (
	templateName = "tvepfile"
)

type TvEpParser struct {
	parser.DefaultPriority
}

func (p *TvEpParser) Parse(entry *dirinfo.Entry) error {
	return nil
}

func gen(configPath string, namedServices *parser.NamedServices, services service.ServiceMap) (parser.Parser, error) {
	parser := &TvEpParser{}
	return parser, nil
}

func init() {
	parser.RegisterParserFn(templateName, gen)
}
