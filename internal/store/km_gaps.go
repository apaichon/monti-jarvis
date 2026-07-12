package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auditctx"
)

// KMGap is a customer question the RAG pipeline could not ground (missing KM).
// Tenants use these rows as FAQ backlog to improve the knowledge base.
type KMGap struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	AgentID          string     `json:"agent_id"`
	Topic            string     `json:"topic"`
	Question         string     `json:"question"`
	QuestionHash     string     `json:"question_hash,omitempty"`
	SessionID        string     `json:"session_id,omitempty"`
	CallID           string     `json:"call_id,omitempty"`
	Source           string     `json:"source"` // chat | voice | embed
	Status           string     `json:"status"` // open | resolved | dismissed | converted
	OccurrenceCount  int        `json:"occurrence_count"`
	LastSeenAt       time.Time  `json:"last_seen_at"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	ResolvedDocID    string     `json:"resolved_document_id,omitempty"`
	Notes            string     `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// KMGapSource* and KMGapStatus* constants.
const (
	KMGapSourceChat  = "chat"
	KMGapSourceVoice = "voice"
	KMGapSourceEmbed = "embed"

	KMGapStatusOpen      = "open"
	KMGapStatusResolved  = "resolved"
	KMGapStatusDismissed = "dismissed"
	KMGapStatusConverted = "converted" // linked to a new KM document
)

func (s *Store) ensureKMGapsSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.km_gaps (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  agent_id text NOT NULL DEFAULT '',
  topic text NOT NULL DEFAULT 'general',
  question text NOT NULL,
  question_hash text NOT NULL,
  session_id text NOT NULL DEFAULT '',
  call_id text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT 'chat',
  status text NOT NULL DEFAULT 'open',
  occurrence_count integer NOT NULL DEFAULT 1,
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  resolved_at timestamptz,
  resolved_document_id text NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS km_gaps_tenant_agent_hash_uidx
ON %s.km_gaps (tenant_id, agent_id, question_hash)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS km_gaps_tenant_status_idx
ON %s.km_gaps (tenant_id, status, last_seen_at DESC)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("km_gaps schema: %w", err)
		}
	}
	return nil
}

// NormalizeQuestionHash returns sha256 hex of lowercased trimmed question.
func NormalizeQuestionHash(question string) string {
	q := strings.ToLower(strings.TrimSpace(question))
	sum := sha256.Sum256([]byte(q))
	return hex.EncodeToString(sum[:])
}

