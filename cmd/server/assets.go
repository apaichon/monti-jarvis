package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const maxAvatarImageBytes = 4 << 20

func (s *server) uploadAvatarImage(w http.ResponseWriter, r *http.Request) {
	avatarID := strings.TrimSpace(strings.ToLower(r.PathValue("id")))
	if avatarID == "" {
		writeError(w, http.StatusBadRequest, "avatar id is required")
		return
	}
	if err := r.ParseMultipartForm(maxAvatarImageBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarImageBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read file")
		return
	}
	if len(data) > maxAvatarImageBytes {
		writeError(w, http.StatusBadRequest, "image exceeds 4MB limit")
		return
	}

	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = mimeFromFilename(header.Filename)
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	_, imageURL, err := s.store.PutAvatarImage(ctx, avatarID, contentType, data)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported image") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	out := map[string]any{"image_url": imageURL, "status": "uploaded"}
	if av, err := s.store.GetAvatar(ctx, avatarID); err == nil {
		av.ImageURL = imageURL
		if updated, err := s.store.UpdateAvatar(ctx, *av); err == nil {
			out["avatar"] = avatarJSON(*updated)
			out["status"] = "uploaded_and_saved"
		}
	}

	writeJSON(w, http.StatusOK, out)
}

func (s *server) serveAvatarAsset(w http.ResponseWriter, r *http.Request) {
	avatarID := strings.TrimSpace(strings.ToLower(r.PathValue("id")))
	filename := strings.TrimSpace(r.PathValue("file"))
	if avatarID == "" || filename == "" {
		writeError(w, http.StatusBadRequest, "invalid asset path")
		return
	}

	ctx, cancel := contextWithTimeout(r, 15*time.Second)
	defer cancel()

	obj, contentType, err := s.store.GetAvatarAsset(ctx, avatarID, filename)
	if err != nil {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey" {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = io.Copy(w, obj)
}

func mimeFromFilename(name string) string {
	dot := strings.LastIndex(name, ".")
	if dot < 0 {
		return ""
	}
	switch strings.ToLower(name[dot:]) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return ""
	}
}

func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}