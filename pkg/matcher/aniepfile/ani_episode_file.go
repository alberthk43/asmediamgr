package aniepfile

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AniEpisodeFileMatcher struct {
	tmdbRegexp       *regexp.Regexp
	seasonNameRegexp *regexp.Regexp
	reginRegexp      *regexp.Regexp

	tmdbService matcher.TmdbService
	renamer     renamer.Renamer
	targetPath  string
}

const (
	tmdbRegexGroupNum  = 3
	nameGroupName      = "name"
	epGroupName        = "ep"
	tmdbRegexpStr      = `^\[ANi\] (?P<name>.*) - (?P<ep>\d+) \[.*`
	seasonNameRegexStr = `第(.*)季`
	reginRegexStr      = `（僅限港澳台地區）`
)

type Option func(*AniEpisodeFileMatcher)

func NewAnimeEpisodeFileMatcher(
	tmdbService matcher.TmdbService,
	renamer renamer.Renamer,
	targetPath string,
	opts ...Option,
) (*AniEpisodeFileMatcher, error) {
	aeMth := &AniEpisodeFileMatcher{
		tmdbRegexp:       regexp.MustCompile(tmdbRegexpStr),
		seasonNameRegexp: regexp.MustCompile(seasonNameRegexStr),
		reginRegexp:      regexp.MustCompile(reginRegexStr),
		tmdbService:      tmdbService,
		renamer:          renamer,
		targetPath:       targetPath,
	}
	for _, opt := range opts {
		opt(aeMth)
	}
	return aeMth, nil
}

func (m *AniEpisodeFileMatcher) Match(
	info *common.Info,
) (bool, error) {
	if info == nil {
		return false, fmt.Errorf("info nil")
	}
	if m.tmdbService == nil {
		return false, fmt.Errorf("tmdbClient nil")
	}
	if len(info.Subs) != 1 {
		return false, fmt.Errorf("info len != 1, no match")
	}
	fileInfo := info.Subs[0]
	if len(fileInfo.Paths) != 0 {
		return false, fmt.Errorf("not single file")
	}
	if fileInfo.IsDir {
		return false, fmt.Errorf("is dir")
	}
	if !common.IsMediaFile(fileInfo.Ext) {
		return false, fmt.Errorf("not media file")
	}
	groups := m.tmdbRegexp.FindStringSubmatch(fileInfo.Name)
	if len(groups) != tmdbRegexGroupNum {
		return false, fmt.Errorf("no match")
	}
	var err error
	season := 1
	var epNum int
	var searchName string
	for i, name := range m.tmdbRegexp.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case epGroupName:
			epNum, err = strconv.Atoi(groups[i])
			if err != nil {
				return false, fmt.Errorf("tmdbID convert err:%s", err)
			}
		case nameGroupName:
			searchName = groups[i]
		default:
			return false, fmt.Errorf("unknown group name:%s", name)
		}
	}
	if searchName == "" {
		return false, fmt.Errorf("searchName not found")
	}
	searchName = regulationRawName(searchName)
	if searchName == "" {
		return false, fmt.Errorf("regulationRawName err")
	}
	if matched := m.seasonNameRegexp.FindStringSubmatch(searchName); len(matched) == 2 {
		rawChineaseSeasonName := matched[1]
		newSeason := ChineseSeasonNameToNum(rawChineaseSeasonName)
		if newSeason != 0 {
			season = newSeason
			searchName = strings.ReplaceAll(searchName, matched[0], "")
			searchName = strings.TrimSpace(searchName)
		}
	}
	matched, err := m.tmdbService.SearchTvByName(searchName)
	if err != nil {
		return false, TMDBErr{err: err}
	}
	matched.EpNum = int32(epNum)
	matched.Season = int32(season)
	new, err := renamehelper.TargetTVEpFilePath(matched, m.targetPath, fileInfo.Ext)
	if err != nil {
		return false, RenameErr{err: err}
	}
	old := renamer.Path{
		info.DirPath,
		fmt.Sprintf("%s%s", fileInfo.Name, fileInfo.Ext),
	}
	err = m.renamer.Rename([]renamer.RenameRecord{{Old: old, New: new}})
	if err != nil {
		return false, err
	}
	return true, nil
}

type TMDBErr struct {
	err error
}

func (e TMDBErr) Error() string {
	return fmt.Sprintf("tmdb search err:%s", e.err)
}

func (e TMDBErr) UnWrap() error {
	return e.err
}

type RenameErr struct {
	err error
}

func (e RenameErr) Error() string {
	return fmt.Sprintf("rename err:%s", e.err)
}

func (e RenameErr) UnWrap() error {
	return e.err
}

func ChineseSeasonNameToNum(name string) int {
	if name == "" {
		return 0
	}
	if num, ok := ChineseSeasonNameToNumMapping[name]; ok {
		return num
	}
	return 0
}

var ChineseSeasonNameToNumMapping = map[string]int{
	"一": 1,
	"二": 2,
	"三": 3,
	"四": 4,
	"五": 5,
	"六": 6,
	"七": 7,
	"八": 8,
	"九": 9,
	"十": 10,
}

func regulationRawName(rawName string) string {
	rawName = strings.TrimSuffix(rawName, `（僅限港澳台地區）`)
	return rawName
}
