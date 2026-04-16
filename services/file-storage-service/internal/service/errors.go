package service

import "errors"

var (
	ErrInvalidFileType          = errors.New("invalid file type")
	ErrFileTooLarge             = errors.New("file exceeds allowed size")
	ErrProfileImageExists       = errors.New("profile image already exists")
	ErrDocumentAlreadyExists    = errors.New("file already exists")
	ErrUnauthorizedAccess       = errors.New("unauthorized access")
	ErrRelationshipLookupFailed = errors.New("failed to verify patient assignment")
	ErrStorageUploadFailed      = errors.New("storage upload failed")
	ErrStorageDeleteFailed      = errors.New("storage delete failed")
	ErrDownloadUnavailable      = errors.New("download unavailable")
)
