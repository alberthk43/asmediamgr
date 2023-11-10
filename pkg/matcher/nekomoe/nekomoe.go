package nekomoe

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher"
	"fmt"
	"regexp"
	"strconv"
)

const (
	fileNameGroupNum  = 3
	fileNameGroupName = "name"
	fileNameGroupEp   = "ep"
	fileNameRegexStr  = `^\[Nekomoe kissaten\]\[(?P<name>.*)\]\[(?P<ep>\d+)\].*`
)

type NekomoeMather struct {
	fileNameRegexp *regexp.Regexp
	tmdbClient     tmdbhttp.TMDBClient
	renamer        renamer.Renamer
	targetPath     string
}

func NewNeomoeMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*NekomoeMather, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("tmdb client nil")
	}
	if renamer == nil {
		return nil, fmt.Errorf("renamer nil")
	}
	if targetPath == "" {
		return nil, fmt.Errorf("tv path empty")
	}
	matcher := &NekomoeMather{
		fileNameRegexp: regexp.MustCompile(fileNameRegexStr),
		tmdbClient:     tmdbClient,
		renamer:        renamer,
		targetPath:     targetPath,
	}
	return matcher, nil
}

var _ matcher.Matcher = (*NekomoeMather)(nil)

func (matcher *NekomoeMather) Match(info *common.Info) (bool, error) {
	return matcher.match(info)
}

func (matcher *NekomoeMather) match(info *common.Info) (bool, error) {
	// pre check is single media file, etc
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
	// regex match file name
	var searchName string
	var epNum int32
	season := int32(1)
	if groups := matcher.fileNameRegexp.FindStringSubmatch(fileInfo.Name); len(groups) == fileNameGroupNum {
		for i, name := range matcher.fileNameRegexp.SubexpNames() {
			if i == 0 {
				continue
			}
			switch name {
			case fileNameGroupName:
				searchName = groups[i]
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
	}
	if searchName == "" {
		return false, fmt.Errorf("search name empty")
	}
	// if match, search tmdb by name
	tmdbResult, err := tmdbhttp.SearchOnlyOneTVByName(matcher.tmdbClient, searchName)
	if err != nil {
		return false, err
	}
	matched, err := tmdbhttp.ConvTV(tmdbResult)
	if err != nil {
		return false, err
	}
	matched.EpNum = epNum
	matched.Season = season
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
	// rename file
	return false, nil
}
