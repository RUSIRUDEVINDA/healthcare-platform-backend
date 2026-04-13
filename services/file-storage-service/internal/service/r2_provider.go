package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"healthcare-platform/services/file-storage-service/internal/config"
	"healthcare-platform/services/file-storage-service/internal/model"
)

type R2Provider struct {
	accountID       string
	accessKeyID     string
	secretAccessKey string
	bucket          string
	endpoint        string
	client          *http.Client
}

func NewR2Provider(cfg *config.Config) (*R2Provider, error) {
	if cfg.R2AccountID == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" || cfg.R2Bucket == "" {
		return nil, fmt.Errorf("cloudflare r2 configuration is incomplete")
	}

	return &R2Provider{
		accountID:       cfg.R2AccountID,
		accessKeyID:     cfg.R2AccessKeyID,
		secretAccessKey: cfg.R2SecretAccessKey,
		bucket:          cfg.R2Bucket,
		endpoint:        strings.TrimRight(cfg.R2Endpoint, "/"),
		client:          &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *R2Provider) Upload(fileName string, data []byte, mimeType, ownerID string) (*StoredObject, error) {
	objectKey := buildR2ObjectKey(ownerID, fileName)
	req, err := http.NewRequest(http.MethodPut, p.objectURL(objectKey), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mimeType)
	if err := p.signRequest(req, data); err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: r2 upload failed: %s", ErrStorageUploadFailed, strings.TrimSpace(string(raw)))
	}

	return &StoredObject{
		Provider:    model.StorageProviderR2,
		R2Bucket:    p.bucket,
		R2ObjectKey: objectKey,
		StoredName:  fileName,
	}, nil
}

func (p *R2Provider) Delete(objectKey string) error {
	req, err := http.NewRequest(http.MethodDelete, p.objectURL(objectKey), nil)
	if err != nil {
		return err
	}
	if err := p.signRequest(req, nil); err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: r2 delete failed: %s", ErrStorageDeleteFailed, strings.TrimSpace(string(raw)))
	}
	return nil
}

func (p *R2Provider) Download(objectKey string) (io.ReadCloser, string, error) {
	req, err := http.NewRequest(http.MethodGet, p.objectURL(objectKey), nil)
	if err != nil {
		return nil, "", err
	}
	if err := p.signRequest(req, nil); err != nil {
		return nil, "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, "", fmt.Errorf("%w: r2 download failed: %s", ErrDownloadUnavailable, strings.TrimSpace(string(raw)))
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func (p *R2Provider) objectURL(objectKey string) string {
	return fmt.Sprintf("%s/%s/%s", p.endpoint, url.PathEscape(p.bucket), escapeObjectKey(objectKey))
}

func (p *R2Provider) signRequest(req *http.Request, body []byte) error {
	payloadHash := sha256Hex(body)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)

	canonicalURI := canonicalPath(req.URL.EscapedPath())
	canonicalQuery := canonicalQueryString(req.URL.Query())
	canonicalHeaders, signedHeaders := canonicalHeaders(req)
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := fmt.Sprintf("%s/auto/s3/aws4_request", dateStamp)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	signingKey := deriveSigningKey(p.secretAccessKey, dateStamp, "auto", "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))
	authorization := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		p.accessKeyID,
		scope,
		signedHeaders,
		signature,
	)
	req.Header.Set("Authorization", authorization)
	return nil
}

func canonicalHeaders(req *http.Request) (string, string) {
	headers := map[string]string{
		"host":                 req.URL.Host,
		"x-amz-content-sha256": req.Header.Get("X-Amz-Content-Sha256"),
		"x-amz-date":           req.Header.Get("X-Amz-Date"),
	}

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(strings.TrimSpace(headers[key]))
		b.WriteString("\n")
	}
	return b.String(), strings.Join(keys, ";")
}

func canonicalQueryString(values url.Values) string {
	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0)
	for _, key := range keys {
		vals := append([]string(nil), values[key]...)
		sort.Strings(vals)
		for _, val := range vals {
			parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(val))
		}
	}
	return strings.Join(parts, "&")
}

func canonicalPath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func escapeObjectKey(objectKey string) string {
	segments := strings.Split(objectKey, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

func deriveSigningKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func buildR2ObjectKey(ownerID, fileName string) string {
	base := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	base = sanitizeToken(base)
	if base == "" {
		base = "file"
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	return fmt.Sprintf("healthcare-platform/%s/%s/%s-%s%s", sanitizeToken(ownerID), time.Now().UTC().Format("2006/01"), base, uuid.NewString(), ext)
}
