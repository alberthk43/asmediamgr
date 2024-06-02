package tvepfile

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
	name = "tvepfile"
)

func init() {
	parser.RegisterParser(name, &TvEpFile{})
}

type TvEpFile struct {
	logger   log.Logger
	patterns []*PatternConfig
}

type tvEpInfo struct {
	name         string
	originalName string
	season       *int
	episode      *int
	tmdbid       *int
	year         *int
}

type PatternConfig struct {
	PatternStr    string   `toml:"pattern"`
	Tmdbid        *int     `toml:"tmdbid"`
	Season        *int     `toml:"season"`
	OptNames      []string `toml:"opt_names"`
	EpisodeOffset *int     `toml:"episode_offset"`
	Pattern       *regexp.Regexp
}

func (p *TvEpFile) IsDefaultEnable() bool {
	return true
}

type Config struct {
	Patterns []*PatternConfig `toml:"patterns"`
}

func loadConfigFile(cfgPath string) (*Config, error) {
	cfg := &Config{}
	_, err := toml.DecodeFile(cfgPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("DecodeFile() error = %v", err)
	}
	return cfg, nil
}

func (p *TvEpFile) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	cfg, _ := loadConfigFile(cfgPath) // allow no-config
	if cfg != nil {
		p.patterns = cfg.Patterns // TODO make 0 as invalid
	}
	p.logger = logger
	for _, pattern := range p.patterns {
		pattern.Pattern, err = regexp.Compile(pattern.PatternStr)
		if err != nil {
			return 0, fmt.Errorf("Compile() error = %v", err)
		}
	}
	return 0, nil
}

func (p *TvEpFile) Parse(entry *dirinfo.Entry, opts *parser.ParserMgrRunOpts) (ok bool, err error) {
	if entry.Type != dirinfo.FileEntry || len(entry.FileList) != 1 {
		return false, nil
	}
	file := entry.FileList[0]
	tvMediaTargetDir, ok := opts.MediaTypeDirs[common.MediaTypeTv]
	if !ok {
		return false, fmt.Errorf("no tv media target dir")
	}
	info, err := p.parse(entry)
	if err != nil {
		return false, fmt.Errorf("parse() error = %v", err)
	}
	if info == nil {
		return false, nil // no match and no error
	}
	level.Info(p.logger).Log("msg", "matched", "file", entry.Name(), "name", info.name, "originalName", info.originalName,
		"season", info.season, "episode", info.episode, "tmdbid", info.tmdbid, "year", info.year)
	diskService := parser.GetDefaultDiskService()
	err = diskService.RenameTvEpisode(&disk.TvEpisodeRenameTask{
		OldPath:      filepath.Join(entry.MotherPath, file.RelPathToMother),
		NewMotherDir: tvMediaTargetDir,
		OriginalName: info.originalName,
		Year:         *info.year,
		Tmdbid:       *info.tmdbid,
		Season:       *info.season,
		Episode:      *info.episode,
	})
	if err != nil {
		return false, fmt.Errorf("diskService.RenameTvEpisode() error = %v", err)
	}
	return true, nil
}

func (p *TvEpFile) parse(entry *dirinfo.Entry) (info *tvEpInfo, err error) {
	for _, pattern := range p.patterns {
		info, err = p.patternMatch(entry, pattern)
		if err != nil {
			return nil, err // error, stop all parsers
		}
		if info != nil {
			return info, nil // matched, return info
		}
	}
	return nil, nil // no match and no error
}

