package huanyue

import (
	"asmediamgr/pkg/component/renamer"
	"asmediamgr/pkg/component/tmdbhttp"
	"asmediamgr/pkg/matcher/basic"
)

const (
	fileNameRegexpStr = `^【幻月字幕组】.*日剧.*【(?P<name>.*)】【(?P<ep>\d+)】.*`
)

type HuanYueMatcher struct {
	*basic.BasicSingleFileMatcher
}

func NewHuanYueMatcher(
	tmdbClient tmdbhttp.TMDBClient,
	renamer renamer.Renamer,
	targetPath string,
) (*HuanYueMatcher, error) {
	basic, err := basic.NewBasicSingleFileMatcher(
		fileNameRegexpStr,
		tmdbClient,
		renamer,
		targetPath,
	)
	if err != nil {
		return nil, err
	}
	matcher := &HuanYueMatcher{
		BasicSingleFileMatcher: basic,
	}
	return matcher, nil
}
