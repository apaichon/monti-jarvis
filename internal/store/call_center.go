package store

import (
	"context"
	"encoding/json"
	"time"
)

type ConversationAnalyticsContext struct {
	Record          ConversationRecord
	Source          string
	SourceUpdatedAt time.Time
}

func (s *Store) GetConversationAnalyticsContext(ctx context.Context, tenantID, recordID string) (ConversationAnalyticsContext, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, `SELECT r.id,r.tenant_id,r.call_id,r.customer_id,r.avatar_id,COALESCE(a.name,''),r.channel,r.status,r.started_at,r.ended_at,r.duration_seconds,r.summary,
(SELECT COUNT(*) FROM `+schema+`.conversation_archive_objects o WHERE o.conversation_record_id=r.id),
(SELECT COUNT(*) FROM `+schema+`.knowledge_gap_candidates g WHERE g.conversation_record_id=r.id),
COALESCE(cs.source,'production'),r.updated_at
FROM `+schema+`.conversation_records r
LEFT JOIN `+schema+`.ai_avatars a ON a.id=r.avatar_id
LEFT JOIN `+schema+`.call_sessions cs ON cs.id=r.call_id
WHERE r.tenant_id=$1 AND r.id=$2`, tenantID, recordID)
	var out ConversationAnalyticsContext
	if err := scanConversationRecordWithAnalytics(row, &out); err != nil {
		return ConversationAnalyticsContext{}, err
	}
	return out, nil
}

type conversationAnalyticsScanner interface {
	Scan(dest ...any) error
}

func scanConversationRecordWithAnalytics(row conversationAnalyticsScanner, out *ConversationAnalyticsContext) error {
	var raw []byte
	if err := row.Scan(&out.Record.ID, &out.Record.TenantID, &out.Record.CallID, &out.Record.CustomerID, &out.Record.AvatarID, &out.Record.AvatarName, &out.Record.Channel, &out.Record.Status,
		&out.Record.StartedAt, &out.Record.EndedAt, &out.Record.DurationSeconds, &raw, &out.Record.ArchiveObjectCount, &out.Record.KnowledgeGapCount, &out.Source, &out.SourceUpdatedAt); err != nil {
		return err
	}
	out.Record.Summary = map[string]any{}
	_ = jsonUnmarshalStore(raw, &out.Record.Summary)
	return nil
}

// jsonUnmarshalStore keeps this file focused on the query while preserving
// the same permissive summary behavior as conversation record reads.
func jsonUnmarshalStore(raw []byte, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}
