package gmteam

import (
	"github.com/BurntSushi/toml"

	"asmediamgr/pkg/parser"
	"asmediamgr/pkg/service"
)

const (
	templateName = "gmteam"
)

type Predefined struct {
	Name      string `toml:"name"`
	TmdbId    int    `toml:"tmdbid"`
	SeasonNum int    `toml:"season_num"`
}

type Configuration struct {
	Predefined []Predefined `toml:"predefined_list"`
}

func loadConfiguration(configPath string) (*Configuration, error) {
	c := &Configuration{}
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func gen(configPath string, namedServices *parser.CommonServices, services service.ServiceMap) (parser.Parser, error) {
	c, err := loadConfiguration(configPath)
	if err != nil {
		return nil, err
	}
	parser := &GmTeamParser{
		c:             c,
		tmdbService:   namedServices.Tmdb,
		distOpService: namedServices.DiskOp,
	}
	return parser, nil
}

func init() {
	parser.RegisterParserFn(templateName, gen)
}
