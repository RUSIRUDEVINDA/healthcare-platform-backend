package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/file-storage-service/internal/config"
	"healthcare-platform/services/file-storage-service/internal/model"
	"healthcare-platform/services/file-storage-service/internal/repository"
)

type FileService struct {
	repo                  *repository.FileRepository
	cloud                 *CloudinaryProvider
	r2                    *R2Provider
	publisher             EventPublisher
	log                   *logger.Logger
	cfg                   *config.Config
	appointmentServiceURL string
	httpClient            *http.Client
}

func NewFileService(repo *repository.FileRepository, cloud *CloudinaryProvider, r2 *R2Provider, publisher EventPublisher, log *logger.Logger, cfg *config.Config) *FileService {
	return &FileService{
		repo:                  repo,
		cloud:                 cloud,
		r2:                    r2,
		publisher:             publisher,
		log:                   log,
		cfg:                   cfg,
		appointmentServiceURL: strings.TrimRight(cfg.AppointmentServiceURL, "/"),
		httpClient:            &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *FileService) UploadImage(ctx context.Context, ownerID, uploaderID string, fileHeader *multipart.FileHeader) (*model.FileRecord, error) {
	record, _, err := s.uploadProfileImage(ctx, ownerID, uploaderID, fileHeader, false)
	return record, err
}

func (s *FileService) UpdateImage(ctx context.Context, ownerID, uploaderID string, fileHeader *multipart.FileHeader) (*model.FileRecord, bool, error) {
	return s.uploadProfileImage(ctx, ownerID, uploaderID, fileHeader, true)
}

func (s *FileService) uploadProfileImage(ctx context.Context, ownerID, uploaderID string, fileHeader *multipart.FileHeader, replaceExisting bool) (*model.FileRecord, bool, error) {
	data, mimeType, checksum, err := readAndInspectFile(fileHeader)
	if err != nil {
		return nil, false, err
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, false, ErrInvalidFileType
	}

	existing, err := s.profileImageExists(ctx, ownerID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil && !replaceExisting {
		return nil, false, ErrProfileImageExists
	}

	stored, err := s.cloud.Upload(fileHeader.Filename, data, mimeType, ownerID)
	if err != nil {
		return nil, false, err
	}

	record := &model.FileRecord{
		ID:                 "",
		OwnerID:            ownerID,
		UploaderID:         uploaderID,
		Kind:               model.FileKindImage,
		StorageProvider:    model.StorageProviderCloudinary,
		OriginalName:       fileHeader.Filename,
		StoredName:         stored.StoredName,
		MimeType:           mimeType,
		SizeBytes:          int64(len(data)),
		Checksum:           checksum,
		CloudinaryPublicID: stored.PublicID,
		CloudinaryURL:      stored.PublicURL,
		IsPublic:           true,
	}

	replaced := false
	if replaceExisting && existing != nil {
		oldPublicID := existing.CloudinaryPublicID
		record.ID = existing.ID
		record.CreatedAt = existing.CreatedAt

		if err := s.repo.Update(ctx, record); err != nil {
			_ = s.cloud.Delete(stored.PublicID)
			return nil, false, err
		}

		replaced = true
		if oldPublicID != "" && oldPublicID != stored.PublicID {
			if err := s.cloud.Delete(oldPublicID); err != nil {
				s.log.Warn("Failed to delete previous profile image asset", "owner_id", ownerID, "file_id", record.ID, "error", err)
			}
		}
	} else {
		if err := s.repo.Create(ctx, record); err != nil {
			_ = s.cloud.Delete(stored.PublicID)
			return nil, false, err
		}
	}
	if s.publisher != nil {
		_ = s.publisher.PublishFileUploaded(rabbitmq.FileUploadedEvent{
			FileID:          record.ID,
			OwnerID:         record.OwnerID,
			UploaderID:      record.UploaderID,
			Kind:            string(record.Kind),
			StorageProvider: string(record.StorageProvider),
			OriginalName:    record.OriginalName,
			MimeType:        record.MimeType,
			SizeBytes:       record.SizeBytes,
		})
	}
	return record, replaced, nil
}

func (s *FileService) UploadDocument(ctx context.Context, ownerID, uploaderID string, fileHeader *multipart.FileHeader) (*model.FileRecord, error) {
	data, mimeType, checksum, err := readAndInspectFile(fileHeader)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(mimeType, "image/") {
		return nil, ErrInvalidFileType
	}
	exists, err := s.repo.ExistsByOwnerChecksumKind(ctx, ownerID, checksum, model.FileKindDocument)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDocumentAlreadyExists
	}

	stored, err := s.r2.Upload(fileHeader.Filename, data, mimeType, ownerID)
	if err != nil {
		return nil, err
	}

	record := &model.FileRecord{
		OwnerID:         ownerID,
		UploaderID:      uploaderID,
		Kind:            model.FileKindDocument,
		StorageProvider: model.StorageProviderR2,
		OriginalName:    fileHeader.Filename,
		StoredName:      stored.StoredName,
		MimeType:        mimeType,
		SizeBytes:       int64(len(data)),
		Checksum:        checksum,
		R2Bucket:        stored.R2Bucket,
		R2ObjectKey:     stored.R2ObjectKey,
		IsPublic:        false,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		_ = s.r2.Delete(stored.R2ObjectKey)
		return nil, err
	}
	if s.publisher != nil {
		_ = s.publisher.PublishFileUploaded(rabbitmq.FileUploadedEvent{
			FileID:          record.ID,
			OwnerID:         record.OwnerID,
			UploaderID:      record.UploaderID,
			Kind:            string(record.Kind),
			StorageProvider: string(record.StorageProvider),
			OriginalName:    record.OriginalName,
			MimeType:        record.MimeType,
			SizeBytes:       record.SizeBytes,
		})
	}
	return record, nil
}

func (s *FileService) ListMyFiles(ctx context.Context, ownerID string) ([]model.FileRecord, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

func (s *FileService) profileImageExists(ctx context.Context, ownerID string) (*model.FileRecord, error) {
	files, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Kind == model.FileKindImage && file.StorageProvider == model.StorageProviderCloudinary {
			f := file
			return &f, nil
		}
	}
	return nil, nil
}

func (s *FileService) UploadPatientFile(ctx context.Context, callerID, role, callerToken, patientID string, fileHeader *multipart.FileHeader) (*model.FileRecord, bool, error) {
	if err := s.canAccessPatientFiles(ctx, callerID, role, callerToken, patientID); err != nil {
		return nil, false, err
	}

	data, mimeType, checksum, err := readAndInspectFile(fileHeader)
	if err != nil {
		return nil, false, err
	}

	if strings.HasPrefix(mimeType, "image/") {
		stored, err := s.cloud.Upload(fileHeader.Filename, data, mimeType, patientID)
		if err != nil {
			return nil, false, err
		}

		record := &model.FileRecord{
			OwnerID:            patientID,
			UploaderID:         callerID,
			Kind:               model.FileKindImage,
			StorageProvider:    model.StorageProviderCloudinary,
			OriginalName:       fileHeader.Filename,
			StoredName:         stored.StoredName,
			MimeType:           mimeType,
			SizeBytes:          int64(len(data)),
			Checksum:           checksum,
			CloudinaryPublicID: stored.PublicID,
			CloudinaryURL:      stored.PublicURL,
			IsPublic:           true,
		}

		if err := s.repo.Create(ctx, record); err != nil {
			_ = s.cloud.Delete(stored.PublicID)
			return nil, false, err
		}
		if s.publisher != nil {
			_ = s.publisher.PublishFileUploaded(rabbitmq.FileUploadedEvent{
				FileID:          record.ID,
				OwnerID:         record.OwnerID,
				UploaderID:      record.UploaderID,
				Kind:            string(record.Kind),
				StorageProvider: string(record.StorageProvider),
				OriginalName:    record.OriginalName,
				MimeType:        record.MimeType,
				SizeBytes:       record.SizeBytes,
			})
		}
		return record, false, nil
	}
	exists, err := s.repo.ExistsByOwnerChecksumKind(ctx, patientID, checksum, model.FileKindDocument)
	if err != nil {
		return nil, false, err
	}
	if exists {
		return nil, false, ErrDocumentAlreadyExists
	}

	stored, err := s.r2.Upload(fileHeader.Filename, data, mimeType, patientID)
	if err != nil {
		return nil, false, err
	}

	record := &model.FileRecord{
		OwnerID:         patientID,
		UploaderID:      callerID,
		Kind:            model.FileKindDocument,
		StorageProvider: model.StorageProviderR2,
		OriginalName:    fileHeader.Filename,
		StoredName:      stored.StoredName,
		MimeType:        mimeType,
		SizeBytes:       int64(len(data)),
		Checksum:        checksum,
		R2Bucket:        stored.R2Bucket,
		R2ObjectKey:     stored.R2ObjectKey,
		IsPublic:        false,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		_ = s.r2.Delete(stored.R2ObjectKey)
		return nil, false, err
	}
	if s.publisher != nil {
		_ = s.publisher.PublishFileUploaded(rabbitmq.FileUploadedEvent{
			FileID:          record.ID,
			OwnerID:         record.OwnerID,
			UploaderID:      record.UploaderID,
			Kind:            string(record.Kind),
			StorageProvider: string(record.StorageProvider),
			OriginalName:    record.OriginalName,
			MimeType:        record.MimeType,
			SizeBytes:       record.SizeBytes,
		})
	}
	return record, false, nil
}

func (s *FileService) ListPatientFiles(ctx context.Context, callerID, role, callerToken, patientID string) ([]model.FileRecord, error) {
	if err := s.canAccessPatientFiles(ctx, callerID, role, callerToken, patientID); err != nil {
		return nil, err
	}
	return s.repo.ListByOwner(ctx, patientID)
}

func (s *FileService) GetFile(ctx context.Context, callerID, role, callerToken, fileID string) (*model.FileRecord, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	allowed, err := s.canAccessFile(ctx, callerID, role, callerToken, file)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrUnauthorizedAccess
	}
	return file, nil
}

