package tvepfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/yaml.v2"

	"github.com/albert43/asmediamgr/pkg/common"
	"github.com/albert43/asmediamgr/pkg/dirinfo"
	"github.com/albert43/asmediamgr/pkg/disk"
	"github.com/albert43/asmediamgr/pkg/parser"
)

const (
	name = "tvepfile"
)

func init() {
	parser.RegisterParser(name, &TvEpFile{})
}

type Config struct {
	Patterns []*Pattern `yaml:"patterns"`
}

type Pattern struct {
	Name          string `yaml:"name"`
	Pattern       string `yaml:"pattern"`
	Tmdbid        *int   `yaml:"tmdbid"`
	Season        *int   `yaml:"season"`
	EpisodeOffset *int   `yaml:"episode_offset"`
	PatternRegexp *regexp.Regexp
}

type TvEpFile struct {
	logger   log.Logger
	patterns []*Pattern
}

func (p *TvEpFile) IsDefaultEnable() bool {
	return true
}

func (p *TvEpFile) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	cfg := &Config{}
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	err = yaml.Unmarshal(file, cfg)
	if err != nil {
		return 0, fmt.Errorf("unmarshal yaml file error %v", err)
	}
	if len(cfg.Patterns) == 0 {
		return 0, fmt.Errorf("no patterns in config")
	}
	p.patterns = cfg.Patterns
	p.logger = logger
	for _, pattern := range p.patterns {
		pattern.PatternRegexp, err = regexp.Compile(pattern.Pattern)
		if err != nil {
			return 0, fmt.Errorf("Compile() error = %v", err)
		}
	}
	return 0, nil
}

type tvEpInfo struct {
	name         string
	originalName string
	season       *int
	episode      *int
	tmdbid       *int
	year         *int
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

func (p *TvEpFile) patternMatch(entry *dirinfo.Entry, pattern *Pattern) (info *tvEpInfo, err error) {
	file := entry.FileList[0]
	groups := pattern.PatternRegexp.FindStringSubmatch(file.PureName)
	if len(groups) == 0 {
		return nil, nil
	}
	info = &tvEpInfo{
		tmdbid: pattern.Tmdbid,
		season: pattern.Season,
	}
	for i, group := range pattern.PatternRegexp.SubexpNames() {
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
	if info.season == nil {
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
	info.originalName = tvDetail.OriginalName
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
