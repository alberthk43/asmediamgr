package fakes

import (
	"fmt"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
)

type FakeDiskOpService struct {
	TvMatching    []tvMatching
	MovieMatching []movieMatching
}

type tvMatching struct {
	entry    *dirinfo.Entry
	old      *dirinfo.File
	tvDetail *tmdb.TVDetails
	season   int
	episode  int
	destType diskop.DestType
}

type movieMatching struct {
	entry       *dirinfo.Entry
	old         *dirinfo.File
	movieDetail *tmdb.MovieDetails
	destType    diskop.DestType
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
		s.TvMatching = append(s.TvMatching, tvMatching{
			entry:    entry,
			old:      old,
			tvDetail: tvDetail,
			season:   season,
			episode:  episode,
			destType: destType,
		})
	}
}

func WithRenameSingleMovieFile(entry *dirinfo.Entry, old *dirinfo.File, movieDetail *tmdb.MovieDetails,
	destType diskop.DestType) FakeDiskOpOption {
	return func(s *FakeDiskOpService) {
		s.MovieMatching = append(s.MovieMatching, movieMatching{
			entry:       entry,
			old:         old,
			movieDetail: movieDetail,
			destType:    destType,
		})
	}
}

func (dop *FakeDiskOpService) RenameSingleTvEpFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails,
	season int, episode int, destType diskop.DestType) error {
	for _, match := range dop.TvMatching {
		if match.entry == entry && match.old == old && match.tvDetail == tvDetail && match.season == season &&
			match.episode == episode && match.destType == destType {
			return nil
		}
	}
	return fmt.Errorf("no matching for RenameSingleTvEpFile")
}

func (dop *FakeDiskOpService) RenameSingleMovieFile(entry *dirinfo.Entry, old *dirinfo.File, movieDetail *tmdb.MovieDetails, destType diskop.DestType) error {
	for _, match := range dop.MovieMatching {
		if match.entry == entry && match.old == old && match.movieDetail == movieDetail && match.destType == destType {
			return nil
		}
	}
	return fmt.Errorf("no matching for RenameSingleMovieFile")
}