func (s *FileService) DeleteFile(ctx context.Context, callerID, role, callerToken, fileID string) error {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	allowed, err := s.canAccessFile(ctx, callerID, role, callerToken, file)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrUnauthorizedAccess
	}
	return s.deleteSingleFile(ctx, file)
}

func (s *FileService) DeleteFilesByOwner(ctx context.Context, ownerID string) error {
	files, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return err
	}
	return s.deleteFiles(ctx, files)
}

func (s *FileService) DeleteDocumentsByOwner(ctx context.Context, ownerID string) error {
	files, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return err
	}

	filtered := make([]model.FileRecord, 0, len(files))
	for _, file := range files {
		if file.Kind == model.FileKindDocument {
			filtered = append(filtered, file)
		}
	}
	return s.deleteFiles(ctx, filtered)
}

func (s *FileService) OpenDownload(ctx context.Context, callerID, role, callerToken, fileID string) (*model.FileRecord, io.ReadCloser, string, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, nil, "", err
	}
	allowed, err := s.canAccessFile(ctx, callerID, role, callerToken, file)
	if err != nil {
		return nil, nil, "", err
	}
	if !allowed {
		return nil, nil, "", ErrUnauthorizedAccess
	}
	if file.StorageProvider != model.StorageProviderR2 {
		return file, nil, "", ErrDownloadUnavailable
	}

	reader, contentType, err := s.r2.Download(file.R2ObjectKey)
	if err != nil {
		return nil, nil, "", err
	}
	return file, reader, contentType, nil
}

