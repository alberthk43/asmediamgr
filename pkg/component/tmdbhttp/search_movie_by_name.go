package tmdbhttp

import (
	"fmt"
	"net/url"
)

func SearchOnlyOneMovieByName(
	tmdbClient TMDBClient,
	name string,
	year int,
) (*TMDBMovieResult, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("client nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name empty")
	}
	url, err := buildSearchMovieURL(name, year)
	if err != nil {
		return nil, err
	}
	data := &SearchMoviesResult{}
	err = tmdbClient.DoTmdbHTTP(url, data)
	if err != nil {
		return nil, err
	}
	if len(data.Results) == 0 {
		return nil, fmt.Errorf("not found")
	}
	if len(data.Results) != 1 {
		return nil, fmt.Errorf("not unique")
	}
	return &data.Results[0], nil
}

func buildSearchMovieURL(name string, year int) (tmdbURL string, err error) {
	if name == "" {
		return "", fmt.Errorf("name empty")
	}
	tmdbURL = fmt.Sprintf("%s%smovie?query=%s&include_adult=true&year=%d",
		baseURL,
		searchURL,
		url.QueryEscape(name),
		year)
	return tmdbURL, nil
}
