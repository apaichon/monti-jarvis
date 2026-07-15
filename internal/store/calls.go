package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/calltypes"
)

type CallSessionContext struct {
	Session    calltypes.Session
	CustomerID string
	AvatarID   string
}

func (s *Store) CreateCallSession(ctx context.Context, id, tenantID, roomName string) (calltypes.Session, error) {
	if s.pg == nil {
		return calltypes.Session{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var session calltypes.Session
	actor := auditctx.ActorID(ctx)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.call_sessions (id, tenant_id, room_name, status, created_by, updated_by)
VALUES ($1, $2, $3, 'active', $4, $4)
RETURNING id, tenant_id, room_name, status, started_at`, schema),
		id, tenantID, roomName, actor,
	).Scan(&session.ID, &session.TenantID, &session.RoomName, &session.Status, &session.StartedAt)
	if err != nil {
		return calltypes.Session{}, err
	}

	if s.redis != nil {
		key := s.cfg.RedisPrefix + "call:active:" + id
		_ = s.redis.HSet(ctx, key,
			"tenant_id", tenantID,
			"room_name", roomName,
			"status", "active",
			"started_at", session.StartedAt.UTC().Format(time.RFC3339),
		).Err()
		_ = s.redis.Expire(ctx, key, 24*time.Hour).Err()
	}
	return session, nil
}

func (s *Store) GetCallSession(ctx context.Context, id string) (calltypes.Session, error) {
	if s.pg == nil {
		return calltypes.Session{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var session calltypes.Session
	var endedAt *time.Time
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, room_name, status, started_at, ended_at
FROM %s.call_sessions WHERE id = $1`, schema), id,
	).Scan(&session.ID, &session.TenantID, &session.RoomName, &session.Status, &session.StartedAt, &endedAt)
	if err != nil {
		return calltypes.Session{}, err
	}
	session.EndedAt = endedAt
	return session, nil
}

func (s *Store) GetCallSessionContext(ctx context.Context, id string) (CallSessionContext, error) {
	if s.pg == nil {
		return CallSessionContext{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var out CallSessionContext
	var endedAt *time.Time
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, room_name, status, started_at, ended_at,
       COALESCE(customer_id, ''), COALESCE(avatar_id, '')
FROM %s.call_sessions WHERE id = $1`, schema), id).Scan(
		&out.Session.ID, &out.Session.TenantID, &out.Session.RoomName, &out.Session.Status,
		&out.Session.StartedAt, &endedAt, &out.CustomerID, &out.AvatarID)
	if err != nil {
		return CallSessionContext{}, err
	}
	out.Session.EndedAt = endedAt
	return out, nil
}

func (s *Store) EndCallSession(ctx context.Context, id string) (calltypes.Session, error) {
	if s.pg == nil {
		return calltypes.Session{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var session calltypes.Session
	var endedAt time.Time
	actor := auditctx.ActorID(ctx)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
UPDATE %s.call_sessions
SET status = 'ended', ended_at = now(), updated_by = $2
WHERE id = $1 AND status = 'active'
RETURNING id, tenant_id, room_name, status, started_at, ended_at`, schema), id, actor,
	).Scan(&session.ID, &session.TenantID, &session.RoomName, &session.Status, &session.StartedAt, &endedAt)
	if err != nil {
		return calltypes.Session{}, err
	}
	session.EndedAt = &endedAt

	if s.redis != nil {
		key := s.cfg.RedisPrefix + "call:active:" + id
		_ = s.redis.Del(ctx, key).Err()
	}
	return session, nil
}

func (s *Store) AddCallTurn(ctx context.Context, callID, role, content string) (calltypes.Turn, error) {
	if s.pg == nil {
		return calltypes.Turn{}, fmt.Errorf("postgres is not available")
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "caller"
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var turn calltypes.Turn
	actor := auditctx.ActorID(ctx)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.call_turns (call_id, role, content, created_by, updated_by)
VALUES ($1, $2, $3, $4, $4)
RETURNING id, role, content, created_at`, schema),
		callID, role, content, actor,
	).Scan(&turn.ID, &turn.Role, &turn.Content, &turn.CreatedAt)
	return turn, err
}

func (s *Store) ListCallTurns(ctx context.Context, callID string) ([]calltypes.Turn, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, role, content, created_at
FROM %s.call_turns
WHERE call_id = $1
ORDER BY id ASC`, schema), callID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []calltypes.Turn
	for rows.Next() {
		var turn calltypes.Turn
		if err := rows.Scan(&turn.ID, &turn.Role, &turn.Content, &turn.CreatedAt); err != nil {
			return nil, err
		}
		turns = append(turns, turn)
	}
	return turns, rows.Err()
}
