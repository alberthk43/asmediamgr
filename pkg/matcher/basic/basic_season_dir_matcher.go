package basic

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"fmt"
	"regexp"
	"strconv"
)

const (
	dirNameGroupName   = "name"
	dirNameGroupTMDBID = "tmdbid"

	dirTMDBIDRegexpStr = `.* tv tmdbid-(?P<tmdbid>\d+)$`

	subNameRegexStr1 = `.*S(?P<season>\d+)(?:EP|E)(?P<ep>\d+).*`
	subNameRegexStr2 = `.*(?:ep|Ep|EP|E)(?P<ep>\d+).*`

	subNameSeason = "season"
	subNameEp     = "ep"
)

type BasicSeasonDirMatcher struct {
	dirNameRegexp *regexp.Regexp

	dirTMDBIDRegexp *regexp.Regexp
	subNameRegexps  []*regexp.Regexp

	tmdbClient tmdbhttp.TMDBClient
	renamer    renamer.Renamer
	targetPath string
}

func NewBasicSeasonDirMatcher(
	dirNameRegexpStr string,
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*BasicSeasonDirMatcher, error) {
	matcher := &BasicSeasonDirMatcher{
		dirNameRegexp: regexp.MustCompile(dirNameRegexpStr),
		subNameRegexps: []*regexp.Regexp{
			regexp.MustCompile(subNameRegexStr1),
			regexp.MustCompile(subNameRegexStr2),
		},
		dirTMDBIDRegexp: regexp.MustCompile(dirTMDBIDRegexpStr),
		tmdbClient:      tmdbClient,
		renamer:         renamer,
		targetPath:      targetPath,
	}
	return matcher, nil
}

func (matcher *BasicSeasonDirMatcher) Match(
	info *common.Info,
) (bool, error) {
	return matcher.match(info)
}

func (matcher *BasicSeasonDirMatcher) match(
	info *common.Info,
) (bool, error) {
	// precheck info, dir
	if info == nil {
		return false, fmt.Errorf("info nil")
	}
	dirInfo := info.Subs[0]
	if !dirInfo.IsDir {
		return false, fmt.Errorf("not dir")
	}
	// try regex with dir, by name or by tmdbid
	var tmdbID int64
	var searchName string
	var dirSeason int32
	if groups := matcher.dirNameRegexp.FindStringSubmatch(dirInfo.Name); len(groups) > 1 {
		for i, name := range matcher.dirNameRegexp.SubexpNames() {
			if i == 0 {
				continue
			}
			switch name {
			case dirNameGroupName:
				searchName = groups[i]
			case subNameSeason:
				n, err := strconv.ParseInt(groups[i], 10, 31)
				if err != nil {
					return false, fmt.Errorf("season not int")
				}
				dirSeason = int32(n)
			default:
				return false, fmt.Errorf("unknown group name")
			}
		}
	}
	// optional search by tmdbid
	if groups := matcher.dirTMDBIDRegexp.FindStringSubmatch(dirInfo.Name); len(groups) > 1 {
		for i, name := range matcher.dirTMDBIDRegexp.SubexpNames() {
			if i == 0 {
				continue
			}
			switch name {
			case dirNameGroupTMDBID:
				tmdbIDStr := groups[i]
				n, err := strconv.ParseInt(tmdbIDStr, 10, 64)
				if err != nil {
					return false, fmt.Errorf("tmdbid not int")
				}
				tmdbID = int64(n)
			default:
				return false, fmt.Errorf("unknown group name")
			}
		}
	}
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
	// try loop all sub files, match by sub regex, >0 media file matched, no sub
	var renamers []renamer.RenameRecord
	for _, subInfo := range info.Subs[1:] {
		if subInfo.IsDir {
			continue
		}
		if common.IsSubtitleFile(subInfo.Ext) {
			return false, fmt.Errorf("subtitle file not supported now")
		}
		if !common.IsMediaFile(subInfo.Ext) {
			continue
		}
		var season int32
		if dirSeason > 0 {
			season = dirSeason
		} else {
			season = 1
		}
		var epNum int32
		for _, subNameRegex := range matcher.subNameRegexps {
			if groups := subNameRegex.FindStringSubmatch(subInfo.Name); len(groups) > 1 {
				for i, name := range subNameRegex.SubexpNames() {
					if i == 0 {
						continue
					}
					switch name {
					case subNameSeason:
						seasonStr := groups[i]
						n, err := strconv.ParseInt(seasonStr, 10, 31)
						if err != nil {
							return false, fmt.Errorf("season not int")
						}
						season = int32(n)
					case subNameEp:
						epNumStr := groups[i]
						n, err := strconv.ParseInt(epNumStr, 10, 31)
						if err != nil {
							return false, fmt.Errorf("epnum not int")
						}
						epNum = int32(n)
					default:
						return false, fmt.Errorf("unknown sub group name")
					}
				}
			}
		}
		if season < 0 || epNum <= 0 {
			continue
		}
		matched.Season = season
		matched.EpNum = epNum
		new, err := renamehelper.TargetTVEpFilePath(matched, matcher.targetPath, subInfo.Ext)
		if err != nil {
			return false, err
		}
		old := renamer.Path{info.DirPath}
		old = append(old, subInfo.Paths...)
		old = append(old, fmt.Sprintf("%s%s", subInfo.Name, subInfo.Ext))
		renamers = append(renamers, renamer.RenameRecord{Old: old, New: new})
	}
	// renamer
	err := matcher.renamer.Rename(renamers)
	if err != nil {
		return false, err
	}
	return true, nil
}
