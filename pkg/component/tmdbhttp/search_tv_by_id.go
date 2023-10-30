package tmdbhttp

import (
	"fmt"
)

func SearchTVByTmdbID(tmdbClient TMDBClient, tmdbID int64) (*TMDBTVResult, error) {
	if tmdbClient == nil {
		return nil, fmt.Errorf("client nil")
	}
	if tmdbID == 0 {
		return nil, fmt.Errorf("tmdbid zero")
	}
	url, err := buildDetailTVURL(tmdbID)
	if err != nil {
		return nil, err
	}
	data := &TMDBTVResult{}
	err = tmdbClient.DoTmdbHTTP(url, data)
	if err != nil {
		return nil, err
	}
	if data.ID != tmdbID {
		return nil, fmt.Errorf("not found")
	}
	return data, nil
}

func buildDetailTVURL(tvTMDBID int64) (tmdbURL string, err error) {
	if tvTMDBID == 0 {
		return "", fmt.Errorf("tmdbid zero")
	}
	tmdbURL = fmt.Sprintf("%s%s%d",
		baseURL,
		tvURL,
		tvTMDBID)
	return tmdbURL, nil
}
