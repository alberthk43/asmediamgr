package fakes

import (
	"fmt"

	tmdb "github.com/cyruzin/golang-tmdb"

	"asmediamgr/pkg/dirinfo"
	"asmediamgr/pkg/diskop"
)

type FakeDiskOpService struct {
	TvMatching           []tvMatching
	MovieMatching        []movieMatching
	movieSubtileMatching []movieSubtileMatching
	isNeedDelDir         bool
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

type movieSubtileMatching struct {
	entry       *dirinfo.Entry
	old         *dirinfo.File
	movieDetail *tmdb.MovieDetails
	destType    diskop.DestType
	language    string
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

func WithRenameMovieSubtile(entry *dirinfo.Entry, old *dirinfo.File, movieDetail *tmdb.MovieDetails,
	destType diskop.DestType, language string) FakeDiskOpOption {
	return func(s *FakeDiskOpService) {
		s.movieSubtileMatching = append(s.movieSubtileMatching, movieSubtileMatching{
			entry:       entry,
			old:         old,
			movieDetail: movieDetail,
			destType:    destType,
			language:    language,
		})
	}
}

func WithNeedDelDir() FakeDiskOpOption {
	return func(s *FakeDiskOpService) {
		s.isNeedDelDir = true
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

func (dop *FakeDiskOpService) RenameMovieSubtiles(entry *dirinfo.Entry, filesMap map[string][]*dirinfo.File, movieDetail *tmdb.MovieDetails, destType diskop.DestType) error {
	if len(filesMap) == 0 && len(dop.movieSubtileMatching) == 0 {
		return nil
	}
	num := 0
	for language, files := range filesMap {
		for _, file := range files {
			num++
			for _, match := range dop.movieSubtileMatching {
				if match.entry == entry && match.old == file && match.movieDetail == movieDetail && match.destType == destType && match.language == language {
					num--
				}
			}
		}
	}
	if num != 0 {
		return fmt.Errorf("no matching for RenameMovieSubtiles")
	}
	return nil
}

func (dop *FakeDiskOpService) DelDirEntry(entry *dirinfo.Entry) error {
	if !dop.isNeedDelDir {
		return fmt.Errorf("no matching for DelDirEntry")
	}
	return nil
}

func (dop *FakeDiskOpService) RenameTvMusicFile(entry *dirinfo.Entry, old *dirinfo.File, tvDetail *tmdb.TVDetails, name string, destType diskop.DestType) error {
	// TODO current not important, not verified
	return nil
}