func (p *TvEpFile) patternMatch(entry *dirinfo.Entry, pattern *PatternConfig) (info *tvEpInfo, err error) {
	file := entry.FileList[0]
	groups := pattern.Pattern.FindStringSubmatch(file.PureName)
	if len(groups) == 0 {
		return nil, nil
	}
	info = &tvEpInfo{
		tmdbid: pattern.Tmdbid,
		season: pattern.Season,
	}
	for i, group := range pattern.Pattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch group {
		case "name":
			info.name = strings.Trim(groups[i], " ")
		case "chineaseseasonname":
			season, name, err := parseChineseSeasonName(groups[i])
			if err != nil {
				return nil, nil
			}
			info.name = strings.Trim(name, " ")
			if info.season == nil {
				info.season = &season
			}
		case "numberseasonname":
			season, name, err := parseNumberSeasonName(groups[i])
			if err != nil {
				return nil, nil
			}
			info.name = strings.Trim(name, " ")
			if info.season == nil {
				info.season = &season
			}
		case "season":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() season error = %v", err)
			}
			if info.season == nil {
				season := int(n)
				info.season = &season
			}
		case "episode":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() episode error = %v", err)
			}
			episode := int(n)
			info.episode = &episode
		case "tmdbid":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() tmdbid error = %v", err)
			}
			tmdbid := int(n)
			info.tmdbid = &tmdbid
		case "year":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() year error = %v", err)
			}
			year := int(n)
			info.year = &year
		default:
			level.Warn(p.logger).Log("msg", "unknown pattern group", "group", group)
		}
	}
	if info.episode == nil {
		return nil, nil
	}
	if pattern.EpisodeOffset != nil {
		*info.episode += *pattern.EpisodeOffset
	}
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid == nil {
		urlOptions := common.DefaultTmdbSearchOpts
		if info.year != nil {
			urlOptions["year"] = strconv.Itoa(*info.year)
		}
		tvs, err := tmdbService.GetSearchTVShow(info.name, urlOptions)
		if err != nil {
			return nil, fmt.Errorf("deal search name and scraped season, name = %s, year = %d, error = %v", info.name, info.year, err)
		}
		if tvs.TotalResults <= 0 {
			return nil, fmt.Errorf("deal search name and scraped season, name = %s, year = %d, no result", info.name, info.year)
		}
		if tvs.TotalResults > 1 {
			var results []string
			for i := 0; i < 3 && i < len(tvs.Results); i++ {
				results = append(results, fmt.Sprintf("%s-%d", tvs.Results[i].Name, tvs.Results[i].ID))
			}
			return nil, fmt.Errorf("deal search name and scraped season, multiple result, search name = %s, year = %d ,first 3 results = %v", info.name, info.year, results)
		}
		tmdbid := int(tvs.Results[0].ID)
		info.tmdbid = &tmdbid
	}
	tvDetail, err := tmdbService.GetTVDetails(*info.tmdbid, common.DefaultTmdbSearchOpts)
	if err != nil {
		return nil, fmt.Errorf("get pre tmdb and season, tmdbid = %d, err = %v", info.tmdbid, err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("get pre tmdb and season, tmdbid = %d, invalid FirstAirDate = %s", info.tmdbid, tvDetail.FirstAirDate)
	}
	info.year = &dt.Year
	return info, nil
}

var (
	numberSeasonNamePattern  = regexp.MustCompile(`(?P<name>.*)第(?P<season>\d+)季`)
	chineseSeasonNamePattern = regexp.MustCompile(`(?P<name>.*)第(?P<seasonch>.*)季`)
)

func parseChineseSeasonName(seasonName string) (season int, name string, err error) {
	groups := chineseSeasonNamePattern.FindStringSubmatch(seasonName)
	if len(groups) == 0 {
		return -1, "", fmt.Errorf("no match")
	}
	numStr := groups[2]
	var n int
	n, err = strconv.Atoi(numStr)
	if err != nil {
		var ok bool
		n, ok = common.ChineseToNum(numStr)
		if !ok {
			return -1, "", fmt.Errorf("ChineseToNum() not chinese number = %s", numStr)
		}
	}
	return n, groups[1], nil
}

func parseNumberSeasonName(seasonName string) (season int, name string, err error) {
	groups := numberSeasonNamePattern.FindStringSubmatch(seasonName)
	if len(groups) == 0 {
		return -1, "", fmt.Errorf("no match")
	}
	n, err := strconv.Atoi(groups[2])
	if err != nil {
		return -1, "", fmt.Errorf("Atoi() error = %v", err)
	}
	return n, groups[1], nil
}
