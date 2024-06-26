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
	season       int
	episode      int
	tmdbid       int
	year         int
}

type PatternConfig struct {
	PatternStr    string   `toml:"pattern"`
	Tmdbid        int      `toml:"tmdbid"`
	Season        int      `toml:"season"`
	OptNames      []string `toml:"opt_names"`
	EpisodeOffset *int     `toml:"episode_offset"`
	Pattern       *regexp.Regexp
	Opts          []PatternOpt
}

type PatternOpt func(entry *dirinfo.Entry, info *tvEpInfo) error

const (
	ChineseSeasonOpt = "chinese_season_name"
)

var (
	patternOpts = map[string]PatternOpt{
		ChineseSeasonOpt: OptChineseSeasonName,
	}
)

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
		for _, optName := range pattern.OptNames {
			opt, ok := patternOpts[optName]
			if !ok {
				return 0, fmt.Errorf("unknown optName = %s", optName)
			}
			pattern.Opts = append(pattern.Opts, opt)
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
		Year:         info.year,
		Tmdbid:       info.tmdbid,
		Season:       info.season,
		Episode:      info.episode,
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
	entryNameWithoutExt, _ := strings.CutSuffix(file.Name, file.Ext)
	groups := pattern.Pattern.FindStringSubmatch(entryNameWithoutExt)
	if len(groups) == 0 {
		return nil, nil
	}
	info = &tvEpInfo{
		name:    "",
		season:  -1,
		episode: -1,
		tmdbid:  -1,
	}
	if pattern.Tmdbid > 0 {
		info.tmdbid = pattern.Tmdbid
	}
	if pattern.Season >= 0 {
		info.season = pattern.Season
	}
	for i, group := range pattern.Pattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch group {
		case "name":
			info.name = groups[i]
		case "season":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() season error = %v", err)
			}
			info.season = int(n)
		case "episode":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() episode error = %v", err)
			}
			info.episode = int(n)
		case "tmdbid":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() tmdbid error = %v", err)
			}
			info.tmdbid = int(n)
		case "year":
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() year error = %v", err)
			}
			info.year = int(n)
		default:
			level.Warn(p.logger).Log("msg", "unknown pattern group", "group", group)
		}
	}
	if pattern.EpisodeOffset != nil {
		info.episode += *pattern.EpisodeOffset
	}
	for _, opt := range pattern.Opts {
		err = opt(entry, info)
		if err != nil {
			return nil, fmt.Errorf("opt() error = %v", err)
		}
	}
	tmdbService := parser.GetDefaultTmdbService()
	if info.tmdbid > 0 && pattern.Season >= 0 && info.episode >= 0 {
		return p.dealPreTmdbAndSeason(tmdbService, pattern, info)
	} else if info.tmdbid > 0 && info.season >= 0 && info.episode >= 0 {
		return p.dealPreTmdbidAndScrapedSeason(tmdbService, info)
	} else if info.name != "" && info.season >= 0 && info.episode >= 0 {
		return p.dealSearchNameAndScrapedSeason(tmdbService, info)
	} else if info.name != "" && pattern.Season >= 0 && info.episode >= 0 {
		return p.dealSearchNameAndPreSeason(tmdbService, pattern, info)
	}
	return nil, fmt.Errorf("invalid info = %+v, pattern = %+v", info, pattern)
}

var (
	defaultTmdbUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *TvEpFile) dealPreTmdbAndSeason(tmdbService parser.TmdbService, pattern *PatternConfig, info *tvEpInfo) (newInfo *tvEpInfo, err error) {
	tvDetail, err := tmdbService.GetTVDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, fmt.Errorf("get pre tmdb and season, tmdbid = %d, err = %v", info.tmdbid, err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("get pre tmdb and season, tmdbid = %d, invalid FirstAirDate = %s", info.tmdbid, tvDetail.FirstAirDate)
	}
	newInfo = &tvEpInfo{
		originalName: tvDetail.OriginalName,
		season:       pattern.Season,
		episode:      info.episode,
		tmdbid:       info.tmdbid,
		year:         dt.Year,
	}
	return newInfo, nil
}

