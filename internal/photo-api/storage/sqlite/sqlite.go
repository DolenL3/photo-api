package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"photoapi/internal/config"
	"photoapi/internal/models"
)

// SQLite is a Storage implementation via SQLite.
type SQLite struct {
	db     *sql.DB
	config *config.DBConfig
}

// New returns new Storage implementation via SQLite.
func New(db *sql.DB, config *config.DBConfig) *SQLite {
	return &SQLite{
		db:     db,
		config: config,
	}
}

const (
	maxPreviewWidth = 400
	maxPreviewHight = 400
)

// MigrateUp performs a database migration to the last available version.
func (s *SQLite) MigrateUp(ctx context.Context) error {
	logger := zap.L()
	driver, err := sqlite3.WithInstance(s.db, &sqlite3.Config{})
	if err != nil {
		return errors.Wrap(err, "get sqlite driver")
	}
	m, err := migrate.NewWithDatabaseInstance(
		s.config.MigrationURL,
		"ql", driver,
	)
	if err != nil {
		return errors.Wrap(err, "get migrate instance")
	}
	err = m.Up()

	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return errors.Wrap(err, "migrate up")
		}
		logger.Info("no change during migration")
		return nil
	}
	logger.Info("database migrated successfully")
	return nil
}

// UploadPhoto uploads photo to database.
func (s *SQLite) UploadPhoto(ctx context.Context, photo *models.Photo) (*models.Photo, error) {
	photo.ID = uuid.New()
	var err error
	photo.Preview, err = generatePreview(ctx, photo)
	if err != nil {
		return nil, errors.Wrap(err, "generate preview")
	}

	query := `
	INSERT INTO photo (id, bytes, preview)
	VALUES (?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, photo.ID, photo.Bytes, photo.Preview)
	if err != nil {
		return nil, errors.Wrap(err, "insert into photo")
	}
	return photo, nil
}

// DeletePhoto deletes photo from the database.
func (s *SQLite) DeletePhoto(ctx context.Context, photoID uuid.UUID) error {
	query := `
	DELETE FROM photo
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, photoID)
	if err != nil {
		return errors.Wrap(err, "delete from photo")
	}
	return nil
}

// GetPhotos returns all photos in the database.
func (s *SQLite) GetPhotos(ctx context.Context) ([]*models.Photo, error) {
	query := `
	SELECT id, bytes, preview
	FROM photo
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "select from photo")
	}
	photos := []*models.Photo{}
	for rows.Next() {
		var photo models.Photo
		err = rows.Scan(&photo.ID, &photo.Bytes, &photo.Preview)
		if err != nil {
			return nil, errors.Wrap(err, "scan from photo")
		}
		photos = append(photos, &photo)
	}
	return photos, nil
}

// GetPhoto returns a photo from the database by given id.
func (s *SQLite) GetPhoto(ctx context.Context, id uuid.UUID) (*models.Photo, error) {
	query := `
	SELECT id, bytes, preview
	FROM photo
	WHERE id = ?
	`
	var photo *models.Photo
	err := s.db.QueryRowContext(ctx, query, id).Scan(photo)
	if err != nil {
		return nil, errors.Wrap(err, "select from photo")
	}
	return photo, nil
}

// generatePreview generates and returns a preview for a photo.
// Resulting preview is always jpeg encoded in base64.
func generatePreview(ctx context.Context, photo *models.Photo) (string, error) {
	logger := zap.L()
	logger.Info("gen prev")
	unbased, err := base64.StdEncoding.DecodeString(photo.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "decode base64 image")
	}
	reader := bytes.NewReader(unbased)
	img, _, err := image.Decode(reader)
	if err != nil {
		return "", errors.Wrap(err, "decode image")
	}
	resizedImg := resize.Thumbnail(maxPreviewWidth, maxPreviewHight, img, resize.Lanczos3)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, resizedImg, &jpeg.Options{Quality: 50})
	if err != nil {
		return "", errors.Wrap(err, "encode image as png")
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
