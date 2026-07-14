package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
)

type patchKnowledgeGapRequest struct {
	Status       string `json:"status"`
	ReviewerNote string `json:"reviewer_note"`
}

func (s *server) listTenantConversationRecords(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	startDate := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endDate := strings.TrimSpace(r.URL.Query().Get("end_date"))
	if startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			writeError(w, http.StatusBadRequest, "start_date must use YYYY-MM-DD")
			return
		}
	}
	if endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			writeError(w, http.StatusBadRequest, "end_date must use YYYY-MM-DD")
			return
		}
	}
	if startDate != "" && endDate != "" && startDate > endDate {
		writeError(w, http.StatusBadRequest, "start_date must be before or equal to end_date")
		return
	}
	rows, err := s.store.ListConversationRecords(r.Context(), tenantID, r.URL.Query().Get("status"), startDate, endDate)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": rows, "next_cursor": nil})
}

func (s *server) getTenantConversationRecord(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rec, err := s.store.GetConversationRecord(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation record not found")
		return
	}
	turns, _ := s.store.ListConversationTranscript(r.Context(), rec.CallID)
	objects, _ := s.store.ListConversationArchiveObjects(r.Context(), tenantID, rec.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"record":          rec,
		"transcript":      turns,
		"archive_objects": objects,
	})
}

func (s *server) retryTenantConversationArchive(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rec, err := s.store.GetConversationRecord(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation record not found")
		return
	}
	payload := map[string]any{"retry": true, "conversation_record_id": rec.ID, "call_id": rec.CallID, "summary": rec.Summary}
	_, err = s.store.ArchiveConversationTranscript(r.Context(), tenantID, rec.CallID, payload, "")
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "retry_queued"})
}

func (s *server) getTenantConversationArchiveObjectContent(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rec, err := s.store.GetConversationRecord(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation record not found")
		return
	}
	obj, contentType, err := s.store.GetConversationArchiveObjectContent(r.Context(), tenantID, rec.ID, r.PathValue("object_id"))
	if err != nil {
		var minioErr minio.ErrorResponse
		if errors.Is(err, pgx.ErrNoRows) || (errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey") {
			writeError(w, http.StatusNotFound, "archive object not found")
			return
		}
		if strings.Contains(err.Error(), "not playable") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer obj.Close()
	if contentType == "" {
		contentType = "audio/wav"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=60")
	_, _ = io.Copy(w, obj)
}

func (s *server) listTenantKnowledgeGaps(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rows, err := s.store.ListKnowledgeGapCandidates(r.Context(), tenantID, r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"gaps": rows})
}

func (s *server) patchTenantKnowledgeGap(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req patchKnowledgeGapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	gap, err := s.store.PatchKnowledgeGapCandidate(r.Context(), tenantID, r.PathValue("id"), strings.TrimSpace(req.Status), req.ReviewerNote)
	if err != nil {
		status := http.StatusBadGateway
		if strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
		}
		if errors.Is(err, pgx.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"gap": gap})
}