func (p *TvEpFile) dealPreTmdbidAndScrapedSeason(tmdbService parser.TmdbService, info *tvEpInfo) (newInfo *tvEpInfo, err error) {
	tvDetail, err := tmdbService.GetTVDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, fmt.Errorf("deal pre tmdb and scraped season, tmdbid = %d, error = %v", info.tmdbid, err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("deal pre tmdb and scraped season, tmdbid = %d, invalid FirstAirDate = %s", info.tmdbid, tvDetail.FirstAirDate)
	}
	newInfo = &tvEpInfo{
		originalName: tvDetail.OriginalName,
		season:       info.season,
		episode:      info.episode,
		tmdbid:       info.tmdbid,
		year:         dt.Year,
	}
	return newInfo, nil
}

func (p *TvEpFile) dealSearchNameAndScrapedSeason(tmdbService parser.TmdbService, info *tvEpInfo) (newInfo *tvEpInfo, err error) {
	urlOptions := defaultTmdbUrlOptions
	if info.year > common.ValidStartYear {
		urlOptions["year"] = strconv.Itoa(info.year)
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
	info.tmdbid = int(tvs.Results[0].ID)
	tvDetail, err := tmdbService.GetTVDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, fmt.Errorf("deal search name and scraped season, get detail of tmdbid = %d, error = %v", info.tmdbid, err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("deal search name and scraped season, invalid FirstAirDate = %s", tvDetail.FirstAirDate)
	}
	newInfo = &tvEpInfo{
		originalName: tvDetail.OriginalName,
		season:       info.season,
		episode:      info.episode,
		tmdbid:       info.tmdbid,
		year:         dt.Year,
	}
	return newInfo, nil
}

func (p *TvEpFile) dealSearchNameAndPreSeason(tmdbService parser.TmdbService, pattern *PatternConfig, info *tvEpInfo) (newInfo *tvEpInfo, err error) {
	urlOptions := defaultTmdbUrlOptions
	if info.year > common.ValidStartYear {
		urlOptions["year"] = strconv.Itoa(info.year)
	}
	tvs, err := tmdbService.GetSearchTVShow(info.name, urlOptions)
	if err != nil {
		return nil, fmt.Errorf("deal search name and pre seaon, name = %s, year = %d, error = %v", info.name, info.year, err)
	}
	if tvs.TotalResults <= 0 {
		return nil, fmt.Errorf("deal search name and pre seaon, name = %s, year = %d, no result", info.name, info.year)
	}
	if tvs.TotalResults > 1 {
		var results []string
		for i := 0; i < 3 && i < len(tvs.Results); i++ {
			results = append(results, fmt.Sprintf("%s-%d", tvs.Results[i].Name, tvs.Results[i].ID))
		}
		return nil, fmt.Errorf("deal search name and pre seaon, name = %s, year = %d, multiple results, first 3 results = %v", info.name, info.year, results)
	}
	info.tmdbid = int(tvs.Results[0].ID)
	tvDetail, err := tmdbService.GetTVDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, fmt.Errorf("deal search name and pre seaon, tmdbid = %d, error = %v", info.tmdbid, err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("deal search name and pre seaon, invalid FirstAirDate = %s", tvDetail.FirstAirDate)
	}
	newInfo = &tvEpInfo{
		originalName: tvDetail.OriginalName,
		season:       pattern.Season,
		episode:      info.episode,
		tmdbid:       info.tmdbid,
		year:         dt.Year,
	}
	return newInfo, nil
}

var (
	chineseSeasonNamePattern = regexp.MustCompile(`(?P<name>.*)第(?P<seasonch>.*)季.*`)
)

func OptChineseSeasonName(entry *dirinfo.Entry, info *tvEpInfo) error {
	if info.season >= 0 {
		return nil
	}
	groups := chineseSeasonNamePattern.FindStringSubmatch(info.name)
	if len(groups) == 0 {
		return nil
	}
	numStr := groups[2]
	var err error
	var n int
	n, err = strconv.Atoi(numStr)
	if err != nil {
		var ok bool
		n, ok = common.ChineseToNum(numStr)
		if !ok {
			return fmt.Errorf("ChineseToNum() not chinese number = %s", numStr)
		}
	}
	info.season = int(n)
	info.name = groups[1]
	return nil
}
