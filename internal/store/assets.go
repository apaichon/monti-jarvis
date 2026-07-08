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
	id := strings.TrimSpace(strings.ToLower(avatarID))
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "jpg"
	}
	return avatarAssetsPrefix + id + "/portrait." + ext
}

func AvatarPortraitURL(avatarID, ext string) string {
	id := strings.TrimSpace(strings.ToLower(avatarID))
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "jpg"
	}
	return "/api/assets/avatars/" + id + "/portrait." + ext
}

func (s *Store) PutAvatarImage(ctx context.Context, avatarID, contentType string, data []byte) (string, string, error) {
	if s.minio == nil {
		return "", "", fmt.Errorf("minio is not available")
	}
	ext := contentTypeToExt(contentType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported image type")
	}
	key := AvatarPortraitKey(avatarID, ext)
	_, err := s.minio.PutObject(ctx, s.cfg.MinioBucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}
	return key, AvatarPortraitURL(avatarID, ext), nil
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
	if !strings.HasPrefix(name, "portrait.") {
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