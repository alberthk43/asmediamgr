package tests

import (
	"asmediamgr/pkg/common"
	"fmt"
)

type MockTmdbService struct {
	ExpectedTv        map[int64]*common.MatchedTV
	ExpectedNameTv    map[string]*common.MatchedTV
	ExpectedMovie     map[int64]*common.MatchedMovie
	ExpectedNameMovie map[string]*common.MatchedMovie
}

func (c *MockTmdbService) SearchTvByTmdbID(tmdbID int64) (tvTmdbInfo *common.MatchedTV, err error) {
	if c.ExpectedTv == nil {
		return nil, fmt.Errorf("mock tmdb service not init")
	}
	if tvTmdbInfo, ok := c.ExpectedTv[tmdbID]; ok {
		return tvTmdbInfo, nil
	}
	return nil, fmt.Errorf("not found")
}

func (c *MockTmdbService) SearchTvByName(name string) (tvTmdbInfo *common.MatchedTV, err error) {
	if c.ExpectedNameTv == nil {
		return nil, fmt.Errorf("mock tmdb service not init")
	}
	if tvTmdbInfo, ok := c.ExpectedNameTv[name]; ok {
		return tvTmdbInfo, nil
	}
	return nil, fmt.Errorf("not found")
}

func (c *MockTmdbService) SearchMovieByTmdbID(tmdbID int64) (*common.MatchedMovie, error) {
	if c.ExpectedMovie == nil {
		return nil, fmt.Errorf("mock tmdb service not init")
	}
	if movieTmdbInfo, ok := c.ExpectedMovie[tmdbID]; ok {
		return movieTmdbInfo, nil
	}
	return nil, fmt.Errorf("not found")
}

func (c *MockTmdbService) SearchMovieByName(name string, year int) (*common.MatchedMovie, error) {
	if c.ExpectedNameMovie == nil {
		return nil, fmt.Errorf("mock tmdb service not init")
	}
	if movieTmdbInfo, ok := c.ExpectedNameMovie[name]; ok {
		if int(movieTmdbInfo.Year) == year {
			return movieTmdbInfo, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
