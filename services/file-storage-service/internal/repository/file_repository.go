package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"healthcare-platform/services/file-storage-service/internal/model"
)

var ErrNotFound = errors.New("file not found")

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *model.FileRecord) error {
	now := time.Now().UTC()
	file.ID = uuid.NewString()
	file.CreatedAt = now
	file.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO files (
			id, owner_id, uploader_id, kind, storage_provider,
			original_name, stored_name, mime_type, size_bytes, checksum,
			cloudinary_public_id, cloudinary_url, r2_bucket, r2_object_key,
			is_public, created_at, updated_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18
		)
	`,
		file.ID,
		file.OwnerID,
		file.UploaderID,
		file.Kind,
		file.StorageProvider,
		file.OriginalName,
		file.StoredName,
		file.MimeType,
		file.SizeBytes,
		file.Checksum,
		file.CloudinaryPublicID,
		file.CloudinaryURL,
		file.R2Bucket,
		file.R2ObjectKey,
		file.IsPublic,
		file.CreatedAt,
		file.UpdatedAt,
		file.DeletedAt,
	)
	return err
}

func (r *FileRepository) Update(ctx context.Context, file *model.FileRecord) error {
	now := time.Now().UTC()
	file.UpdatedAt = now
	file.DeletedAt = nil

	res, err := r.db.ExecContext(ctx, `
		UPDATE files
		SET owner_id = $2,
		    uploader_id = $3,
		    kind = $4,
		    storage_provider = $5,
		    original_name = $6,
		    stored_name = $7,
		    mime_type = $8,
		    size_bytes = $9,
		    checksum = $10,
		    cloudinary_public_id = $11,
		    cloudinary_url = $12,
		    r2_bucket = $13,
		    r2_object_key = $14,
		    is_public = $15,
		    updated_at = $16,
		    deleted_at = NULL
		WHERE id = $1 AND deleted_at IS NULL
	`,
		file.ID,
		file.OwnerID,
		file.UploaderID,
		file.Kind,
		file.StorageProvider,
		file.OriginalName,
		file.StoredName,
		file.MimeType,
		file.SizeBytes,
		file.Checksum,
		file.CloudinaryPublicID,
		file.CloudinaryURL,
		file.R2Bucket,
		file.R2ObjectKey,
		file.IsPublic,
		file.UpdatedAt,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*model.FileRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, uploader_id, kind, storage_provider,
		       original_name, stored_name, mime_type, size_bytes, checksum,
		       cloudinary_public_id, cloudinary_url, r2_bucket, r2_object_key,
		       is_public, created_at, updated_at, deleted_at
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	return scanFileRecord(row)
}

func (r *FileRepository) ListByOwner(ctx context.Context, ownerID string) ([]model.FileRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, owner_id, uploader_id, kind, storage_provider,
		       original_name, stored_name, mime_type, size_bytes, checksum,
		       cloudinary_public_id, cloudinary_url, r2_bucket, r2_object_key,
		       is_public, created_at, updated_at, deleted_at
		FROM files
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.FileRecord
	for rows.Next() {
		file, err := scanFileRecord(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, *file)
	}
	return files, rows.Err()
}

func (r *FileRepository) ExistsByOwnerChecksumKind(ctx context.Context, ownerID, checksum string, kind model.FileKind) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM files
			WHERE owner_id = $1
			  AND checksum = $2
			  AND kind = $3
			  AND deleted_at IS NULL
		)
	`, ownerID, checksum, kind).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *FileRepository) DeleteByID(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM files
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func scanFileRecord(scanner interface {
	Scan(dest ...any) error
}) (*model.FileRecord, error) {
	var file model.FileRecord
	var deletedAt sql.NullTime
	var cloudinaryPublicID, cloudinaryURL, r2Bucket, r2ObjectKey sql.NullString

	err := scanner.Scan(
		&file.ID,
		&file.OwnerID,
		&file.UploaderID,
		&file.Kind,
		&file.StorageProvider,
		&file.OriginalName,
		&file.StoredName,
		&file.MimeType,
		&file.SizeBytes,
		&file.Checksum,
		&cloudinaryPublicID,
		&cloudinaryURL,
		&r2Bucket,
		&r2ObjectKey,
		&file.IsPublic,
		&file.CreatedAt,
		&file.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if cloudinaryPublicID.Valid {
		file.CloudinaryPublicID = cloudinaryPublicID.String
	}
	if cloudinaryURL.Valid {
		file.CloudinaryURL = cloudinaryURL.String
	}
	if r2Bucket.Valid {
		file.R2Bucket = r2Bucket.String
	}
	if r2ObjectKey.Valid {
		file.R2ObjectKey = r2ObjectKey.String
	}
	if deletedAt.Valid {
		file.DeletedAt = &deletedAt.Time
	}
	return &file, nil
}
