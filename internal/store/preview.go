package store

import (
	"context"
	"fmt"
)

// ensurePreviewSchema adds call_sessions.source for S17 preview tagging.
func (s *Store) ensurePreviewSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	// Idempotent column add (Postgres 9.5+ IF NOT EXISTS on ADD COLUMN).
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
ALTER TABLE %s.call_sessions
  ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'production'`, schema))
	if err != nil {
		return fmt.Errorf("preview schema: %w", err)
	}
	return nil
}

// CreatePreviewCallSession inserts a call session tagged source=preview.
func (s *Store) CreatePreviewCallSession(ctx context.Context, id, tenantID, roomName string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := "preview"
	if id == "" || tenantID == "" {
		return fmt.Errorf("id and tenant_id required")
	}
	if roomName == "" {
		roomName = "preview-" + id
	}
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.call_sessions (id, tenant_id, room_name, status, source, created_by, updated_by)
VALUES ($1, $2, $3, 'active', 'preview', $4, $4)
ON CONFLICT (id) DO NOTHING`, schema), id, tenantID, roomName, actor)
	return err
}
