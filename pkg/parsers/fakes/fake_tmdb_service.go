package fakes

import (
	"fmt"

	tmdb "github.com/cyruzin/golang-tmdb"
)

type FakeTmdbService struct {
	QueryMapping map[string]*tmdb.SearchTVShows
	IdMapping    map[int]*tmdb.TVDetails
}

func NewFakeTmdbService(opts ...FakeTmdbOption) *FakeTmdbService {
	ret := &FakeTmdbService{
		QueryMapping: make(map[string]*tmdb.SearchTVShows),
		IdMapping:    make(map[int]*tmdb.TVDetails),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

type FakeTmdbOption func(*FakeTmdbService)

func WithQueryMapping(query string, searchTvShows *tmdb.SearchTVShows) FakeTmdbOption {
	return func(s *FakeTmdbService) {
		s.QueryMapping[query] = searchTvShows
	}
}

func WithIdMapping(id int, tvDetails *tmdb.TVDetails) FakeTmdbOption {
	return func(s *FakeTmdbService) {
		s.IdMapping[id] = tvDetails
	}
}

func (ts *FakeTmdbService) GetSearchTVShow(query string, urlOptions map[string]string) (*tmdb.SearchTVShows, error) {
	if ret, ok := ts.QueryMapping[query]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("no matching for GetSearchTVShow")
}

func (ts *FakeTmdbService) GetTVDetails(id int, urlOptions map[string]string) (*tmdb.TVDetails, error) {
	if ret, ok := ts.IdMapping[id]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("no matching for GetTVDetails")
}
