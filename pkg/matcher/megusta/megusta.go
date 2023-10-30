package megusta

import (
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher/basic"
)

const (
	fileNameRegexpStr = `^(?P<name>.*)\.S(?P<season>\d+)E(?P<ep>\d+)\..*`
)

type MegustaMatcher struct {
	*basic.BasicSingleFileMatcher
}

func NewMegustaMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*MegustaMatcher, error) {
	basic, err := basic.NewBasicSingleFileMatcher(
		fileNameRegexpStr,
		tmdbClient,
		renamer,
		targetPath,
		basic.OptionalReplaceSearchName(map[string]string{
			".": " ",
		}),
	)
	if err != nil {
		return nil, err
	}
	matcher := &MegustaMatcher{
		BasicSingleFileMatcher: basic,
	}
	return matcher, nil
}
