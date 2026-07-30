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

const avatarAssetsPrefix = "avatars/"

func AvatarPortraitKey(avatarID, ext string) string {
	return AvatarPortraitVariantKey(avatarID, "", ext)
}

func AvatarPortraitVariantKey(avatarID, variant, ext string) string {
	id := strings.TrimSpace(strings.ToLower(avatarID))
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "jpg"
	}
	name := avatarPortraitBaseName(variant)
	return avatarAssetsPrefix + id + "/" + name + "." + ext
}

func AvatarPortraitURL(avatarID, ext string) string {
	return AvatarPortraitVariantURL(avatarID, "", ext)
}

func AvatarPortraitVariantURL(avatarID, variant, ext string) string {
	id := strings.TrimSpace(strings.ToLower(avatarID))
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "jpg"
	}
	name := avatarPortraitBaseName(variant)
	return "/api/assets/avatars/" + id + "/" + name + "." + ext
}

func (s *Store) PutAvatarImage(ctx context.Context, avatarID, contentType string, data []byte) (string, string, error) {
	return s.PutAvatarImageVariant(ctx, avatarID, "", contentType, data)
}

func (s *Store) PutAvatarImageVariant(ctx context.Context, avatarID, variant, contentType string, data []byte) (string, string, error) {
	if s.minio == nil {
		return "", "", fmt.Errorf("minio is not available")
	}
	ext := contentTypeToExt(contentType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported image type")
	}
	key := AvatarPortraitVariantKey(avatarID, variant, ext)
	_, err := s.minio.PutObject(ctx, s.cfg.MinioBucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}
	return key, AvatarPortraitVariantURL(avatarID, variant, ext), nil
}

func (s *Store) GetAvatarAsset(ctx context.Context, avatarID, filename string) (io.ReadCloser, string, error) {
	if s.minio == nil {
		return nil, "", fmt.Errorf("minio is not available")
	}
	id := strings.TrimSpace(strings.ToLower(avatarID))
	name := path.Base(strings.TrimSpace(filename))
	if id == "" || name == "" || name == "." || strings.Contains(name, "..") {
		return nil, "", fmt.Errorf("invalid asset path")
	}
	if !validAvatarPortraitFilename(name) {
		return nil, "", fmt.Errorf("invalid asset filename")
	}
	key := avatarAssetsPrefix + id + "/" + name
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

func avatarPortraitBaseName(variant string) string {
	switch strings.ToLower(strings.TrimSpace(variant)) {
	case "dark":
		return "portrait-dark"
	case "light":
		return "portrait-light"
	default:
		return "portrait"
	}
}

func validAvatarPortraitFilename(name string) bool {
	return strings.HasPrefix(name, "portrait.") ||
		strings.HasPrefix(name, "portrait-dark.") ||
		strings.HasPrefix(name, "portrait-light.")
}

func contentTypeToExt(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	default:
		return ""
	}
}

func extToContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}
