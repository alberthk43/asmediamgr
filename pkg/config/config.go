package config

import (
	"time"

	"github.com/BurntSushi/toml"
)

type MontherDir struct {
	DirPath       string        `toml:"dir_path"`
	SleepInterval time.Duration `toml:"sleep_interval"`
}

type Configuration struct {
	ServiceConfDir    string       `toml:"service_conf_dir"`
	ParserConfDir     string       `toml:"parser_conf_dir"`
	TmdbSock5Proxy    string       `toml:"tmdb_sock5_proxy"`
	DestTvOnAirDir    string       `toml:"dest_tv_on_air_dir"`
	DestMovieOnAirDir string       `toml:"dest_movie_on_air_dir"`
	MotherDirs        []MontherDir `toml:"mother_dirs"`
}

func LoadConfigurationFromFile(file string) (*Configuration, error) {
	c := &Configuration{}
	_, err := toml.DecodeFile(file, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
