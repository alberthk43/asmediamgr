package tvepfile

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/fileinfocheck"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pelletier/go-toml"
)

var _ matcher.Matcher = (*TvEpisodeFileMatcher)(nil)

const (
	configFileName = "tvepfile.toml"
)

type formatInfo struct {
	exp       *regexp.Regexp
	originStr string
}

type TvEpisodeFileMatcher struct {
	tmdbService   matcher.TmdbService
	renameService matcher.RenamerService
	targetService matcher.TargetService

	regexpList []*formatInfo
}

func NewTvEpisodeFileMatcher(
	configDir string,
	tmdbService matcher.TmdbService,
	renameService matcher.RenamerService,
	targetService matcher.TargetService,
) (*TvEpisodeFileMatcher, error) {
	if configDir == "" {
		return nil, fmt.Errorf("config dir nil")
	}
	if tmdbService == nil {
		return nil, fmt.Errorf("tmdbService nil")
	}
	if renameService == nil {
		return nil, fmt.Errorf("renameService nil")
	}
	m := &TvEpisodeFileMatcher{
		tmdbService:   tmdbService,
		renameService: renameService,
		targetService: targetService,
	}
	configFilePath := filepath.Join(configDir, configFileName)
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = toml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	for _, format := range config.Formats {
		regexpStr := format.Regexp
		aRegexp, err := regexp.Compile(regexpStr)
		if err != nil {
			return nil, err
		}
		m.regexpList = append(m.regexpList, &formatInfo{
			exp:       aRegexp,
			originStr: regexpStr,
		})
	}
	if len(config.Formats) <= 0 {
		return nil, fmt.Errorf("config is empty")
	}
	return m, nil
}

type TvEpFormat struct {
	Regexp string
}

type Config struct {
	Formats []TvEpFormat
}

func (m *TvEpisodeFileMatcher) Match(
	info *common.Info,
) (bool, error) {
	err := fileinfocheck.CheckFileInfo(info,
		fileinfocheck.CheckFileInfoHasOnlyOneMediaFile(),
		fileinfocheck.CheckFileInfoHasOnlyOneFileSizeGreaterThan(10*1024*1024),
	)
	if err != nil {
		return false, err
	}
	for _, aRegexp := range m.regexpList {
		err = m.matchWithRegexp(info, aRegexp)
		if err == nil {
			return true, nil
		}
	}
	return false, fmt.Errorf("no match in any one of regexpList")
}

func (m *TvEpisodeFileMatcher) matchWithRegexp(
	info *common.Info,
	formatInfo *formatInfo,
) error {
	fileInfo := &info.Subs[0]
	regexp := formatInfo.exp
	subexpNames := regexp.SubexpNames()
	groups := regexp.FindStringSubmatch(fileInfo.Name)
	if len(groups) != len(subexpNames) {
		return fmt.Errorf("no match")
	}
	var tvname string
	var seasonNum int32
	var epNum int32
	var tmdbid int64
	for i := 0; i < len(subexpNames); i++ {
		if i == 0 {
			continue
		}
		subexpName := subexpNames[i]
		group := groups[i]
		switch subexpName {
		case "tvname":
			tvname = group
		case "season":
			n, err := strconv.ParseInt(group, 10, 31)
			if err != nil {
				return fmt.Errorf("season not number")
			}
			seasonNum = int32(n)
		case "epnum":
			n, err := strconv.ParseInt(group, 10, 31)
			if err != nil {
				return fmt.Errorf("epnum not number")
			}
			epNum = int32(n)
		case "tmdbid":
			n, err := strconv.ParseInt(group, 10, 61)
			if err != nil {
				return fmt.Errorf("tmdbid not number")
			}
			tmdbid = n
		default:
			return fmt.Errorf("subexpName not defined, name:%s", subexpName)
		}
	}
	var tvInfo *common.MatchedTV
	if tmdbid != 0 {
		var err error
		tvInfo, err = m.tmdbService.SearchTvByTmdbID(tmdbid)
		if err != nil {
			return fmt.Errorf("tv search by tmdbid not found, id:%d", tmdbid)
		}
	} else {
		var err error
		tvInfo, err = m.tmdbService.SearchTvByName(tvname)
		if err != nil {
			return fmt.Errorf("tv search by name not found, name:%s", tvname)
		}
	}
	tvInfo.Season = seasonNum
	tvInfo.EpNum = epNum
	targetDir := m.targetService.TargetDir()
	aRecord, err := renamehelper.BuildRenameRecordFromSubInfo(
		info.DirPath,
		targetDir,
		fileInfo,
		tvInfo,
		seasonNum,
		epNum,
	)
	if err != nil {
		return err
	}
	err = m.renameService.Rename([]renamer.RenameRecord{*aRecord})
	if err != nil {
		return err
	}
	return nil
}
