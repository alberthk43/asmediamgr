package basic

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	fileNameGroupName   = "name"
	fileNameGroupSeason = "season"
	fileNameGroupEp     = "ep"
	fileNameGroupTMDBID = "tmdbid"
)

type BasicSingleFileMatcher struct {
	fileNameRegexp *regexp.Regexp

	tmdbClient tmdbhttp.TMDBClient
	renamer    renamer.Renamer
	targetPath string

	optionalSearchNameMapping map[string]string
}

type Optional func(*BasicSingleFileMatcher)

func OptionalReplaceSearchName(mapping map[string]string) Optional {
	return func(matcher *BasicSingleFileMatcher) {
		matcher.optionalSearchNameMapping = mapping
	}
}

func NewBasicSingleFileMatcher(
	fileNameRegexpStr string,
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
	opts ...Optional,
) (*BasicSingleFileMatcher, error) {
	matcher := &BasicSingleFileMatcher{
		fileNameRegexp: regexp.MustCompile(fileNameRegexpStr),
		tmdbClient:     tmdbClient,
		renamer:        renamer,
		targetPath:     targetPath,
	}
	for _, opt := range opts {
		opt(matcher)
	}
	return matcher, nil
}

func (matcher *BasicSingleFileMatcher) Match(
	info *common.Info,
) (bool, error) {
	return matcher.match(info)
}

func (matcher *BasicSingleFileMatcher) match(
	info *common.Info,
) (bool, error) {
	// precheck info, single media file
	if info == nil {
		return false, fmt.Errorf("info nil")
	}
	if len(info.Subs) != 1 {
		return false, fmt.Errorf("not single file")
	}
	fileInfo := info.Subs[0]
	if fileInfo.IsDir {
		return false, fmt.Errorf("not single file")
	}
	if !common.IsMediaFile(fileInfo.Ext) {
		return false, fmt.Errorf("not media file")
	}
	// try regexp match, search for group
	var tmdbID int64
	var searchName string
	var season, epNum int32
	season = 1
	if groups := matcher.fileNameRegexp.FindStringSubmatch(fileInfo.Name); len(groups) > 1 {
		for i, name := range matcher.fileNameRegexp.SubexpNames() {
			if i == 0 {
				continue
			}
			switch name {
			case fileNameGroupName:
				searchName = groups[i]
			case fileNameGroupSeason:
				seasonStr := groups[i]
				n, err := strconv.Atoi(seasonStr)
				if err != nil {
					return false, fmt.Errorf("season num not int")
				}
				season = int32(n)
			case fileNameGroupEp:
				epNumStr := groups[i]
				n, err := strconv.Atoi(epNumStr)
				if err != nil {
					return false, fmt.Errorf("ep num not int")
				}
				epNum = int32(n)
			case fileNameGroupTMDBID:
				tmdbIDStr := groups[i]
				n, err := strconv.ParseInt(tmdbIDStr, 10, 64)
				if err != nil {
					return false, fmt.Errorf("tmdbID num not int")
				}
				tmdbID = int64(n)
			default:
				return false, fmt.Errorf("unknown group name")
			}
		}
	}
	if season < 0 {
		return false, fmt.Errorf("season < 0")
	}
	if epNum < 0 {
		return false, fmt.Errorf("epNum < 0")
	}
	// check if regexp matched, name(optional), season, ep, tmdbid(optional)
	var matched *common.MatchedTV
	if tmdbID > 0 {
		data, err := tmdbhttp.SearchTVByTmdbID(matcher.tmdbClient, tmdbID)
		if err != nil {
			return false, err
		}
		matched, err = tmdbhttp.ConvTV(data)
		if err != nil {
			return false, err
		}
	} else {
		if matcher.optionalSearchNameMapping != nil {
			for from, to := range matcher.optionalSearchNameMapping {
				searchName = strings.ReplaceAll(searchName, from, to)
			}
		}
		if searchName == "" {
			return false, fmt.Errorf("searchName not found")
		}
		data, err := tmdbhttp.SearchOnlyOneTVByName(matcher.tmdbClient, searchName)
		if err != nil {
			return false, err
		}
		matched, err = tmdbhttp.ConvTV(data)
		if err != nil {
			return false, err
		}
	}
	if matched == nil {
		return false, fmt.Errorf("no match")
	}
	matched.Season = season
	matched.EpNum = epNum
	// rename
	new, err := renamehelper.TargetTVEpFilePath(matched, matcher.targetPath, fileInfo.Ext)
	if err != nil {
		return false, err
	}
	old := renamer.Path{
		info.DirPath,
		fmt.Sprintf("%s%s", fileInfo.Name, fileInfo.Ext),
	}
	err = matcher.renamer.Rename([]renamer.RenameRecord{{Old: old, New: new}})
	if err != nil {
		return false, err
	}
	return true, nil
}
