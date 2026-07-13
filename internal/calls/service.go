package calls

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/calltypes"
	"github.com/libra/monti-jarvis/internal/lktoken"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/store"
)

type Service struct {
	store    *store.Store
	bus      *natsbus.Bus
	livekit  lktoken.Config
	tenantID string
}

type TokenResponse struct {
	Token    string `json:"token"`
	URL      string `json:"url"`
	Identity string `json:"identity"`
	RoomName string `json:"room_name"`
}

func New(st *store.Store, bus *natsbus.Bus, lk lktoken.Config, tenantID string) *Service {
	return &Service{
		store:    st,
		bus:      bus,
		livekit:  lk,
		tenantID: strings.TrimSpace(tenantID),
	}
}

func (s *Service) Create(ctx context.Context, sessionID, roomName string) (calltypes.Session, error) {
	return s.CreateForTenant(ctx, strings.TrimSpace(s.tenantID), sessionID, roomName)
}

func (s *Service) CreateForTenant(ctx context.Context, tenantID, sessionID, roomName string) (calltypes.Session, error) {
	if s.store == nil {
		return calltypes.Session{}, fmt.Errorf("store is not available")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = strings.TrimSpace(s.tenantID)
	}
	sessionID = strings.TrimSpace(sessionID)
	roomName = strings.TrimSpace(roomName)
	if sessionID == "" || roomName == "" {
		return calltypes.Session{}, fmt.Errorf("session id and room name are required")
	}

	session, err := s.store.CreateCallSession(ctx, sessionID, tenantID, roomName)
	if err != nil {
		return calltypes.Session{}, err
	}

	greeting := "Welcome to Monti. An AI agent will join shortly. Please describe how we can help."
	if _, err := s.store.AddCallTurn(ctx, sessionID, "system", greeting); err == nil {
		_ = s.publishTurn(ctx, sessionID, session.RoomName, "system", greeting)
	}

	_ = s.publish(ctx, "call.started", sessionID, session.RoomName, "", "")
	return session, nil
}

func (s *Service) Get(ctx context.Context, sessionID string) (calltypes.Session, error) {
	return s.store.GetCallSession(ctx, strings.TrimSpace(sessionID))
}

func (s *Service) IssueToken(ctx context.Context, sessionID, identity string) (TokenResponse, error) {
	sessionID = strings.TrimSpace(sessionID)
	identity = strings.TrimSpace(identity)
	if identity == "" {
		if len(sessionID) >= 8 {
			identity = "caller-" + sessionID[:8]
		} else {
			identity = "caller"
		}
	}

	session, err := s.store.GetCallSession(ctx, sessionID)
	if err != nil {
		return TokenResponse{}, err
	}
	if session.Status != "active" {
		return TokenResponse{}, fmt.Errorf("call session is not active")
	}

	token, err := s.livekit.JoinToken(session.RoomName, identity, time.Hour)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		Token:    token,
		URL:      s.livekit.LiveURL,
		Identity: identity,
		RoomName: session.RoomName,
	}, nil
}

func (s *Service) End(ctx context.Context, sessionID string) (calltypes.Session, error) {
	sessionID = strings.TrimSpace(sessionID)
	session, err := s.store.EndCallSession(ctx, sessionID)
	if err != nil {
		return calltypes.Session{}, err
	}

	farewell := "Thank you for calling Monti. This call has ended."
	if _, err := s.store.AddCallTurn(ctx, sessionID, "system", farewell); err == nil {
		_ = s.publishTurn(ctx, sessionID, session.RoomName, "system", farewell)
	}

	_ = s.publish(ctx, "call.ended", sessionID, session.RoomName, "", "")
	return session, nil
}

func (s *Service) AddTurn(ctx context.Context, sessionID, role, content string) (calltypes.Turn, error) {
	sessionID = strings.TrimSpace(sessionID)
	role = strings.TrimSpace(role)
	content = strings.TrimSpace(content)
	if content == "" {
		return calltypes.Turn{}, fmt.Errorf("content is required")
	}
	if role == "" {
		role = "caller"
	}

	session, err := s.store.GetCallSession(ctx, sessionID)
	if err != nil {
		return calltypes.Turn{}, err
	}

	turn, err := s.store.AddCallTurn(ctx, sessionID, role, content)
	if err != nil {
		return calltypes.Turn{}, err
	}
	_ = s.publishTurn(ctx, sessionID, session.RoomName, role, content)
	return turn, nil
}

func (s *Service) ListTurns(ctx context.Context, sessionID string) ([]calltypes.Turn, error) {
	return s.store.ListCallTurns(ctx, strings.TrimSpace(sessionID))
}

func (s *Service) publish(ctx context.Context, event, sessionID, roomName, role, content string) error {
	if s.bus == nil || !s.bus.Enabled() {
		return nil
	}
	return s.bus.PublishCallEvent(ctx, natsbus.Subject(event), natsbus.CallEvent{
		Event:     event,
		SessionID: sessionID,
		TenantID:  s.tenantID,
		RoomName:  roomName,
		Role:      role,
		Content:   content,
		At:        time.Now().UTC(),
	})
}

func (s *Service) publishTurn(ctx context.Context, sessionID, roomName, role, content string) error {
	return s.publish(ctx, "call.turn.created", sessionID, roomName, role, content)
}
