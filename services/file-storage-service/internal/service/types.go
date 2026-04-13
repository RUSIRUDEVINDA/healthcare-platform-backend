package service

import (
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/file-storage-service/internal/model"
)

type StoredObject struct {
	Provider    model.StorageProvider
	PublicURL   string
	PublicID    string
	R2Bucket    string
	R2ObjectKey string
	StoredName  string
}

type EventPublisher interface {
	PublishFileUploaded(event rabbitmq.FileUploadedEvent) error
	PublishFileDeleted(event rabbitmq.FileDeletedEvent) error
}
