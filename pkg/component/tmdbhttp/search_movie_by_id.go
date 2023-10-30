package tmdbhttp

import (
	"fmt"
)

func SearchMovieByTmdbID(tmdbClient TMDBClient, tmdbID int64) (*TMDBMovieResult, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("client nil")
	}
	if tmdbID == 0 {
		return nil, fmt.Errorf("tmdbid zero")
	}
	url, err := buildDetailMovieURL(tmdbID)
	if err != nil {
		return nil, err
	}
	data := &TMDBMovieResult{}
	err = tmdbClient.DoTmdbHTTP(url, data)
	if err != nil {
		return nil, err
	}
	if data.ID != tmdbID {
		return nil, fmt.Errorf("not found")
	}
	return data, nil
}

func buildDetailMovieURL(movieTMDBID int64) (tmdbURL string, err error) {
	if movieTMDBID == 0 {
		return "", fmt.Errorf("tmdbid zero")
	}
	optionURL := ""
	tmdbURL = fmt.Sprintf("%s%s%d%s",
		baseURL,
		movieURL,
		movieTMDBID,
		optionURL)
	return tmdbURL, nil
}
