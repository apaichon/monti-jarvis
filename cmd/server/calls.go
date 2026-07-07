package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/libra/monti-jarvis/internal/calltypes"
)

type createCallRequest struct {
	TenantID string `json:"tenant_id"`
}

type tokenRequest struct {
	Identity string `json:"identity"`
}

type turnRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *server) createCall(w http.ResponseWriter, r *http.Request) {
	var req createCallRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	sessionID := newID()
	roomName := "monti-" + sessionID[:12]
	session, err := s.calls.Create(r.Context(), sessionID, roomName)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, session)
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