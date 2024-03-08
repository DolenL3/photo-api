package photoapi

// PhotoService is a service's business logic.
type PhotoService struct {
	Storage Storage
}

// New creates new PhotoService.
func New(storage Storage) *PhotoService {
	return &PhotoService{
		Storage: storage,
	}
}
