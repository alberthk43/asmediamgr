package sevenacg

import (
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher/basic"
)

const (
	dirNameRegexStr = `\[BDrip\] (?P<name>.*) S(?P<season>\d+) \[7³ACG\]`
)

type SevenAcgMatcher struct {
	*basic.BasicSeasonDirMatcher
}

func NewSevenAcgMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*SevenAcgMatcher, error) {
	basic, err := basic.NewBasicSeasonDirMatcher(
		dirNameRegexStr,
		tmdbClient,
		renamer,
		targetPath,
	)
	if err != nil {
		return nil, err
	}
	matcher := &SevenAcgMatcher{
		BasicSeasonDirMatcher: basic,
	}
	return matcher, nil
}
