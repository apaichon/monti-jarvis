package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
)

type kycUpdateRequest struct {
	ContactName    string `json:"contact_name"`
	ContactPhone   string `json:"contact_phone"`
	ContactAddress string `json:"contact_address"`
}

func (s *server) getTenantKYC(w http.ResponseWriter, r *http.Request) {
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	if ac.Role != auth.RoleTenantAdmin || ac.TenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	profile, err := s.store.GetTenantKYCProfile(r.Context(), ac.TenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, kycProfileJSON(profile))
}

func (s *server) updateTenantKYC(w http.ResponseWriter, r *http.Request) {
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	if ac.Role != auth.RoleTenantAdmin || ac.TenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req kycUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	profile, err := s.store.UpsertTenantKYCProfile(r.Context(), ac.TenantID, req.ContactName, req.ContactPhone, req.ContactAddress)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, kycProfileJSON(profile))
}

func (s *server) uploadTenantKYCPhoto(w http.ResponseWriter, r *http.Request) {
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	if ac.Role != auth.RoleTenantAdmin || ac.TenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 6<<20)
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid upload")
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "photo is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 5<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid photo")
		return
	}
	contentType := header.Header.Get("Content-Type")
	key, assetURL, err := s.store.PutKYCPhoto(r.Context(), ac.TenantID, contentType, data)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := s.store.SetTenantKYCPhoto(r.Context(), ac.TenantID, key); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"photo_url": assetURL, "object_key": key})
}

func (s *server) uploadTenantKYCDocument(w http.ResponseWriter, r *http.Request) {
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	if ac.Role != auth.RoleTenantAdmin || ac.TenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid upload")
		return
	}
	file, header, err := r.FormFile("document")
	if err != nil {
		writeError(w, http.StatusBadRequest, "document is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document")
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	key, assetURL, err := s.store.PutKYCDocument(r.Context(), ac.TenantID, header.Filename, contentType, data)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := s.store.AppendTenantKYCDocument(r.Context(), ac.TenantID, key); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"document_url": assetURL, "object_key": key})
}

func (s *server) submitTenantKYC(w http.ResponseWriter, r *http.Request) {
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	if ac.Role != auth.RoleTenantAdmin || ac.TenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	profile, err := s.store.SubmitTenantKYC(r.Context(), ac.TenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, kycProfileJSON(profile))
}

func (s *server) serveKYCAsset(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	kind := strings.TrimSpace(r.PathValue("kind"))
	file := strings.TrimSpace(r.PathValue("file"))
	obj, contentType, err := s.store.GetKYCAsset(r.Context(), tenantID, kind, file)
	if err != nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	defer obj.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = io.Copy(w, obj)
}

func kycProfileJSON(profile store.TenantKYCProfile) map[string]any {
	docs := make([]map[string]string, 0, len(profile.BusinessDocKeys))
	for _, key := range profile.BusinessDocKeys {
		docs = append(docs, map[string]string{
			"object_key": key,
			"url":        store.KYCAssetURL(profile.TenantID, "docs", key[strings.LastIndex(key, "/")+1:]),
		})
	}
	photoURL := ""
	if profile.PhotoObjectKey != "" {
		photoURL = store.KYCAssetURL(profile.TenantID, "photo", profile.PhotoObjectKey[strings.LastIndex(profile.PhotoObjectKey, "/")+1:])
	}
	out := map[string]any{
		"tenant_id":        profile.TenantID,
		"contact_name":     profile.ContactName,
		"contact_phone":    profile.ContactPhone,
		"contact_address":  profile.ContactAddress,
		"photo_url":        photoURL,
		"photo_object_key": profile.PhotoObjectKey,
		"documents":        docs,
		"status":           profile.Status,
		"updated_at":       profile.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if profile.SubmittedAt != nil {
		out["submitted_at"] = profile.SubmittedAt.UTC().Format(time.RFC3339)
	}
	if profile.ReviewedAt != nil {
		out["reviewed_at"] = profile.ReviewedAt.UTC().Format(time.RFC3339)
	}
	if profile.ReviewedBy != "" {
		out["reviewed_by"] = profile.ReviewedBy
	}
	if profile.RejectionReason != "" {
		out["rejection_reason"] = profile.RejectionReason
	}
	return out
}