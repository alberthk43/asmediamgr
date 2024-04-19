package tmdb

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/go-kit/log"
	"golang.org/x/net/proxy"

	tmdb "github.com/cyruzin/golang-tmdb"
)

type Configuration struct {
	Sock5Proxy    string
	ValidCacheDur time.Duration
	Logger        log.Logger
}

type TmdbService struct {
	logger        log.Logger
	httpClient    *tmdb.Client
	cache         *searchCache
	validCacheDur time.Duration
}

func NewTmdbService(c *Configuration) (*TmdbService, error) {
	tmdbClient, err := tmdb.Init(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		return nil, err
	}
	tmdbClient.SetClientAutoRetry()
	customClient := http.Client{
		Timeout: time.Second * 10,
	}
	if c.Sock5Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", c.Sock5Proxy, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}
		httpTransport := &http.Transport{}
		httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		customClient.Transport = httpTransport
	}
	tmdbClient.SetClientConfig(customClient)
	if c.ValidCacheDur == 0 {
		c.ValidCacheDur = DefaultValidCacheDuration
	}
	if c.Logger == nil {
		c.Logger = log.NewNopLogger()
	}
	return &TmdbService{
		logger:        c.Logger,
		httpClient:    tmdbClient,
		cache:         newSearchCache(),
		validCacheDur: c.ValidCacheDur,
	}, nil
}

func (tc *TmdbService) cleanInvalid() {
	now := time.Now()
	for k, v := range tc.cache.movieResults {
		if v.validBefore.Before(now) {
			delete(tc.cache.movieResults, k)
		}
	}
	for k, v := range tc.cache.movieDetails {
		if v.validBefore.Before(now) {
			delete(tc.cache.movieDetails, k)
		}
	}
	for k, v := range tc.cache.tvResults {
		if v.validBefore.Before(now) {
			delete(tc.cache.tvResults, k)
		}
	}
	for k, v := range tc.cache.tvDetails {
		if v.validBefore.Before(now) {
			delete(tc.cache.tvDetails, k)
		}
	}
}

func (tc *TmdbService) GetSearchMovies(query string, urlOptions map[string]string) (*tmdb.SearchMovies, error) {
	tc.cleanInvalid()
	key := buildQueryKey(query, urlOptions)
	if v, ok := tc.cache.movieResults[key]; ok {
		return v.any, nil
	}
	results, err := tc.httpClient.GetSearchMovies(query, urlOptions)
	if err != nil {
		return nil, err
	}
	tc.cache.movieResults[key] = &movieResultsCache{
		validBefore: time.Now().Add(tc.validCacheDur),
		any:         results,
	}
	return results, nil
}

func (tc *TmdbService) GetMovieDetails(id int, urlOptions map[string]string) (*tmdb.MovieDetails, error) {
	tc.cleanInvalid()
	key := buildIdKey(id)
	if v, ok := tc.cache.movieDetails[key]; ok {
		return v.any, nil
	}
	detail, err := tc.httpClient.GetMovieDetails(id, urlOptions)
	if err != nil {
		return nil, err
	}
	tc.cache.movieDetails[key] = &movieDetailCache{
		validBefore: time.Now().Add(tc.validCacheDur),
		any:         detail,
	}
	return detail, nil
}

func (tc *TmdbService) GetSearchTVShow(query string, urlOptions map[string]string) (*tmdb.SearchTVShows, error) {
	tc.cleanInvalid()
	key := buildQueryKey(query, urlOptions)
	if v, ok := tc.cache.tvResults[key]; ok {
		return v.any, nil
	}
	results, err := tc.httpClient.GetSearchTVShow(query, urlOptions)
	if err != nil {
		return nil, err
	}
	tc.cache.tvResults[key] = &tvResultsCache{
		validBefore: time.Now().Add(tc.validCacheDur),
		any:         results,
	}
	return results, err
}

func (tc *TmdbService) GetTVDetails(id int, urlOptions map[string]string) (*tmdb.TVDetails, error) {
	tc.cleanInvalid()
	key := buildIdKey(id)
	if v, ok := tc.cache.tvDetails[key]; ok {
		return v.any, nil
	}
	detail, err := tc.httpClient.GetTVDetails(id, urlOptions)
	if err != nil {
		return nil, err
	}
	tc.cache.tvDetails[key] = &tvDetailCache{
		validBefore: time.Now().Add(tc.validCacheDur),
		any:         detail,
	}
	return detail, err
}

const (
	DefaultValidCacheDuration = time.Hour * 6
)

type idKey struct {
	id int
}

type queryKey struct {
	query        string
	plainUrlOpts string
}

type movieDetailCache struct {
	validBefore time.Time
	any         *tmdb.MovieDetails
}

type movieResultsCache struct {
	validBefore time.Time
	any         *tmdb.SearchMovies
}

type tvDetailCache struct {
	validBefore time.Time
	any         *tmdb.TVDetails
}

type tvResultsCache struct {
	validBefore time.Time
	any         *tmdb.SearchTVShows
}

type searchCache struct {
	movieResults map[queryKey]*movieResultsCache
	movieDetails map[idKey]*movieDetailCache
	tvResults    map[queryKey]*tvResultsCache
	tvDetails    map[idKey]*tvDetailCache
}

func newSearchCache() *searchCache {
	return &searchCache{
		movieResults: make(map[queryKey]*movieResultsCache),
		movieDetails: make(map[idKey]*movieDetailCache),
		tvResults:    make(map[queryKey]*tvResultsCache),
		tvDetails:    make(map[idKey]*tvDetailCache),
	}
}

func buildIdKey(id int) idKey {
	return idKey{
		id: id,
	}
}

func buildQueryKey(query string, urlOptions map[string]string) queryKey {
	// sort urlOptions to make sure the key is unique
	keys := make([]string, 0, len(urlOptions))
	for k := range urlOptions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	str := ""
	for _, k := range keys {
		str += fmt.Sprintf("&%s=%s", k, urlOptions[k])
	}
	return queryKey{
		query:        query,
		plainUrlOpts: str,
	}
}

func BuildTmdbMovieLink(id int) string {
	return fmt.Sprintf("https://www.themoviedb.org/movie/%d", id)
}

func BuildTmdbTvLink(id int) string {
	return fmt.Sprintf("https://www.themoviedb.org/tv/%d", id)
}
