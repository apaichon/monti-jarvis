package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/calltypes"
)

type createCallRequest struct {
	TenantID string `json:"tenant_id"`
	AgentID  string `json:"agent_id"`
}

type tokenRequest struct {
	Identity string `json:"identity"`
}

type turnRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type audioArchiveRequest struct {
	Streams []audioArchiveStream `json:"streams"`
}

type audioArchiveStream struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Data        string `json:"data_base64"`
}

type callRatingRequest struct {
	Score  int    `json:"score"`
	Review string `json:"review"`
}

func (s *server) createCall(w http.ResponseWriter, r *http.Request) {
	var req createCallRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	sessionID := newID()
	roomName := "monti-" + sessionID[:12]
	tenantID := auth.ResolveTenant(r.Context(), req.TenantID, s.cfg.AuthDisabled, s.cfg.DemoTenantID)
	customer, settings, ok := s.enforceCustomerPortalAccess(w, r, tenantID, req.AgentID, "voice")
	if !ok {
		return
	}
	if settings.CustomerMaxCallSeconds > 0 && settings.CustomerDailyCallSeconds > 0 && settings.CustomerMaxCallSeconds > settings.CustomerDailyCallSeconds {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "call duration limit exceeded", "code": "call_duration_limit_exceeded"})
		return
	}
	session, err := s.calls.CreateForTenant(r.Context(), tenantID, sessionID, roomName)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if s.store != nil {
		customerID := ""
		if customer != nil {
			customerID = customer.ID
		}
		// Keep the selected avatar on the call session for both authenticated
		// and anonymous callers. Conversation records are created after the
		// call ends and derive their metadata from this row.
		_ = s.store.UpdateCallSessionContext(r.Context(), session.ID, customerID, req.AgentID)
	}
	if customer != nil {
		_ = s.store.RecordCustomerUsage(r.Context(), tenantID, customer.ID, session.ID, req.AgentID, "voice", 0, "reserved", "")
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *server) archiveCallAudio(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store is not available")
		return
	}
	id := r.PathValue("id")
	session, err := s.calls.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	var req audioArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if len(req.Streams) == 0 {
		writeError(w, http.StatusBadRequest, "audio streams are required")
		return
	}
	if len(req.Streams) > 4 {
		writeError(w, http.StatusBadRequest, "too many audio streams")
		return
	}
	objects := make([]any, 0, len(req.Streams))
	for _, stream := range req.Streams {
		name := strings.TrimSpace(stream.Name)
		if name == "" {
			name = "recording"
		}
		contentType := strings.TrimSpace(stream.ContentType)
		if contentType == "" {
			contentType = "audio/wav"
		}
		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(stream.Data))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid audio data")
			return
		}
		if len(data) > 64*1024*1024 {
			writeError(w, http.StatusRequestEntityTooLarge, "audio stream is too large")
			return
		}
		obj, err := s.store.ArchiveConversationAudio(r.Context(), session.TenantID, id, name, contentType, data, "")
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		objects = append(objects, obj)
	}
	writeJSON(w, http.StatusCreated, map[string]any{"objects": objects})
}

func (s *server) submitCallRating(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store is not available")
		return
	}
	callID := r.PathValue("id")
	tenantID := s.quotaTenant(r)
	if session, err := s.calls.Get(r.Context(), callID); err == nil {
		if session.TenantID != tenantID {
			writeError(w, http.StatusNotFound, "call not found")
			return
		}
		tenantID = session.TenantID
	} else if _, recordErr := s.store.GetConversationRecordByCallID(r.Context(), tenantID, callID); recordErr != nil {
		writeError(w, http.StatusNotFound, "call not found")
		return
	}
	var req callRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Score < 1 || req.Score > 5 {
		writeError(w, http.StatusBadRequest, "score must be between 1 and 5")
		return
	}
	if len([]rune(req.Review)) > 2000 {
		writeError(w, http.StatusBadRequest, "review is too long")
		return
	}
	if err := s.store.SaveConversationRating(r.Context(), tenantID, callID, req.Score, req.Review); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"status": "saved"})
}

func (s *server) getCall(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, err := s.calls.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *server) issueCallToken(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req tokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	token, err := s.calls.IssueToken(r.Context(), id, req.Identity)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s *server) endCall(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, err := s.calls.End(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if s.store != nil {
		turns, _ := s.calls.ListTurns(r.Context(), id)
		payload := map[string]any{"call": session, "turns": turns}
		_, _ = s.store.ArchiveConversationTranscript(r.Context(), session.TenantID, id, payload, "")
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *server) listCallTurns(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	turns, err := s.calls.ListTurns(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if turns == nil {
		turns = []calltypes.Turn{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"turns": turns})
}

func (s *server) addCallTurn(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req turnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	turn, err := s.calls.AddTurn(r.Context(), id, req.Role, req.Content)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, turn)
}

func (s *server) callEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	lastID := int64(0)
	if v := r.URL.Query().Get("after"); v != "" {
		_, _ = fmt.Sscan(v, &lastID)
	}

	for i := 0; i < 120 && r.Context().Err() == nil; i++ {
		turns, err := s.calls.ListTurns(r.Context(), id)
		if err != nil {
			return
		}
		for _, turn := range turns {
			if turn.ID <= lastID {
				continue
			}
			payload, _ := json.Marshal(turn)
			_, _ = w.Write([]byte("event: turn\ndata: " + string(payload) + "\n\n"))
			flusher.Flush()
			lastID = turn.ID
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}