func (s *FileService) DownloadDocument(ctx context.Context, callerID, role, callerToken, fileID string) (io.ReadCloser, string, error) {
	file, reader, contentType, err := s.OpenDownload(ctx, callerID, role, callerToken, fileID)
	if err != nil {
		return nil, "", err
	}
	if file == nil || reader == nil {
		return nil, "", ErrDownloadUnavailable
	}
	return reader, contentType, nil
}

func (s *FileService) canAccessFile(ctx context.Context, callerID, role, callerToken string, file *model.FileRecord) (bool, error) {
	if role == "admin" {
		return true, nil
	}
	if callerID != "" && callerID == file.OwnerID {
		return true, nil
	}
	if role == "doctor" && callerToken != "" && file.OwnerID != "" {
		allowed, err := s.doctorAssignedToPatient(ctx, callerToken, file.OwnerID)
		if err != nil {
			return false, err
		}
		return allowed, nil
	}
	return false, nil
}

func (s *FileService) canAccessPatientFiles(ctx context.Context, callerID, role, callerToken, patientID string) error {
	if strings.TrimSpace(patientID) == "" {
		return ErrUnauthorizedAccess
	}
	if role == "admin" {
		return nil
	}
	if callerID != "" && callerID == patientID {
		return nil
	}
	if role == "doctor" && callerToken != "" {
		allowed, err := s.doctorAssignedToPatient(ctx, callerToken, patientID)
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
	}
	return ErrUnauthorizedAccess
}

