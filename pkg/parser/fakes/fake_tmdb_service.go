package fakes

import (
	"fmt"

	tmdb "github.com/cyruzin/golang-tmdb"
)

type FakeTmdbService struct {
	TvQueryMapping    map[string]*tmdb.SearchTVShows
	TvIdMapping       map[int]*tmdb.TVDetails
	MovieQueryMapping map[string]*tmdb.SearchMovies
	MovieIdMapping    map[int]*tmdb.MovieDetails
}

func NewFakeTmdbService(opts ...FakeTmdbOption) *FakeTmdbService {
	ret := &FakeTmdbService{
		TvQueryMapping:    make(map[string]*tmdb.SearchTVShows),
		TvIdMapping:       make(map[int]*tmdb.TVDetails),
		MovieQueryMapping: make(map[string]*tmdb.SearchMovies),
		MovieIdMapping:    make(map[int]*tmdb.MovieDetails),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

type FakeTmdbOption func(*FakeTmdbService)

func WithTvQueryMapping(query string, searchTvShows *tmdb.SearchTVShows) FakeTmdbOption {
	return func(s *FakeTmdbService) {
		s.TvQueryMapping[query] = searchTvShows
	}
}

func WithTvIdMapping(id int, tvDetails *tmdb.TVDetails) FakeTmdbOption {
	return func(s *FakeTmdbService) {
		s.TvIdMapping[id] = tvDetails
	}
}

func WithMovieQueryMapping(query string, searchMovie *tmdb.SearchMovies) FakeTmdbOption {
	return func(s *FakeTmdbService) {
		s.MovieQueryMapping[query] = searchMovie
	}
}

func WithMovieIdMapping(id int, movieDetails *tmdb.MovieDetails) FakeTmdbOption {
	return func(s *FakeTmdbService) {
		s.MovieIdMapping[id] = movieDetails
	}
}

func (ts *FakeTmdbService) GetSearchTVShow(query string, urlOptions map[string]string) (*tmdb.SearchTVShows, error) {
	if ret, ok := ts.TvQueryMapping[query]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("no matching for GetSearchTVShow")
}

func (ts *FakeTmdbService) GetTVDetails(id int, urlOptions map[string]string) (*tmdb.TVDetails, error) {
	if ret, ok := ts.TvIdMapping[id]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("no matching for GetTVDetails")
}

func (ts *FakeTmdbService) GetMovieDetails(id int, urlOptions map[string]string) (*tmdb.MovieDetails, error) {
	if ret, ok := ts.MovieIdMapping[id]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("no matching for GetMovieDetails")
}

func (ts *FakeTmdbService) GetSearchMovies(query string, urlOptions map[string]string) (*tmdb.SearchMovies, error) {
	if ret, ok := ts.MovieQueryMapping[query]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("no matching for GetSearchMovies")
}
