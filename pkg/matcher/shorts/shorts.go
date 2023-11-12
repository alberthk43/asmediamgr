package shorts

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
	fileNameRegexpStr = `(?P<publisher>.*) movie short tmdbid-(?P<tmdbid>\d*)$`
)

type ShortMatcher struct {
	fileNameRegex *regexp.Regexp
	tmdbClient    tmdbhttp.TMDBClient
	renamer       renamer.Renamer
	targetPath    string
}

func NewShortMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*ShortMatcher, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("tmdb client nil")
	}
	if renamer == nil {
		return nil, fmt.Errorf("renamer nil")
	}
	if targetPath == "" {
		return nil, fmt.Errorf("tv path empty")
	}
	matcher := &ShortMatcher{
		fileNameRegex: regexp.MustCompile(fileNameRegexpStr),
		tmdbClient:    tmdbClient,
		renamer:       renamer,
		targetPath:    targetPath,
	}
	return matcher, nil
}

func (matcher *ShortMatcher) Match(
	info *common.Info,
) (bool, error) {
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
	var err error
	var tmdbID int64
	var publisher string
	if groups := matcher.fileNameRegex.FindStringSubmatch(fileInfo.Name); len(groups) > 1 {
		for i, name := range matcher.fileNameRegex.SubexpNames() {
			if i == 0 {
				continue
			}
			switch name {
			case "tmdbid":
				tmdbID, err = strconv.ParseInt(groups[i], 10, 63)
				if err != nil {
					return false, fmt.Errorf("tmdbID convert err:%s", err)
				}
			case "publisher":
				publisher = groups[i]
			default:
				return false, fmt.Errorf("unknown group name")
			}
		}
	}
	if tmdbID == 0 {
		return false, fmt.Errorf("tmdbID not found")
	}
	data, err := tmdbhttp.SearchMovieByTmdbID(matcher.tmdbClient, tmdbID)
	if err != nil {
		return false, fmt.Errorf("search tmdb err:%s", err)
	}
	matched, err := tmdbhttp.ConvMovie(data)
	if err != nil {
		return false, fmt.Errorf("conv tmdb err:%s", err)
	}
	new, err := renamehelper.TargetMovieShortFilePath(matched, matcher.targetPath, publisher, fileInfo.Ext)
	if err != nil {
		return false, fmt.Errorf("target path err:%s", err)
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
