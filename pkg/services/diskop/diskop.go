package diskop

type DiskOpService struct {
}

func NewDiskOpService() (*DiskOpService, error) {
	return &DiskOpService{}, nil
}

var (
	defaultDiskOpService *DiskOpService
)

func SetDefault(diskop *DiskOpService) {
	defaultDiskOpService = diskop
}

func GetDefault() *DiskOpService {
	return defaultDiskOpService
}