// RecordKMGap inserts or bumps occurrence for a missing-KM question (best-effort).
// No-op when Postgres is unavailable or question is empty.
func (s *Store) RecordKMGap(ctx context.Context, gap KMGap) (*KMGap, error) {
	if s.pg == nil {
		return nil, nil
	}
	gap.TenantID = strings.TrimSpace(gap.TenantID)
	gap.Question = strings.TrimSpace(gap.Question)
	if gap.TenantID == "" || gap.Question == "" {
		return nil, nil
	}
	gap.AgentID = strings.ToLower(strings.TrimSpace(gap.AgentID))
	gap.Topic = strings.TrimSpace(gap.Topic)
	if gap.Topic == "" {
		gap.Topic = "general"
	}
	gap.Source = strings.TrimSpace(gap.Source)
	if gap.Source == "" {
		gap.Source = KMGapSourceChat
	}
	gap.QuestionHash = NormalizeQuestionHash(gap.Question)
	if gap.ID == "" {
		gap.ID = newStoreID()
	}
	if gap.Status == "" {
		gap.Status = KMGapStatusOpen
	}
	actor := auditctx.ActorID(ctx)
	if actor == "" {
		actor = "system"
	}
	schema := quoteIdent(s.cfg.PostgresSchema)

	// Upsert: same tenant+agent+question hash → increment count, refresh last_seen.
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.km_gaps (
  id, tenant_id, agent_id, topic, question, question_hash,
  session_id, call_id, source, status, occurrence_count, last_seen_at,
  created_by, updated_by
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,1,now(),$11,$11
)
ON CONFLICT (tenant_id, agent_id, question_hash) DO UPDATE SET
  occurrence_count = %s.km_gaps.occurrence_count + 1,
  last_seen_at = now(),
  topic = EXCLUDED.topic,
  session_id = CASE WHEN EXCLUDED.session_id <> '' THEN EXCLUDED.session_id ELSE %s.km_gaps.session_id END,
  call_id = CASE WHEN EXCLUDED.call_id <> '' THEN EXCLUDED.call_id ELSE %s.km_gaps.call_id END,
  source = EXCLUDED.source,
  updated_at = now(),
  updated_by = EXCLUDED.updated_by,
  -- Re-open if previously dismissed/resolved and question appears again
  status = CASE
    WHEN %s.km_gaps.status IN ('dismissed') THEN 'open'
    ELSE %s.km_gaps.status
  END
RETURNING id, tenant_id, agent_id, topic, question, question_hash,
  session_id, call_id, source, status, occurrence_count, last_seen_at,
  resolved_at, COALESCE(resolved_document_id,''), COALESCE(notes,''),
  created_at, updated_at`,
		schema, schema, schema, schema, schema, schema),
		gap.ID, gap.TenantID, gap.AgentID, gap.Topic, gap.Question, gap.QuestionHash,
		gap.SessionID, gap.CallID, gap.Source, gap.Status, actor,
	)

	var out KMGap
	var resolvedAt *time.Time
	err := row.Scan(
		&out.ID, &out.TenantID, &out.AgentID, &out.Topic, &out.Question, &out.QuestionHash,
		&out.SessionID, &out.CallID, &out.Source, &out.Status, &out.OccurrenceCount, &out.LastSeenAt,
		&resolvedAt, &out.ResolvedDocID, &out.Notes,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	out.ResolvedAt = resolvedAt
	return &out, nil
}

// ListKMGaps returns gaps for a tenant, newest activity first.
func (s *Store) ListKMGaps(ctx context.Context, tenantID, status, agentID string, limit int) ([]KMGap, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id required")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	args := []any{tenantID}
	where := []string{"tenant_id = $1"}
	n := 2
	if st := strings.TrimSpace(status); st != "" {
		where = append(where, fmt.Sprintf("status = $%d", n))
		args = append(args, st)
		n++
	}
	if ag := strings.ToLower(strings.TrimSpace(agentID)); ag != "" {
		where = append(where, fmt.Sprintf("agent_id = $%d", n))
		args = append(args, ag)
		n++
	}
	args = append(args, limit)
	q := fmt.Sprintf(`
SELECT id, tenant_id, agent_id, topic, question, question_hash,
  session_id, call_id, source, status, occurrence_count, last_seen_at,
  resolved_at, COALESCE(resolved_document_id,''), COALESCE(notes,''),
  created_at, updated_at
FROM %s.km_gaps
WHERE %s
ORDER BY last_seen_at DESC
LIMIT $%d`, schema, strings.Join(where, " AND "), n)

	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []KMGap
	for rows.Next() {
		var g KMGap
		var resolvedAt *time.Time
		if err := rows.Scan(
			&g.ID, &g.TenantID, &g.AgentID, &g.Topic, &g.Question, &g.QuestionHash,
			&g.SessionID, &g.CallID, &g.Source, &g.Status, &g.OccurrenceCount, &g.LastSeenAt,
			&resolvedAt, &g.ResolvedDocID, &g.Notes,
			&g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		g.ResolvedAt = resolvedAt
		out = append(out, g)
	}
	return out, rows.Err()
}

// UpdateKMGapStatus sets status/notes for a tenant-owned gap.
func (s *Store) UpdateKMGapStatus(ctx context.Context, tenantID, id, status, notes, resolvedDocID string) (*KMGap, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	status = strings.TrimSpace(status)
	switch status {
	case KMGapStatusOpen, KMGapStatusResolved, KMGapStatusDismissed, KMGapStatusConverted:
	default:
		return nil, fmt.Errorf("invalid status")
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var resolvedAt any
	if status == KMGapStatusResolved || status == KMGapStatusConverted || status == KMGapStatusDismissed {
		resolvedAt = time.Now().UTC()
	}

	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
UPDATE %s.km_gaps SET
  status = $3,
  notes = COALESCE(NULLIF($4,''), notes),
  resolved_document_id = CASE WHEN $5 <> '' THEN $5 ELSE resolved_document_id END,
  resolved_at = CASE WHEN $6::timestamptz IS NOT NULL THEN $6::timestamptz ELSE resolved_at END,
  updated_at = now(),
  updated_by = $7
WHERE tenant_id = $1 AND id = $2
RETURNING id, tenant_id, agent_id, topic, question, question_hash,
  session_id, call_id, source, status, occurrence_count, last_seen_at,
  resolved_at, COALESCE(resolved_document_id,''), COALESCE(notes,''),
  created_at, updated_at`, schema),
		tenantID, id, status, notes, resolvedDocID, resolvedAt, actor,
	)
	var g KMGap
	var ra *time.Time
	err := row.Scan(
		&g.ID, &g.TenantID, &g.AgentID, &g.Topic, &g.Question, &g.QuestionHash,
		&g.SessionID, &g.CallID, &g.Source, &g.Status, &g.OccurrenceCount, &g.LastSeenAt,
		&ra, &g.ResolvedDocID, &g.Notes,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	g.ResolvedAt = ra
	return &g, nil
}
