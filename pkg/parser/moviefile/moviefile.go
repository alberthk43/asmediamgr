package moviefile

import (
	"github.com/go-kit/log"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/parser"
)

const (
	name = "moviefile"
)

func init() {
	parser.RegisterParser(name, &MovieFile{})
}

type MovieFile struct{}

func (m *MovieFile) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	return 0, nil
}

func (m *MovieFile) IsDefaultEnable() bool {
	return true
}

func (p *MovieFile) Parse(entry *dirinfo.Entry, opts *parser.ParserMgrRunOpts) (ok bool, err error) {
	// Parse the movie file
	return false, nil
}
