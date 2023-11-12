package singlemoviefile

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher"
	"fmt"
	"regexp"
	"strconv"
)

type SingleMovieFileMatcher struct {
	tmdbRegexp  *regexp.Regexp
	tmdbService matcher.TmdbService
	renamer     renamer.Renamer
	targetPath  string
}

var _ (matcher.Matcher) = (*SingleMovieFileMatcher)(nil)

const (
	tmdbRegexGroupNum = 2
	tmdbIDGroupName   = "tmdbid"
	tmdbRegexpStr     = `.* movie tmdbid-(?P<tmdbid>\d*)$`
)

type Option func(*SingleMovieFileMatcher)

func NewSingleMovieFileMatcher(
	tmdbService matcher.TmdbService,
	renamer renamer.Renamer,
	targetPath string,
	opts ...Option,
) (*SingleMovieFileMatcher, error) {
	if tmdbService == nil {
		return nil, fmt.Errorf("tmdbClient nil")
	}
	if renamer == nil {
		return nil, fmt.Errorf("renamer nil")
	}
	if targetPath == "" {
		return nil, fmt.Errorf("targetPath empty")
	}
	smlM := &SingleMovieFileMatcher{
		tmdbRegexp:  regexp.MustCompile(tmdbRegexpStr),
		tmdbService: tmdbService,
		renamer:     renamer,
		targetPath:  targetPath,
	}
	for _, opt := range opts {
		opt(smlM)
	}
	return smlM, nil
}

func (m *SingleMovieFileMatcher) Match(
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
		return false, fmt.Errorf("file name regex no match")
	}
	var tmdbID int64
	for i, name := range m.tmdbRegexp.SubexpNames() {
		if i == 0 {
			continue
		}
		if name == tmdbIDGroupName {
			var err error
			tmdbID, err = strconv.ParseInt(groups[i], 10, 63)
			if err != nil {
				return false, fmt.Errorf("tmdbID convert err:%s", err)
			}
		}
	}
	if tmdbID == 0 {
		return false, fmt.Errorf("tmdbID not found")
	}
	matched, err := m.tmdbService.SearchMovieByTmdbID(tmdbID)
	if err != nil {
		return false, err
	}
	new, err := renamehelper.TargetMovieFilePath(matched, m.targetPath, fileInfo.Ext)
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
	return true, nil
}
