package model

import "time"

type FileKind string

type StorageProvider string

const (
	FileKindImage    FileKind = "image"
	FileKindDocument FileKind = "document"
)

const (
	StorageProviderCloudinary StorageProvider = "cloudinary"
	StorageProviderR2         StorageProvider = "r2"
)

func IsValidFileKind(kind FileKind) bool {
	return kind == FileKindImage || kind == FileKindDocument
}

func IsValidStorageProvider(provider StorageProvider) bool {
	return provider == StorageProviderCloudinary || provider == StorageProviderR2
}

type FileRecord struct {
	ID                 string          `json:"id"`
	OwnerID            string          `json:"owner_id"`
	UploaderID         string          `json:"uploader_id"`
	Kind               FileKind        `json:"kind"`
	StorageProvider    StorageProvider `json:"storage_provider"`
	OriginalName       string          `json:"original_name"`
	StoredName         string          `json:"stored_name"`
	MimeType           string          `json:"mime_type"`
	SizeBytes          int64           `json:"size_bytes"`
	Checksum           string          `json:"checksum"`
	CloudinaryPublicID string          `json:"cloudinary_public_id,omitempty"`
	CloudinaryURL      string          `json:"cloudinary_url,omitempty"`
	R2Bucket           string          `json:"r2_bucket,omitempty"`
	R2ObjectKey        string          `json:"r2_object_key,omitempty"`
	IsPublic           bool            `json:"is_public"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          *time.Time      `json:"deleted_at,omitempty"`
}
