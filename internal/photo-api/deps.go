package photoapi

import (
	"context"

	"github.com/google/uuid"

	"photoapi/internal/models"
)

// Storage is an interface to interact with storage.
type Storage interface {
	// MigrateUp performs a database migration to the last available version.
	MigrateUp(ctx context.Context) error
	// UploadPhoto uploads photo to database.
	UploadPhoto(ctx context.Context, photo *models.Photo) (*models.Photo, error)
	// DeletePhoto deletes photo from the database by given id.
	DeletePhoto(ctx context.Context, photoID uuid.UUID) error
	// GetPhotos returns all photos in the database.
	GetPhotos(ctx context.Context) ([]*models.Photo, error)
	// GetPhoto returns a photo from the database by given id.
	GetPhoto(ctx context.Context, id uuid.UUID) (*models.Photo, error)
}
