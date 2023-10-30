package tmdbhttp

import (
	"fmt"
	"net/url"
)

func SearchOnlyOneTVByName(tmdbClient TMDBClient, name string) (*TMDBTVResult, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("client nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name empty")
	}
	url, err := buildSearchTVURL(name)
	if err != nil {
		return nil, err
	}
	data := &SearchTvsResult{}
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

func buildSearchTVURL(name string) (tmdbURL string, err error) {
	if name == "" {
		return "", fmt.Errorf("name empty")
	}
	tmdbURL = fmt.Sprintf("%s%stv?query=%s&include_adult=true",
		baseURL,
		searchURL,
		url.QueryEscape(name))
	return tmdbURL, nil
}
