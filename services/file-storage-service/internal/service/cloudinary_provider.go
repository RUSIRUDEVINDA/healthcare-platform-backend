package service

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"healthcare-platform/services/file-storage-service/internal/config"
	"healthcare-platform/services/file-storage-service/internal/model"
)

type CloudinaryProvider struct {
	cloudName string
	apiKey    string
	apiSecret string
	folder    string
	client    *http.Client
}

type cloudinaryUploadResponse struct {
	PublicID  string `json:"public_id"`
	SecureURL string `json:"secure_url"`
	Bytes     int64  `json:"bytes"`
}

type cloudinaryDeleteResponse struct {
	Result string `json:"result"`
}

func NewCloudinaryProvider(cfg *config.Config) (*CloudinaryProvider, error) {
	if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("cloudinary configuration is incomplete")
	}
	return &CloudinaryProvider{
		cloudName: cfg.CloudinaryCloudName,
		apiKey:    cfg.CloudinaryAPIKey,
		apiSecret: cfg.CloudinaryAPISecret,
		folder:    cfg.CloudinaryUploadFolder,
		client:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *CloudinaryProvider) Upload(fileName string, data []byte, mimeType, ownerID string) (*StoredObject, error) {
	publicID := uuid.NewString()
	folder := fmt.Sprintf("%s/%s", strings.Trim(p.folder, "/"), sanitizeToken(ownerID))
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	params := map[string]string{
		"folder":    folder,
		"public_id": publicID,
		"timestamp": ts,
	}
	signature := signCloudinaryParams(params, p.apiSecret)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("api_key", p.apiKey)
	_ = writer.WriteField("timestamp", ts)
	_ = writer.WriteField("folder", folder)
	_ = writer.WriteField("public_id", publicID)
	_ = writer.WriteField("signature", signature)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", p.cloudName), &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: cloudinary upload failed: %s", ErrStorageUploadFailed, strings.TrimSpace(string(raw)))
	}

	var parsed cloudinaryUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return &StoredObject{
		Provider:   model.StorageProviderCloudinary,
		PublicURL:  parsed.SecureURL,
		PublicID:   parsed.PublicID,
		StoredName: fileName,
	}, nil
}

func (p *CloudinaryProvider) Delete(publicID string) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"public_id": publicID,
		"timestamp": ts,
	}
	signature := signCloudinaryParams(params, p.apiSecret)

	form := url.Values{}
	form.Set("public_id", publicID)
	form.Set("timestamp", ts)
	form.Set("api_key", p.apiKey)
	form.Set("signature", signature)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", p.cloudName), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: cloudinary delete failed: %s", ErrStorageDeleteFailed, strings.TrimSpace(string(raw)))
	}

	var parsed cloudinaryDeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}
	if parsed.Result != "ok" && parsed.Result != "not found" {
		return fmt.Errorf("%w: unexpected cloudinary response: %s", ErrStorageDeleteFailed, parsed.Result)
	}
	return nil
}

func signCloudinaryParams(params map[string]string, apiSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	sum := sha1.Sum([]byte(strings.Join(parts, "&") + apiSecret))
	return hex.EncodeToString(sum[:])
}

func sanitizeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-':
			return r
		default:
			return '-'
		}
	}, value)
	return strings.Trim(value, "-")
}
