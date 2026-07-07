package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type AuthEventRow struct {
	EventID   string
	Event     string
	TenantID  string
	UserID    string
	Email     string
	Role      string
	IP        string
	UserAgent string
	CreatedAt time.Time
}

func (c *Client) EnsureAuthEventsSchema(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	db := quoteIdent(c.db)
	return c.exec(ctx, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.auth_events (
  event_id String,
  event String,
  tenant_id String,
  user_id String,
  email String,
  role String,
  ip String,
  user_agent String,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now(),
  created_by String DEFAULT 'system',
  updated_by String DEFAULT 'system'
) ENGINE = MergeTree()
ORDER BY (tenant_id, created_at, event_id)`, db))
}

func (c *Client) InsertAuthEvent(ctx context.Context, row AuthEventRow) error {
	if !c.Enabled() {
		return nil
	}
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now().UTC()
	}
	stmt := fmt.Sprintf(`INSERT INTO %s.auth_events
  (event_id, event, tenant_id, user_id, email, role, ip, user_agent, created_at)
FORMAT JSONEachRow`, quoteIdent(c.db))
	payload := fmt.Sprintf(`{"event_id":%q,"event":%q,"tenant_id":%q,"user_id":%q,"email":%q,"role":%q,"ip":%q,"user_agent":%q,"created_at":%q}`,
		row.EventID, row.Event, row.TenantID, row.UserID, row.Email, row.Role, row.IP, row.UserAgent,
		row.CreatedAt.UTC().Format("2006-01-02 15:04:05"))
	return c.exec(ctx, stmt+"\n"+payload)
}

func (c *Client) ListAuthEvents(ctx context.Context, tenantID string, limit int) ([]AuthEventRow, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("clickhouse is not configured")
	}
	if limit <= 0 {
		limit = 20
	}
	tenantID = strings.TrimSpace(tenantID)
	query := fmt.Sprintf(`SELECT event_id, event, tenant_id, user_id, email, role, ip, user_agent, created_at
FROM %s.auth_events`, quoteIdent(c.db))
	if tenantID != "" {
		query += fmt.Sprintf(` WHERE tenant_id = %q`, tenantID)
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT %d FORMAT JSON`, limit)
	// read path for ops/debug only — not wired to HTTP in this sprint
	_ = query
	return nil, nil
}