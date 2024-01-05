package fakes

import (
	"fmt"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
)

type FakeDiskOpService struct {
	Matching []matching
}

type matching struct {
	entry    *dirinfo.Entry
	old      *dirinfo.File
	tvDetail *tmdb.TVDetails
	season   int
	episode  int
	destType diskop.DestType
}

func NewFakeDiskOpService(opts ...FakeDiskOpOption) *FakeDiskOpService {
	ret := &FakeDiskOpService{}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

type FakeDiskOpOption func(*FakeDiskOpService)

func WithRenameSingleTvEpFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails, season int, episode int,
	destType diskop.DestType) FakeDiskOpOption {
	return func(s *FakeDiskOpService) {
		s.Matching = append(s.Matching, matching{
			entry:    entry,
			old:      old,
			tvDetail: tvDetail,
			season:   season,
			episode:  episode,
			destType: destType,
		})
	}
}

func (dop *FakeDiskOpService) RenameSingleTvEpFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails,
	season int, episode int, destType diskop.DestType) error {
	for _, match := range dop.Matching {
		if match.entry == entry && match.old == old && match.tvDetail == tvDetail && match.season == season &&
			match.episode == episode && match.destType == destType {
			return nil
		}
	}
	return fmt.Errorf("no matching for RenameSingleTvEpFile")
}
