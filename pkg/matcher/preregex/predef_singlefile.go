package preregex

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/fileinfo"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pelletier/go-toml"
)

const (
	fileNameGroupNum = 2
	fileNameGroupEp  = "ep"
)

type PreRegexMatcher struct {
	preDefRegexp []subRegexMatcher
	tmdbService  matcher.TmdbService
	renamer      renamer.Renamer
	targetPath   string
}

type subRegexMatcher struct {
	regexp *regexp.Regexp
	tmdbid int64
	season int32
}

func NewPreRegexMatcher(
	confPath string,
	tmdbService matcher.TmdbService,
	renamer renamer.Renamer, // TODO to diskop service
	targetPath string, // TODO to target path service
) (*PreRegexMatcher, error) {
	if tmdbService == nil {
		return nil, fmt.Errorf("tmdb client nil")
	}
	if renamer == nil {
		return nil, fmt.Errorf("renamer nil")
	}
	if targetPath == "" {
		return nil, fmt.Errorf("tv path empty")
	}
	matcher := &PreRegexMatcher{
		tmdbService: tmdbService,
		renamer:     renamer,
		targetPath:  targetPath,
	}
	configFilePath := filepath.Join(confPath, "regexp.toml")
	atLeastOne := false
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var conf predefinedConfig
	err = toml.Unmarshal(bytes, &conf)
	if err != nil {
		return nil, err
	}
	for _, regexStr := range conf.Regexps {
		aRegex, err := regexp.Compile(regexStr.Regexp)
		if err != nil {
			return nil, err
		}
		if regexStr.Tmdbid <= 0 {
			return nil, fmt.Errorf("tmdbid <= 0")
		}
		if regexStr.Season < 0 {
			return nil, fmt.Errorf("season < 0")
		}
		matcher.preDefRegexp = append(matcher.preDefRegexp, subRegexMatcher{
			regexp: aRegex,
			tmdbid: regexStr.Tmdbid,
			season: regexStr.Season,
		})
		atLeastOne = true
	}
	if !atLeastOne {
		return nil, fmt.Errorf("at least one regex")
	}
	return matcher, nil
}

var _ matcher.Matcher = (*PreRegexMatcher)(nil)

type predefinedConfig struct {
	Regexps []RegexRecord
}

type RegexRecord struct {
	Regexp string
	Tmdbid int64
	Season int32
}

func (m *PreRegexMatcher) Match(info *common.Info) (bool, error) {
	// pre check is single media file, etc
	if err := fileinfo.CheckIsSingleMediaFile(info); err != nil {
		return false, err
	}
	fileInfo := &info.Subs[0]
	// for loop all sub pre defined regex match file name
	var tmdbID int64
	var season int32
	var epNum int32
	for _, subRegex := range m.preDefRegexp {
		if groups := subRegex.regexp.FindStringSubmatch(fileInfo.Name); len(groups) == fileNameGroupNum {
			for i, name := range subRegex.regexp.SubexpNames() {
				if i == 0 {
					continue
				}
				switch name {
				case fileNameGroupEp:
					epNumStr := groups[i]
					n, err := strconv.Atoi(epNumStr)
					if err != nil {
						return false, fmt.Errorf("ep num not int")
					}
					epNum = int32(n)
				default:
					return false, fmt.Errorf("unknown group name")
				}
			}
			tmdbID = subRegex.tmdbid
			season = subRegex.season
			break
		}
	}
	// if match, search tmdb by name
	if tmdbID <= 0 {
		return false, fmt.Errorf("tmdbid <= 0, not match")
	}
	if season < 0 {
		return false, fmt.Errorf("invalid season:%d", season)
	}
	if epNum < 0 {
		return false, fmt.Errorf("invalid ep num:%d", epNum)
	}
	tmdbTvInfo, err := m.tmdbService.SearchTvByTmdbID(tmdbID)
	if err != nil {
		return false, err
	}
	tmdbTvInfo.EpNum = epNum
	tmdbTvInfo.Season = season
	new, err := renamer.TargetTVEpFilePath(tmdbTvInfo, m.targetPath, fileInfo.Ext)
	if err != nil {
		return false, err
	}
	old := renamer.Path{
		info.DirPath,
		fmt.Sprintf("%s%s", fileInfo.Name, fileInfo.Ext),
	}
	err = m.renamer.Rename([]renamer.RenameRecord{{Old: old, New: new}})
	if err != nil {
		return false, err
	}
	// rename file
	return true, nil
}
