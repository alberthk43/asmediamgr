package ddhdtv

import (
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher/basic"
)

const (
	dirNameRegexStr = `【高清剧集网发布 www\.DDHDTV\.com】(?P<name>.*)\[全\d+集\].* tv tmdbid-101306.*`
)

type DDHDTVMatcher struct {
	*basic.BasicSeasonDirMatcher
}

func NewDDHDTVMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*DDHDTVMatcher, error) {
	basic, err := basic.NewBasicSeasonDirMatcher(
		dirNameRegexStr,
		tmdbClient,
		renamer,
		targetPath,
	)
	if err != nil {
		return nil, err
	}
	matcher := &DDHDTVMatcher{
		BasicSeasonDirMatcher: basic,
	}
	return matcher, nil
}
