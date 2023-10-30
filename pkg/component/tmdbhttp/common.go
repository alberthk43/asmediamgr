package tmdbhttp

import (
	"asmediamgr/pkg/common"
	"asmediamgr/pkg/matcher"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

type TMDBMovieResult struct {
	ID               int64  `json:"id"`
	OriginalLanguage string `json:"original_language"`
	OriginalTitle    string `json:"original_title"`
	Adult            bool   `json:"adult"`
	ReleaseDate      string `json:"release_date"`
}

type TMDBTVResult struct {
	ID               int64  `json:"id"`
	OriginalLanguage string `json:"original_language"`
	OriginalName     string `json:"original_name"`
	Adult            bool   `json:"adult"`
	FirstAirDate     string `json:"first_air_date"`
}

type SearchMoviesResult struct {
	Page         int64             `json:"page"`
	TotalResults int64             `json:"total_results"`
	TotalPages   int64             `json:"total_pages"`
	Results      []TMDBMovieResult `json:"results"`
}

type SearchTvsResult struct {
	Page         int64          `json:"page"`
	TotalResults int64          `json:"total_results"`
	TotalPages   int64          `json:"total_pages"`
	Results      []TMDBTVResult `json:"results"`
}

const (
	beginYear = 1900
)

func ConvMovie(data *TMDBMovieResult) (*common.MatchedMovie, error) {
	movieMatched := &common.MatchedMovie{
		MatchedCommon: common.MatchedCommon{
			TmdbID: data.ID,
		},
	}
	movieMatched.OriginalTitle = data.OriginalTitle
	movieMatched.OriginalLanguage = data.OriginalLanguage
	movieMatched.Adult = data.Adult
	time, err := time.Parse("2006-01-02", data.ReleaseDate)
	if err != nil {
		return nil, err
	}
	movieMatched.Year = int32(time.Year())
	if movieMatched.Year < beginYear {
		return nil, fmt.Errorf("year:%d less than %d", movieMatched.Year, beginYear)
	}
	return movieMatched, nil
}

func ConvTV(data *TMDBTVResult) (*common.MatchedTV, error) {
	tvMatched := &common.MatchedTV{
		MatchedCommon: common.MatchedCommon{
			TmdbID: data.ID,
		},
	}
	tvMatched.OriginalTitle = data.OriginalName
	tvMatched.OriginalLanguage = data.OriginalLanguage
	tvMatched.Adult = data.Adult
	time, err := time.Parse("2006-01-02", data.FirstAirDate)
	if err != nil {
		return nil, err
	}
	tvMatched.Year = int32(time.Year())
	if tvMatched.Year < beginYear {
		return nil, fmt.Errorf("year:%d less than %d", tvMatched.Year, beginYear)
	}
	return tvMatched, nil
}

type TMDBClient interface {
	DoTmdbHTTP(url string, data interface{}) error
}

type TmdbHttpClient struct {
	http   http.Client
	apiKey string
}

type Option func(*TmdbHttpClient)

type TmdbService struct {
	httpClient *TmdbHttpClient
}

var _ (matcher.TmdbService) = (*TmdbService)(nil)

// SearchTvByTmdbId search tmdb website database by unique id and return the only result
func (s *TmdbService) SearchTvByTmdbID(tmdbID int64) (*common.MatchedTV, error) {
	r, err := SearchTVByTmdbID(s.httpClient, tmdbID)
	if err != nil {
		return nil, err
	}
	tvTmdbInfo, err := ConvTV(r)
	if err != nil {
		return nil, err
	}
	return tvTmdbInfo, nil
}

// SearchTvByName search tmdb website database by tv name and return the only result
// if there are too many results, return error
func (s *TmdbService) SearchTvByName(name string) (*common.MatchedTV, error) {
	r, err := SearchOnlyOneTVByName(s.httpClient, name)
	if err != nil {
		return nil, err
	}
	tvTmdbInfo, err := ConvTV(r)
	if err != nil {
		return nil, err
	}
	return tvTmdbInfo, nil
}

// SearchMovieByTmdbId search tmdb website database by unique id and return the only result
func (s *TmdbService) SearchMovieByTmdbID(tmdbID int64) (*common.MatchedMovie, error) {
	r, err := SearchMovieByTmdbID(s.httpClient, tmdbID)
	if err != nil {
		return nil, err
	}
	movieTmdbInfo, err := ConvMovie(r)
	if err != nil {
		return nil, err
	}
	return movieTmdbInfo, nil
}

// SearchMovieByName search tmdb website database by tv name and return the only result
// if there are too many results, return error
func (s *TmdbService) SearchMovieByName(name string, year int) (*common.MatchedMovie, error) {
	r, err := SearchOnlyOneMovieByName(s.httpClient, name, year)
	if err != nil {
		return nil, err
	}
	tvTmdbInfo, err := ConvMovie(r)
	if err != nil {
		return nil, err
	}
	return tvTmdbInfo, nil
}

func NewTmdbService(httpClient *TmdbHttpClient) (*TmdbService, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient nil")
	}
	return &TmdbService{
		httpClient: httpClient,
	}, nil
}

func NewTmdbHttpClient(apiKey string, opts ...Option) (*TmdbHttpClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey empty")
	}

	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:54366", nil, proxy.Direct) // TODO
	if err != nil {
		return nil, err
	}
	httpTransport := &http.Transport{}
	httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}

	thc := &TmdbHttpClient{
		http: http.Client{Transport: httpTransport},
		// http:   http.Client{},
		apiKey: apiKey,
	}
	return thc, nil
}

func (c *TmdbHttpClient) DoTmdbHTTP(url string, data interface{}) error {
	if url == "" {
		return fmt.Errorf("url empty")
	}
	if data == nil {
		return fmt.Errorf("data nil")
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http req nil")
	}
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Add("accept", "application/json")
	for {
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusTooManyRequests {
			time.Sleep(10 * time.Second)
			continue
		}
		if res.StatusCode == http.StatusNoContent {
			return fmt.Errorf("nothing")
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("http err, code:%d", res.StatusCode)
		}
		if err := json.NewDecoder(res.Body).Decode(data); err != nil {
			return err
		}
		break
	}
	return nil
}

const (
	baseURL   = "https://api.themoviedb.org/3"
	searchURL = "/search/"
	movieURL  = "/movie/"
	tvURL     = "/tv/"
)
