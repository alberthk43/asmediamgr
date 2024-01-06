package tmdb

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/proxy"

	tmdb "github.com/cyruzin/golang-tmdb"
)

type Configuration struct {
	Sock5Proxy string
}

type TmdbService struct {
	*tmdb.Client
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
	return &TmdbService{
		Client: tmdbClient,
	}, nil
}
