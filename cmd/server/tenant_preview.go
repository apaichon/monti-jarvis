package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/rag"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type previewChatRequest struct {
	SessionID string           `json:"session_id"`
	AgentID   string           `json:"agent_id"`
	Topic     string           `json:"topic"`
	Message   string           `json:"message"`
	History   []gemini.Message `json:"history"`
	// Lang: en | th | auto (optional session language preference)
	Lang string `json:"lang"`
}

// GET /api/tenant/preview/scenarios — static suggested questions (no AI).
func (s *server) listPreviewScenarios(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.tenantIDFromAuth(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"scenarios": []map[string]any{
			{"id": "greeting", "topic": "general", "label": "Greeting", "label_th": "ทักทาย", "question": "Hello, what can you help me with today?"},
			{"id": "hours", "topic": "general", "label": "Business hours", "label_th": "เวลาทำการ", "question": "What are your business hours?"},
			{"id": "billing", "topic": "billing", "label": "Billing FAQ", "label_th": "บิล/ใบแจ้งหนี้", "question": "How can I get a copy of my invoice?"},
			{"id": "price", "topic": "billing", "label": "Pricing", "label_th": "ราคา", "question": "How much does the service cost?"},
			{"id": "tech", "topic": "technical", "label": "Tech reset", "label_th": "รีเซ็ต", "question": "How do I reset my password or device?"},
			{"id": "km_probe", "topic": "general", "label": "KM probe", "label_th": "ทดสอบความรู้", "question": "What products or policies do you know about?"},
		},
	})
}

// POST /api/tenant/preview/chat — tenant-admin preview with package rate limits (same as production chat).
func (s *server) tenantPreviewChat(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req previewChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	agent := workforce.Resolve(req.AgentID)
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = newID()
	}

	message := req.Message
	if topic := strings.TrimSpace(req.Topic); topic != "" && !strings.EqualFold(topic, "general") {
		message = "[" + topic + "] " + message
	}
	history := compactHistory(req.History, message)
	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	// Same package rate limits as production chat.
	if s.quota != nil {
		if err := s.quota.AllowRate(ctx, tenantID, quota.BucketChat); err != nil {
			writeQuotaError(w, err)
			return
		}
	}

	topic := strings.TrimSpace(req.Topic)
	var ragResult rag.Result
	useRAG := true
	if s.quota != nil {
		if err := s.quota.CheckFeature(ctx, tenantID, quota.DimRAGEnabled); err != nil {
			useRAG = false
		}
	}
	ragSvc := s.rag
	if ragSvc != nil && tenantID != "" {
		ragSvc = ragSvc.WithTenant(tenantID)
	}
	if useRAG && ragSvc != nil {
		var err error
		ragResult, err = ragSvc.Retrieve(ctx, agent.ID, topic, req.Message)
		if err != nil {
			log.Printf("preview chat rag tenant=%s agent=%s: %v", tenantID, agent.ID, err)
		}
	}
	prompt := workforce.SystemPrompt(agent)
	lang := strings.ToLower(strings.TrimSpace(req.Lang))
	switch lang {
	case "th":
		prompt += "\n\nReply in Thai (ภาษาไทย) for this conversation unless the user switches language."
	case "en":
		prompt += "\n\nReply in English for this conversation unless the user switches language."
	case "auto":
		prompt += "\n\nDetect the caller's language and reply in that language; you may switch if they switch."
	default:
		if s.store != nil {
			if hint := s.store.AIReplyLocaleHint(ctx, tenantID); hint != "" {
				prompt += "\n\n" + hint
			}
		}
	}
	if useRAG && ragSvc != nil {
		prompt = ragSvc.AugmentPrompt(prompt, agent.ID, topic, req.Message, ragResult)
	}

	if s.ai == nil {
		writeError(w, http.StatusServiceUnavailable, "AI not configured")
		return
	}
	reply, err := s.ai.Reply(ctx, prompt, history)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	if s.store != nil {
		s.store.SaveExchange(context.Background(), sessionID, agent.ID, req.Message, reply)
		// Tag a lightweight session row when possible.
		_ = s.store.CreatePreviewCallSession(context.Background(), sessionID, tenantID, "preview-"+sessionID)
		if ragResult.MissingKM {
			_, _ = s.store.RecordKMGap(context.Background(), store.KMGap{
				TenantID:  tenantID,
				AgentID:   agent.ID,
				Topic:     topic,
				Question:  req.Message,
				SessionID: sessionID,
				Source:    store.KMGapSourceChat,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID,
		"agent_id":   agent.ID,
		"reply":      reply,
		"sources":    ragResult.Sources,
		"missing_km": ragResult.MissingKM,
		"mode":       "preview",
		"tenant_id":  tenantID,
	})
}

// GET /ws/tenant/preview/voice — tenant-admin preview voice.
// Auth: Authorization Bearer or ?access_token=
// Uses the same package quotas as production (rate, concurrent, monthly minutes, S16 daily/per-call).
func (s *server) tenantPreviewVoiceWS(w http.ResponseWriter, r *http.Request) {
	ac, err := s.parseTenantAdminFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID := strings.TrimSpace(ac.TenantID)
	if tenantID == "" {
		writeError(w, http.StatusForbidden, "tenant required")
		return
	}
	if s.store != nil {
		active, aerr := s.store.IsTenantActive(r.Context(), tenantID)
		if aerr != nil {
			writeError(w, http.StatusBadGateway, aerr.Error())
			return
		}
		if !active {
			writeError(w, http.StatusForbidden, "tenant not active")
			return
		}
		// Log preview session (still charged against package).
		_ = s.store.CreatePreviewCallSession(r.Context(), newID(), tenantID, "preview-voice")
	}

	// JWT tenant for RAG (query agent/topic from client).
	q := r.URL.Query()
	q.Set("tenant_id", tenantID)
	r.URL.RawQuery = q.Encode()

	// Same metering as production /ws/voice.
	s.voiceWithPackageQuota(w, r, tenantID)
}

func (s *server) parseTenantAdminFromRequest(r *http.Request) (auth.AuthContext, error) {
	if s.cfg.AuthDisabled {
		return auth.AuthContext{
			UserID:   "preview-dev",
			Role:     auth.RoleTenantAdmin,
			TenantID: s.cfg.DemoTenantID,
		}, nil
	}
	if s.auth == nil || !s.auth.Enabled() {
		return auth.AuthContext{}, auth.ErrNotConfigured
	}
	header := r.Header.Get("Authorization")
	if header == "" {
		if tok := strings.TrimSpace(r.URL.Query().Get("access_token")); tok != "" {
			header = "Bearer " + tok
		}
	}
	ac, err := s.auth.ParseBearer(header)
	if err != nil {
		return auth.AuthContext{}, err
	}
	if ac.Role != auth.RoleTenantAdmin {
		return auth.AuthContext{}, auth.ErrForbidden
	}
	return ac, nil
}
