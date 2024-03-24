package tvepfile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"asmediamgr/pkg/common"
	"asmediamgr/pkg/dirinfo"
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
	PatternStr string `toml:"pattern"`
	Pattern    *regexp.Regexp
	Tmdbid     int      `toml:"tmdbid"`
	Season     int      `toml:"season"`
	OptNames   []string `toml:"opt_names"`
	Opts       []PatternOpt
}

type PatternOpt func(entry *dirinfo.Entry, info *tvEpInfo) error

func (p *TvEpFile) IsDefaultEnable() bool {
	return true
}

func (p *TvEpFile) Init(cfgPath string, logger log.Logger) (priority float32, err error) {
	p.logger = logger
	for _, pattern := range p.patterns {
		pattern.Pattern, err = regexp.Compile(pattern.PatternStr)
		if err != nil {
			return 0, fmt.Errorf("Compile() error = %v", err)
		}
		// TODO pattern opt
		// for _, optName := range pattern.OptNames {
		// 	opt, ok := patternOpts[optName]
		// 	if !ok {
		// 		return 0, fmt.Errorf("unknown optName = %s", optName)
		// 	}
		// 	pattern.Opts = append(pattern.Opts, opt)
		// }
	}
	return 0, nil
}

func (p *TvEpFile) Parse(entry *dirinfo.Entry) (ok bool, err error) {
	if entry.Type != dirinfo.FileEntry || len(entry.FileList) != 1 {
		return false, nil
	}
	info, err := p.parse(entry)
	if err != nil {
		return false, fmt.Errorf("parse() error = %v", err)
	}
	// do something with info
	level.Info(p.logger).Log("msg", "parse success", "info", info)
	return true, nil
}

func (p *TvEpFile) parse(entry *dirinfo.Entry) (info *tvEpInfo, err error) {
	for _, patter := range p.patterns {
		info, err = p.patternMatch(entry, patter)
		if err != nil {
			level.Warn(p.logger).Log("msg", "patternMatch() error", "err", err)
			continue
		}
		if info != nil {
			break
		}
	}
	if info == nil {
		return nil, fmt.Errorf("no pattern match")
	}
	return info, nil
}

const (
	groupName   = "name"
	groupSeason = "season"
	groupEpisod = "episode"
	groupTmdbid = "tmdbid"
	groupYear   = "year"
)

func (p *TvEpFile) patternMatch(entry *dirinfo.Entry, pattern *PatternConfig) (info *tvEpInfo, err error) {
	file := entry.FileList[0]
	entryNameWithoutExt, _ := strings.CutSuffix(file.Name, file.Ext)
	groups := pattern.Pattern.FindStringSubmatch(entryNameWithoutExt)
	if len(groups) == 0 {
		return nil, nil
	}
	info = &tvEpInfo{}
	for i, group := range pattern.Pattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch group {
		case "":
			continue
		case groupName:
			info.name = groups[i]
		case groupSeason:
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() season error = %v", err)
			}
			info.season = int(n)
		case groupEpisod:
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() episode error = %v", err)
			}
			info.episode = int(n)
		case groupTmdbid:
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() tmdbid error = %v", err)
			}
			info.tmdbid = int(n)
		case groupYear:
			n, err := strconv.ParseInt(groups[i], 10, 31)
			if err != nil {
				return nil, fmt.Errorf("ParseInt() year error = %v", err)
			}
			info.year = int(n)
		default:
			level.Warn(p.logger).Log("msg", "unknown pattern group", "group", group)
		}
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
	}
	return nil, nil
}

var (
	defaultTmdbUrlOptions = map[string]string{
		"include_adult": "true",
	}
)

func (p *TvEpFile) dealPreTmdbAndSeason(tmdbService parser.TmdbService, pattern *PatternConfig, info *tvEpInfo) (newInfo *tvEpInfo, err error) {
	tvDetail, err := tmdbService.GetTVDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, fmt.Errorf("GetTVDetails() error = %v", err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("invalid FirstAirDate = %s", tvDetail.FirstAirDate)
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
		return nil, fmt.Errorf("GetTVDetails() error = %v", err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("invalid FirstAirDate = %s", tvDetail.FirstAirDate)
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
		return nil, fmt.Errorf("GetSearchTVShow() error = %v", err)
	}
	if tvs.TotalResults <= 0 {
		return nil, fmt.Errorf("GetSearchTVShow() no result")
	}
	if tvs.TotalResults > 1 {
		return nil, fmt.Errorf("GetSearchTVShow() multiple result")
	}
	info.tmdbid = int(tvs.Results[0].ID)
	tvDetail, err := tmdbService.GetTVDetails(info.tmdbid, defaultTmdbUrlOptions)
	if err != nil {
		return nil, fmt.Errorf("GetTVDetails() error = %v", err)
	}
	dt, err := common.ParseTmdbDateStr(tvDetail.FirstAirDate)
	if err != nil {
		return nil, fmt.Errorf("invalid FirstAirDate = %s", tvDetail.FirstAirDate)
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
