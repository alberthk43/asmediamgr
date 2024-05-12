package moviefile

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/common"
	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/disk"
	"asmediamgr/pkg/parser"
)

const (
	name = "moviefile"
)

func init() {
	parser.RegisterParser(name, &MovieFile{})
}

type PatternConfig struct {
	PatternStr string `toml:"pattern"`
	Pattern    *regexp.Regexp
}

type MovieFile struct {
	logger   log.Logger
	patterns []*PatternConfig
}

type Config struct {
	Patterns []*PatternConfig `toml:"patterns"`
}

func loadConfigFile(cfgPath string) (*Config, error) {
	cfg := &Config{}
	_, err := toml.DecodeFile(cfgPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return cfg, nil
}

func (p *MovieFile) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	cfg, _ := loadConfigFile(cfgPath)
	if cfg != nil {
		p.patterns = cfg.Patterns
	}
	p.logger = logger
	for _, pattern := range p.patterns {
		pattern.Pattern, err = regexp.Compile(pattern.PatternStr)
		if err != nil {
			return 0, fmt.Errorf("failed to compile pattern: %w", err)
		}
	}
	return 0, nil
}

func (p *MovieFile) IsDefaultEnable() bool {
	return true
}

func (p *MovieFile) Parse(entry *dirinfo.Entry, opts *parser.ParserMgrRunOpts) (ok bool, err error) {
	if entry.Type != dirinfo.FileEntry || len(entry.FileList) != 1 {
		return false, nil
	}
	file := entry.FileList[0]
	movieTargetDir, ok := opts.MediaTypeDirs[common.MediaTypeMovie]
	if !ok {
		return false, fmt.Errorf("movie target dir not found")
	}
	info, err := p.parse(entry)
	if err != nil {
		return false, err
	}
	if info == nil {
		return false, nil
	}
	level.Info(p.logger).Log("msg", "matched", "file", entry.Name(), "name", info.name, "originalName", info.originalName, "year", info.year, "tmdbid", info.tmdbid)
	diskService := parser.GetDefaultDiskService()
	err = diskService.RenameMovie(&disk.MovieRenameTask{
		OldPath:      filepath.Join(entry.MotherPath, file.RelPathToMother),
		NewMotherDir: movieTargetDir,
		OriginalName: info.originalName,
		Year:         info.year,
		Tmdbid:       info.tmdbid,
	})
	if err != nil {
		return false, fmt.Errorf("failed to rename movie: %w", err)
	}
	return true, nil
}

type movieInfo struct {
	name         string
	originalName string
	year         int
	tmdbid       int
}

func (p *MovieFile) parse(entry *dirinfo.Entry) (*movieInfo, error) {
	for _, pattern := range p.patterns {
		info, err := p.patternMatch(entry, pattern)
		if err != nil {
			return nil, err
		}
		if info != nil {
			return info, nil
		}
	}
	return nil, nil
}

var (
	defaultTmdbUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *MovieFile) patternMatch(entry *dirinfo.Entry, pattern *PatternConfig) (*movieInfo, error) {
	file := entry.FileList[0]
	entryNameWithoutExt, _ := strings.CutSuffix(file.Name, file.Ext)
	groups := pattern.Pattern.FindStringSubmatch(entryNameWithoutExt)
	if len(groups) == 0 {
		return nil, nil
	}
	info := &movieInfo{
		name:   "",
		tmdbid: -1,
	}
	for i, group := range pattern.Pattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch group {
		case "":
			continue
		case "name":
			info.name = groups[i]
		case "tmdbid":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, fmt.Errorf("failed to parse tmdbid: %w", err)
			}
			info.tmdbid = n
		case "year":
			n, err := strconv.Atoi(groups[i])
			if err != nil {
				return nil, fmt.Errorf("failed to parse year: %w", err)
			}
			info.year = n
		default:
			return nil, fmt.Errorf("unknown group: %s", group)
		}
	}
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid <= 0 {
		searchOpts := defaultTmdbUrlOptions
		if info.year > common.ValidStartYear {
			searchOpts["year"] = strconv.Itoa(info.year)
		}
		results, err := tmdbService.GetSearchMovies(info.name, searchOpts)
		if err != nil {
			return nil, err
		}
		if results.TotalResults <= 0 {
			return nil, fmt.Errorf("no movie found")
		}
		if results.TotalResults > 1 {
			var hits []string
			for i := 0; i < 3 && i < len(results.Results); i++ {
				hits = append(hits, fmt.Sprintf("%s-%d", results.Results[i].Title, results.Results[i].ID))
			}
			return nil, fmt.Errorf("multiple movies found, first 3 hits: %v", hits)
		}
		info.tmdbid = int(results.Results[0].ID)
	}
	detail, err := tmdbService.GetMovieDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, err
	}
	dt, err := common.ParseTmdbDateStr(detail.ReleaseDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse release date: %w", err)
	}
	info.originalName = detail.OriginalTitle
	info.year = dt.Year
	return info, nil
}
