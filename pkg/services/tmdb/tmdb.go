package tmdb

type TmdbService struct {
}

func NewTmdbService(configPath string) (*TmdbService, error) {
	return &TmdbService{}, nil
}

var (
	defaultTmdbService *TmdbService
)

func SetDefault(service *TmdbService) {
	defaultTmdbService = service
}

func GetDefault() *TmdbService {
	return defaultTmdbService
}
