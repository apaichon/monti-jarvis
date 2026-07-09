package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
)

const kycAssetsPrefix = "kyc/"

func KYCTenantPrefix(tenantID string) string {
	return kycAssetsPrefix + strings.TrimSpace(strings.ToLower(tenantID)) + "/"
}

func KYCPhotoKey(tenantID, ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "jpg"
	}
	return KYCTenantPrefix(tenantID) + "photo." + ext
}

func KYCDocumentKey(tenantID, filename string) string {
	name := path.Base(strings.TrimSpace(filename))
	if name == "" || name == "." || strings.Contains(name, "..") {
		name = "document.bin"
	}
	return KYCTenantPrefix(tenantID) + "docs/" + name
}

func KYCAssetURL(tenantID, kind, filename string) string {
	id := strings.TrimSpace(strings.ToLower(tenantID))
	switch kind {
	case "photo":
		return "/api/assets/kyc/" + id + "/photo/" + path.Base(filename)
	default:
		return "/api/assets/kyc/" + id + "/docs/" + path.Base(filename)
	}
}

func (s *Store) PutKYCPhoto(ctx context.Context, tenantID, contentType string, data []byte) (string, string, error) {
	if s.minio == nil {
		return "", "", fmt.Errorf("minio is not available")
	}
	ext := contentTypeToExt(contentType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported image type")
	}
	key := KYCPhotoKey(tenantID, ext)
	_, err := s.minio.PutObject(ctx, s.cfg.MinioBucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}
	return key, KYCAssetURL(tenantID, "photo", "photo."+ext), nil
}

func (s *Store) PutKYCDocument(ctx context.Context, tenantID, filename, contentType string, data []byte) (string, string, error) {
	if s.minio == nil {
		return "", "", fmt.Errorf("minio is not available")
	}
	key := KYCDocumentKey(tenantID, filename)
	_, err := s.minio.PutObject(ctx, s.cfg.MinioBucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}
	return key, KYCAssetURL(tenantID, "docs", path.Base(key)), nil
}

func (s *Store) GetKYCAsset(ctx context.Context, tenantID, kind, filename string) (io.ReadCloser, string, error) {
	if s.minio == nil {
		return nil, "", fmt.Errorf("minio is not available")
	}
	id := strings.TrimSpace(strings.ToLower(tenantID))
	name := path.Base(strings.TrimSpace(filename))
	if id == "" || name == "" || name == "." || strings.Contains(name, "..") {
		return nil, "", fmt.Errorf("invalid asset path")
	}
	var key string
	switch kind {
	case "photo":
		key = KYCTenantPrefix(id) + name
	case "docs":
		key = KYCTenantPrefix(id) + "docs/" + name
	default:
		return nil, "", fmt.Errorf("invalid asset kind")
	}
	obj, err := s.minio.GetObject(ctx, s.cfg.MinioBucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	stat, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, "", err
	}
	ct := stat.ContentType
	if ct == "" {
		ct = extToContentType(path.Ext(name))
	}
	return obj, ct, nil
}