type appointmentSummary struct {
	PatientID string `json:"patient_id"`
}

func (s *FileService) doctorAssignedToPatient(ctx context.Context, callerToken, patientID string) (bool, error) {
	if s.appointmentServiceURL == "" {
		return false, ErrRelationshipLookupFailed
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.appointmentServiceURL+"/api/v1/appointments", nil)
	if err != nil {
		return false, ErrRelationshipLookupFailed
	}
	req.Header.Set("Authorization", "Bearer "+callerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, ErrRelationshipLookupFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, ErrRelationshipLookupFailed
	}

	var appointments []appointmentSummary
	if err := json.NewDecoder(resp.Body).Decode(&appointments); err != nil {
		return false, ErrRelationshipLookupFailed
	}
	for _, appt := range appointments {
		if strings.TrimSpace(appt.PatientID) == strings.TrimSpace(patientID) {
			return true, nil
		}
	}
	return false, nil
}

func (s *FileService) deleteFiles(ctx context.Context, files []model.FileRecord) error {
	for _, file := range files {
		if err := s.deleteSingleFile(ctx, &file); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileService) deleteSingleFile(ctx context.Context, file *model.FileRecord) error {
	switch file.StorageProvider {
	case model.StorageProviderCloudinary:
		if err := s.cloud.Delete(file.CloudinaryPublicID); err != nil {
			return err
		}
	case model.StorageProviderR2:
		if err := s.r2.Delete(file.R2ObjectKey); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown storage provider")
	}

	if err := s.repo.DeleteByID(ctx, file.ID); err != nil {
		return err
	}

	if s.publisher != nil {
		_ = s.publisher.PublishFileDeleted(rabbitmq.FileDeletedEvent{
			FileID:          file.ID,
			OwnerID:         file.OwnerID,
			StorageProvider: string(file.StorageProvider),
		})
	}

	return nil
}

func readAndInspectFile(fileHeader *multipart.FileHeader) ([]byte, string, string, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return nil, "", "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, "", "", err
	}
	if len(data) == 0 {
		return nil, "", "", fmt.Errorf("file is empty")
	}

	mimeType := http.DetectContentType(data)
	sum := sha256.Sum256(data)
	return data, mimeType, hex.EncodeToString(sum[:]), nil
}
