package album

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leoluyi/go-api-template/internal/entity"
	"github.com/leoluyi/go-api-template/pkg/dbcontext"
	"github.com/leoluyi/go-api-template/pkg/log"
)

// Repository encapsulates the logic to access albums from the data source.
type Repository interface {
	// Get returns the album with the specified album ID.
	Get(ctx context.Context, id string) (entity.Album, error)
	// Count returns the number of albums.
	Count(ctx context.Context) (int, error)
	// Query returns the list of albums with the given offset and limit.
	Query(ctx context.Context, offset, limit int) ([]entity.Album, error)
	// Create saves a new album in the storage.
	Create(ctx context.Context, album entity.Album) error
	// Update updates the album with given ID in the storage.
	Update(ctx context.Context, album entity.Album) error
	// Delete removes the album with given ID from the storage.
	Delete(ctx context.Context, id string) error
}

// repository persists albums in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new album repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the album with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Album, error) {
	var album entity.Album
	err := sqlx.GetContext(ctx, r.db.With(ctx), &album, "SELECT * FROM album WHERE id=$1", id)
	return album, err
}

// Create saves a new album record in the database.
func (r repository) Create(ctx context.Context, album entity.Album) error {
	_, err := r.db.With(ctx).ExecContext(ctx,
		"INSERT INTO album (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		album.ID, album.Name, album.CreatedAt, album.UpdatedAt,
	)
	return err
}

// Update saves the changes to an album in the database.
func (r repository) Update(ctx context.Context, album entity.Album) error {
	_, err := r.db.With(ctx).ExecContext(ctx,
		"UPDATE album SET name=$1, updated_at=$2 WHERE id=$3",
		album.Name, album.UpdatedAt, album.ID,
	)
	return err
}

// Delete deletes an album with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	album, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	_, err = r.db.With(ctx).ExecContext(ctx, "DELETE FROM album WHERE id=$1", album.ID)
	return err
}

// Count returns the number of the album records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := sqlx.GetContext(ctx, r.db.With(ctx), &count, "SELECT COUNT(*) FROM album")
	return count, err
}

// Query retrieves the album records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, offset, limit int) ([]entity.Album, error) {
	var albums []entity.Album
	err := sqlx.SelectContext(ctx, r.db.With(ctx), &albums,
		"SELECT * FROM album ORDER BY id OFFSET $1 LIMIT $2", offset, limit,
	)
	return albums, err
}
