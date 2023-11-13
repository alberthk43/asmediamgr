package short

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/component/fileinfocheck"
	"asmediamgr/pkg/component/regexparser"
	"asmediamgr/pkg/component/renamehelper"
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/matcher"
	"fmt"
	"regexp"
)

var _ matcher.Matcher = (*ShortMatcher)(nil)

const (
	movieShortRegStr = `^movieshort tmdbid-(?P<tmdbid>\d+)$`
	tvShortRegStr    = `^tvshort tmdbid-(?P<tmdbid>\d+)$`
)

type ShortMatcher struct {
	tmdbService   matcher.TmdbService
	renameService matcher.RenamerService
	pathService   matcher.PathService

	movieRegex *regexp.Regexp
	tvRegex    *regexp.Regexp
}

func NewShortMatcher(
	tmdbService matcher.TmdbService,
	renameService matcher.RenamerService,
	pathService matcher.PathService,
) (*ShortMatcher, error) {
	m := &ShortMatcher{
		tmdbService:   tmdbService,
		renameService: renameService,
		pathService:   pathService,
	}
	movieRegex, err := regexp.Compile(movieShortRegStr)
	if err != nil {
		return nil, err
	}
	m.movieRegex = movieRegex
	tvRegex, err := regexp.Compile(tvShortRegStr)
	if err != nil {
		return nil, err
	}
	m.tvRegex = tvRegex
	return m, nil
}

func (m *ShortMatcher) Match(info *common.Info) (ok bool, err error) {
	err = fileinfocheck.CheckFileInfo(info,
		fileinfocheck.CheckFileInfoHasOnlyOneMediaFile(),
		fileinfocheck.CheckFileInfoHasOnlyOneFileSizeGreaterThan(10*1024*1024),
	)
	if err != nil {
		return false, err
	}
	err = m.tryMovieShort(info)
	if err == nil {
		return true, nil
	}
	return false, err
}

func (m *ShortMatcher) tryMovieShort(info *common.Info) (err error) {
	fileInfo := info.Subs[0]
	groups := m.movieRegex.FindStringSubmatch(fileInfo.Name)
	subnames := m.movieRegex.SubexpNames()
	if len(groups) != len(subnames) {
		return fmt.Errorf("no movie match")
	}
	tmdbid, err := regexparser.ParseTmdbID(m.movieRegex, fileInfo.Name)
	if err != nil {
		return err
	}
	movieInfo, err := m.tmdbService.SearchMovieByTmdbID(tmdbid)
	if err != nil {
		return err
	}
	record := renamer.RenameRecord{
		Old: renamer.Path{
			m.pathService.MotherPath(),
			fileInfo.Name + fileInfo.Ext,
		},
		New: renamer.Path{
			m.pathService.TargetTvPath(),
			renamehelper.TargetMovieDir(movieInfo),
			renamehelper.TargetMovieShortDir(),
			fileInfo.Name + fileInfo.Ext,
		},
	}
	err = m.renameService.Rename([]renamer.RenameRecord{record})
	if err != nil {
		return err
	}
	return nil
}
