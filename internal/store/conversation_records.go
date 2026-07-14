package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/minio/minio-go/v7"
)

type CustomerUsageSummary struct {
	TenantID              string    `json:"tenant_id"`
	CustomerID            string    `json:"customer_id,omitempty"`
	DailyRemainingSeconds *int      `json:"daily_remaining_seconds,omitempty"`
	DailyLimitSeconds     int       `json:"daily_limit_seconds"`
	MaxCallSeconds        int       `json:"max_call_seconds"`
	UsedSeconds           int       `json:"used_seconds"`
	ResetAt               time.Time `json:"reset_at"`
	State                 string    `json:"state"`
}

type ConversationRecord struct {
	ID                 string         `json:"id"`
	TenantID           string         `json:"tenant_id,omitempty"`
	CallID             string         `json:"call_id,omitempty"`
	CustomerID         string         `json:"customer_id,omitempty"`
	AvatarID           string         `json:"avatar_id,omitempty"`
	AvatarName         string         `json:"avatar_name,omitempty"`
	Channel            string         `json:"channel"`
	Status             string         `json:"status"`
	StartedAt          time.Time      `json:"started_at"`
	EndedAt            *time.Time     `json:"ended_at,omitempty"`
	DurationSeconds    int            `json:"duration_seconds"`
	Summary            map[string]any `json:"summary,omitempty"`
	ArchiveObjectCount int            `json:"archive_object_count,omitempty"`
	KnowledgeGapCount  int            `json:"knowledge_gap_count,omitempty"`
}

type ConversationArchiveObject struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id,omitempty"`
	ConversationRecordID string    `json:"conversation_record_id"`
	ObjectKey            string    `json:"object_key,omitempty"`
	ObjectType           string    `json:"object_type"`
	ContentType          string    `json:"content_type"`
	SizeBytes            int64     `json:"size_bytes"`
	ChecksumSHA256       string    `json:"checksum_sha256"`
	ProtectionMode       string    `json:"protection_mode"`
	Status               string    `json:"status"`
	ErrorCode            string    `json:"error_code,omitempty"`
	StoredAt             time.Time `json:"stored_at,omitempty"`
}

type ConversationTranscriptLine struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type KnowledgeGapCandidate struct {
	ID                   string     `json:"id"`
	TenantID             string     `json:"tenant_id,omitempty"`
	ConversationRecordID string     `json:"conversation_record_id,omitempty"`
	AvatarID             string     `json:"avatar_id,omitempty"`
	CustomerID           string     `json:"customer_id,omitempty"`
	SourceTurnID         string     `json:"source_turn_id,omitempty"`
	Question             string     `json:"question"`
	AnswerExcerpt        string     `json:"answer_excerpt,omitempty"`
	GapReason            string     `json:"gap_reason"`
	Confidence           float64    `json:"confidence"`
	Status               string     `json:"status"`
	ReviewerNote         string     `json:"reviewer_note,omitempty"`
	SnoozedUntil         *time.Time `json:"snoozed_until,omitempty"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at,omitempty"`
	UpdatedAt            time.Time  `json:"updated_at,omitempty"`
}

type ConversationRecordInput struct {
	TenantID        string
	CallID          string
	CustomerID      string
	AvatarID        string
	Channel         string
	Status          string
	StartedAt       time.Time
	EndedAt         *time.Time
	DurationSeconds int
	Summary         map[string]any
}

type KnowledgeGapInput struct {
	TenantID             string
	ConversationRecordID string
	AvatarID             string
	CustomerID           string
	SourceTurnID         string
	Question             string
	AnswerExcerpt        string
	GapReason            string
	Confidence           float64
}

func (s *Store) ensureConversationRecordsSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`ALTER TABLE %s.call_sessions ADD COLUMN IF NOT EXISTS customer_id text`, schema),
		fmt.Sprintf(`ALTER TABLE %s.call_sessions ADD COLUMN IF NOT EXISTS avatar_id text`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_usage_events (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  customer_id text NOT NULL DEFAULT '',
  session_id text NOT NULL DEFAULT '',
  avatar_id text NOT NULL DEFAULT '',
  usage_type text NOT NULL CHECK (usage_type IN ('chat','voice')),
  reserved_seconds integer NOT NULL DEFAULT 0,
  consumed_seconds integer NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'committed' CHECK (status IN ('reserved','committed','released','denied')),
  deny_reason text NOT NULL DEFAULT '',
  usage_date date NOT NULL DEFAULT CURRENT_DATE,%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customer_usage_events_daily_idx ON %s.customer_usage_events (tenant_id, customer_id, usage_date)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.conversation_records (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  call_id text NOT NULL DEFAULT '',
  customer_id text NOT NULL DEFAULT '',
  avatar_id text NOT NULL DEFAULT '',
  channel text NOT NULL CHECK (channel IN ('chat','voice')),
  status text NOT NULL DEFAULT 'recording' CHECK (status IN ('recording','archived','archive_failed')),
  started_at timestamptz NOT NULL DEFAULT now(),
  ended_at timestamptz,
  duration_seconds integer NOT NULL DEFAULT 0,
  summary jsonb NOT NULL DEFAULT '{}'::jsonb,%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS conversation_records_tenant_idx ON %s.conversation_records (tenant_id, started_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.conversation_ratings (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  call_id text NOT NULL,
  conversation_record_id text NOT NULL DEFAULT '',
  customer_id text NOT NULL DEFAULT '',
  avatar_id text NOT NULL DEFAULT '',
  channel text NOT NULL DEFAULT 'voice' CHECK (channel IN ('chat','voice')),
  score integer NOT NULL CHECK (score BETWEEN 1 AND 5),
  review text NOT NULL DEFAULT '',
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, call_id)
)`, schema),
		fmt.Sprintf(`ALTER TABLE %s.conversation_ratings
ADD COLUMN IF NOT EXISTS conversation_record_id text NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS customer_id text NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS avatar_id text NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS channel text NOT NULL DEFAULT 'voice',
ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system',
ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system'`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS conversation_ratings_tenant_date_idx
ON %s.conversation_ratings (tenant_id, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS conversation_ratings_tenant_dimensions_idx
ON %s.conversation_ratings (tenant_id, avatar_id, channel, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.conversation_archive_objects (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  conversation_record_id text NOT NULL REFERENCES %s.conversation_records(id) ON DELETE CASCADE,
  object_key text NOT NULL DEFAULT '',
  object_type text NOT NULL CHECK (object_type IN ('transcript','audio','metadata')),
  content_type text NOT NULL DEFAULT 'application/json',
  size_bytes bigint NOT NULL DEFAULT 0,
  checksum_sha256 text NOT NULL DEFAULT '',
  protection_mode text NOT NULL DEFAULT 'none',
  status text NOT NULL DEFAULT 'stored' CHECK (status IN ('stored','failed','deleted')),
  error_code text NOT NULL DEFAULT '',
  stored_at timestamptz,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.knowledge_gap_candidates (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  conversation_record_id text NOT NULL DEFAULT '',
  avatar_id text NOT NULL DEFAULT '',
  customer_id text NOT NULL DEFAULT '',
  source_turn_id text NOT NULL DEFAULT '',
  question text NOT NULL,
  answer_excerpt text NOT NULL DEFAULT '',
  gap_reason text NOT NULL DEFAULT 'no_source',
  confidence numeric NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','snoozed','resolved','ignored')),
  reviewer_note text NOT NULL DEFAULT '',
  snoozed_until timestamptz,
  resolved_at timestamptz,%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS knowledge_gap_candidates_tenant_idx ON %s.knowledge_gap_candidates (tenant_id, status, created_at DESC)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("conversation records schema: %w", err)
		}
	}
	return nil
}

func (s *Store) SaveConversationRating(ctx context.Context, tenantID, callID string, score int, review string) error {
	if s == nil || s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	if score < 1 || score > 5 {
		return fmt.Errorf("score must be between 1 and 5")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var conversationRecordID, customerID, avatarID, channel, recordStatus string
	recordErr := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,customer_id,avatar_id,channel,status
FROM %s.conversation_records WHERE tenant_id=$1 AND call_id=$2 ORDER BY created_at DESC LIMIT 1`, schema), tenantID, callID).
		Scan(&conversationRecordID, &customerID, &avatarID, &channel, &recordStatus)
	if recordErr == nil {
		if recordStatus != "archived" {
			return fmt.Errorf("conversation has not ended")
		}
	} else {
		var startedAt time.Time
		var endedAt *time.Time
		customerID, avatarID, startedAt, endedAt, _ = s.conversationContextForCall(ctx, callID)
		if startedAt.IsZero() {
			return fmt.Errorf("call session not found")
		}
		if endedAt == nil {
			return fmt.Errorf("call session has not ended")
		}
		channel = "voice"
	}
	if channel != "chat" && channel != "voice" {
		channel = "voice"
	}
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.conversation_ratings
(id,tenant_id,call_id,conversation_record_id,customer_id,avatar_id,channel,score,review,created_at,updated_at,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,now(),now(),$10,$10)
ON CONFLICT (tenant_id,call_id) DO UPDATE SET
conversation_record_id=COALESCE(NULLIF(EXCLUDED.conversation_record_id,''),conversation_ratings.conversation_record_id),
customer_id=COALESCE(NULLIF(EXCLUDED.customer_id,''),conversation_ratings.customer_id),
avatar_id=COALESCE(NULLIF(EXCLUDED.avatar_id,''),conversation_ratings.avatar_id),
channel=EXCLUDED.channel,score=EXCLUDED.score,review=EXCLUDED.review,updated_at=now(),updated_by=EXCLUDED.updated_by`, schema),
		"crat_"+newStoreID(), tenantID, callID, conversationRecordID, customerID, avatarID, channel, score, trimBounded(review, 2000), actor)
	return err
}

func (s *Store) CustomerUsageSummary(ctx context.Context, tenantID, customerID string, dailyLimitSeconds, maxCallSeconds int, now time.Time) (CustomerUsageSummary, error) {
	if now.IsZero() {
		now = time.Now()
	}
	resetAt := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	out := CustomerUsageSummary{TenantID: tenantID, CustomerID: customerID, DailyLimitSeconds: dailyLimitSeconds, MaxCallSeconds: maxCallSeconds, ResetAt: resetAt, State: "quota_available"}
	if strings.TrimSpace(customerID) == "" || dailyLimitSeconds <= 0 || s.pg == nil {
		return out, nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	usageDate := now.Format("2006-01-02")
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT COALESCE(SUM(consumed_seconds),0)
FROM %s.customer_usage_events WHERE tenant_id=$1 AND customer_id=$2 AND usage_date=$3 AND status='committed'`, schema),
		tenantID, customerID, usageDate).Scan(&out.UsedSeconds)
	if err != nil {
		return out, err
	}
	remaining := dailyLimitSeconds - out.UsedSeconds
	if remaining < 0 {
		remaining = 0
	}
	out.DailyRemainingSeconds = &remaining
	if remaining <= 0 {
		out.State = "quota_exhausted"
	}
	return out, nil
}

func (s *Store) RecordCustomerUsage(ctx context.Context, tenantID, customerID, sessionID, avatarID, usageType string, consumedSeconds int, status, denyReason string) error {
	if s == nil || s.pg == nil || strings.TrimSpace(customerID) == "" {
		return nil
	}
	if consumedSeconds < 0 {
		consumedSeconds = 0
	}
	if status == "" {
		status = "committed"
	}
	id := "cuevt_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_usage_events
(id,tenant_id,customer_id,session_id,avatar_id,usage_type,consumed_seconds,status,deny_reason,usage_date,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,CURRENT_DATE,$10,$10)`, schema),
		id, tenantID, customerID, sessionID, avatarID, usageType, consumedSeconds, status, denyReason, actor)
	return err
}

func (s *Store) UpdateCallSessionContext(ctx context.Context, callID, customerID, avatarID string) error {
	if s == nil || s.pg == nil || strings.TrimSpace(callID) == "" {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.call_sessions SET customer_id=$2, avatar_id=$3, updated_at=now() WHERE id=$1`, schema),
		callID, strings.TrimSpace(customerID), strings.TrimSpace(avatarID))
	return err
}

func (s *Store) UpsertConversationRecord(ctx context.Context, in ConversationRecordInput) (ConversationRecord, error) {
	if s == nil || s.pg == nil {
		return ConversationRecord{}, fmt.Errorf("postgres is not available")
	}
	id := strings.TrimSpace(in.CallID)
	if id == "" {
		id = "crec_" + newStoreID()
	} else {
		id = "crec_" + id
	}
	if in.Channel == "" {
		in.Channel = "chat"
	}
	if in.Status == "" {
		in.Status = "recording"
	}
	if in.StartedAt.IsZero() {
		in.StartedAt = time.Now().UTC()
	}
	raw, _ := json.Marshal(in.Summary)
	if in.Summary == nil {
		raw = []byte("{}")
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.conversation_records
(id,tenant_id,call_id,customer_id,avatar_id,channel,status,started_at,ended_at,duration_seconds,summary,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)
ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status,ended_at=EXCLUDED.ended_at,
duration_seconds=EXCLUDED.duration_seconds,
customer_id=COALESCE(NULLIF(EXCLUDED.customer_id,''),conversation_records.customer_id),
avatar_id=COALESCE(NULLIF(EXCLUDED.avatar_id,''),conversation_records.avatar_id),
summary=EXCLUDED.summary,updated_by=EXCLUDED.updated_by,updated_at=now()`, schema),
		id, in.TenantID, in.CallID, in.CustomerID, in.AvatarID, in.Channel, in.Status, in.StartedAt, in.EndedAt, in.DurationSeconds, raw, actor)
	if err != nil {
		return ConversationRecord{}, err
	}
	return s.GetConversationRecord(ctx, in.TenantID, id)
}

func (s *Store) GetConversationRecord(ctx context.Context, tenantID, id string) (ConversationRecord, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT r.id,r.tenant_id,r.call_id,r.customer_id,r.avatar_id,COALESCE(a.name,''),r.channel,r.status,r.started_at,r.ended_at,r.duration_seconds,r.summary,
(SELECT COUNT(*) FROM %s.conversation_archive_objects o WHERE o.conversation_record_id=r.id),
(SELECT COUNT(*) FROM %s.knowledge_gap_candidates g WHERE g.conversation_record_id=r.id)
FROM %s.conversation_records r LEFT JOIN %s.ai_avatars a ON a.id=r.avatar_id
WHERE r.tenant_id=$1 AND r.id=$2`, schema, schema, schema, schema), tenantID, id)
	return scanConversationRecord(row)
}

func (s *Store) GetConversationRecordByCallID(ctx context.Context, tenantID, callID string) (ConversationRecord, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT r.id,r.tenant_id,r.call_id,r.customer_id,r.avatar_id,COALESCE(a.name,''),r.channel,r.status,r.started_at,r.ended_at,r.duration_seconds,r.summary,
(SELECT COUNT(*) FROM %s.conversation_archive_objects o WHERE o.conversation_record_id=r.id),
(SELECT COUNT(*) FROM %s.knowledge_gap_candidates g WHERE g.conversation_record_id=r.id)
FROM %s.conversation_records r LEFT JOIN %s.ai_avatars a ON a.id=r.avatar_id
WHERE r.tenant_id=$1 AND r.call_id=$2
ORDER BY r.created_at DESC LIMIT 1`, schema, schema, schema, schema), tenantID, callID)
	return scanConversationRecord(row)
}

func (s *Store) ListConversationArchiveObjects(ctx context.Context, tenantID, conversationRecordID string) ([]ConversationArchiveObject, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id,tenant_id,conversation_record_id,object_key,object_type,content_type,size_bytes,checksum_sha256,protection_mode,status,error_code,stored_at
FROM %s.conversation_archive_objects
WHERE tenant_id=$1 AND conversation_record_id=$2
ORDER BY created_at ASC`, schema), tenantID, conversationRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ConversationArchiveObject
	for rows.Next() {
		var obj ConversationArchiveObject
		if err := rows.Scan(&obj.ID, &obj.TenantID, &obj.ConversationRecordID, &obj.ObjectKey, &obj.ObjectType, &obj.ContentType,
			&obj.SizeBytes, &obj.ChecksumSHA256, &obj.ProtectionMode, &obj.Status, &obj.ErrorCode, &obj.StoredAt); err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, rows.Err()
}

func (s *Store) GetConversationArchiveObjectContent(ctx context.Context, tenantID, conversationRecordID, objectID string) (*minio.Object, string, error) {
	if s == nil || s.pg == nil || s.minio == nil {
		return nil, "", fmt.Errorf("archive storage is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var objectKey, objectType, contentType, status string
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT object_key,object_type,content_type,status
FROM %s.conversation_archive_objects
WHERE tenant_id=$1 AND conversation_record_id=$2 AND id=$3`, schema),
		tenantID, conversationRecordID, objectID).Scan(&objectKey, &objectType, &contentType, &status)
	if err != nil {
		return nil, "", err
	}
	if status != "stored" || objectType != "audio" || strings.TrimSpace(objectKey) == "" {
		return nil, "", fmt.Errorf("archive object is not playable")
	}
	obj, err := s.minio.GetObject(ctx, s.cfg.MinioBucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	return obj, contentType, nil
}

func (s *Store) ListConversationTranscript(ctx context.Context, callID string) ([]ConversationTranscriptLine, error) {
	if strings.TrimSpace(callID) == "" {
		return []ConversationTranscriptLine{}, nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id::text, role, content, created_at
FROM %s.call_turns
WHERE call_id=$1
ORDER BY id ASC`, schema), callID)
	if err != nil {
		return nil, err
	}
	turns, err := scanTranscriptLines(rows)
	if err != nil {
		return nil, err
	}
	if len(turns) > 0 {
		return turns, nil
	}
	rows, err = s.pg.Query(ctx, fmt.Sprintf(`SELECT id::text, role, content, created_at
FROM %s.messages
WHERE call_id=$1
ORDER BY id ASC`, schema), callID)
	if err != nil {
		return nil, err
	}
	return scanTranscriptLines(rows)
}

func scanTranscriptLines(rows pgx.Rows) ([]ConversationTranscriptLine, error) {
	defer rows.Close()
	var out []ConversationTranscriptLine
	for rows.Next() {
		var line ConversationTranscriptLine
		if err := rows.Scan(&line.ID, &line.Role, &line.Content, &line.CreatedAt); err != nil {
			return nil, err
		}
		if line.Role == "agent" {
			line.Role = "assistant"
		}
		out = append(out, line)
	}
	return out, rows.Err()
}

func (s *Store) ListConversationRecords(ctx context.Context, tenantID, status, startDate, endDate string) ([]ConversationRecord, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := fmt.Sprintf(`SELECT r.id,r.tenant_id,r.call_id,r.customer_id,r.avatar_id,COALESCE(a.name,''),r.channel,r.status,r.started_at,r.ended_at,r.duration_seconds,r.summary,
(SELECT COUNT(*) FROM %s.conversation_archive_objects o WHERE o.conversation_record_id=r.id),
(SELECT COUNT(*) FROM %s.knowledge_gap_candidates g WHERE g.conversation_record_id=r.id)
FROM %s.conversation_records r LEFT JOIN %s.ai_avatars a ON a.id=r.avatar_id
WHERE r.tenant_id=$1`, schema, schema, schema, schema)
	args := []any{tenantID}
	if strings.TrimSpace(status) != "" {
		q += fmt.Sprintf(" AND r.status=$%d", len(args)+1)
		args = append(args, strings.TrimSpace(status))
	}
	if strings.TrimSpace(startDate) != "" {
		q += fmt.Sprintf(" AND r.started_at >= $%d::date", len(args)+1)
		args = append(args, strings.TrimSpace(startDate))
	}
	if strings.TrimSpace(endDate) != "" {
		q += fmt.Sprintf(" AND r.started_at < ($%d::date + INTERVAL '1 day')", len(args)+1)
		args = append(args, strings.TrimSpace(endDate))
	}
	q += " ORDER BY r.started_at DESC LIMIT 100"
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ConversationRecord
	for rows.Next() {
		rec, err := scanConversationRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func scanConversationRecord(row pgx.Row) (ConversationRecord, error) {
	var rec ConversationRecord
	var raw []byte
	err := row.Scan(&rec.ID, &rec.TenantID, &rec.CallID, &rec.CustomerID, &rec.AvatarID, &rec.AvatarName, &rec.Channel, &rec.Status,
		&rec.StartedAt, &rec.EndedAt, &rec.DurationSeconds, &raw, &rec.ArchiveObjectCount, &rec.KnowledgeGapCount)
	if err != nil {
		return ConversationRecord{}, err
	}
	rec.Summary = map[string]any{}
	_ = json.Unmarshal(raw, &rec.Summary)
	return rec, nil
}

func (s *Store) ArchiveConversationTranscript(ctx context.Context, tenantID, callID string, payload any, protectionMode string) (ConversationArchiveObject, error) {
	rec, err := s.UpsertConversationRecord(ctx, ConversationRecordInput{TenantID: tenantID, CallID: callID, Channel: "chat", Status: "archived", Summary: map[string]any{"archived_by": "server"}})
	if err != nil {
		return ConversationArchiveObject{}, err
	}
	raw, _ := json.MarshalIndent(payload, "", "  ")
	keyTenant := safeObjectPart(tenantID)
	keyCall := safeObjectPart(callID)
	if keyCall == "" {
		keyCall = strings.TrimPrefix(rec.ID, "crec_")
	}
	objectKey := "calls/" + keyTenant + "/" + keyCall + "/transcript.json"
	sum := sha256.Sum256(raw)
	obj := ConversationArchiveObject{
		ID: "cobj_" + newStoreID(), TenantID: tenantID, ConversationRecordID: rec.ID, ObjectKey: objectKey,
		ObjectType: "transcript", ContentType: "application/json", SizeBytes: int64(len(raw)),
		ChecksumSHA256: hex.EncodeToString(sum[:]), ProtectionMode: normalizeProtectionMode(protectionMode), Status: "stored", StoredAt: time.Now().UTC(),
	}
	if s.minio != nil {
		_, err = s.minio.PutObject(ctx, s.cfg.MinioBucket, objectKey, bytes.NewReader(raw), int64(len(raw)), minio.PutObjectOptions{ContentType: obj.ContentType})
		if err != nil {
			obj.Status = "failed"
			obj.ErrorCode = "archive_write_failed"
			_, _ = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.conversation_records SET status='archive_failed',updated_at=now() WHERE id=$1`, quoteIdent(s.cfg.PostgresSchema)), rec.ID)
		}
	} else {
		obj.Status = "failed"
		obj.ErrorCode = "minio_disabled"
	}
	if err := s.insertArchiveObject(ctx, obj); err != nil {
		return obj, err
	}
	return obj, err
}

func (s *Store) ArchiveConversationAudio(ctx context.Context, tenantID, callID, streamName, contentType string, data []byte, protectionMode string) (ConversationArchiveObject, error) {
	if s == nil || s.pg == nil {
		return ConversationArchiveObject{}, fmt.Errorf("postgres is not available")
	}
	streamName = safeObjectPart(streamName)
	if streamName == "" {
		streamName = "recording"
	}
	if contentType == "" {
		contentType = "audio/wav"
	}
	customerID, avatarID, startedAt, endedAt, durationSeconds := s.conversationContextForCall(ctx, callID)
	rec, err := s.UpsertConversationRecord(ctx, ConversationRecordInput{
		TenantID: tenantID, CallID: callID, CustomerID: customerID, AvatarID: avatarID,
		Channel: "voice", Status: "archived", StartedAt: startedAt, EndedAt: endedAt, DurationSeconds: durationSeconds,
		Summary: map[string]any{"archived_by": "server", "audio_stream": streamName},
	})
	if err != nil {
		return ConversationArchiveObject{}, err
	}
	keyTenant := safeObjectPart(tenantID)
	keyCall := safeObjectPart(callID)
	if keyCall == "" {
		keyCall = strings.TrimPrefix(rec.ID, "crec_")
	}
	objectKey := "calls/" + keyTenant + "/" + keyCall + "/audio/" + streamName + ".wav"
	sum := sha256.Sum256(data)
	obj := ConversationArchiveObject{
		ID: "cobj_" + newStoreID(), TenantID: tenantID, ConversationRecordID: rec.ID, ObjectKey: objectKey,
		ObjectType: "audio", ContentType: contentType, SizeBytes: int64(len(data)),
		ChecksumSHA256: hex.EncodeToString(sum[:]), ProtectionMode: normalizeProtectionMode(protectionMode), Status: "stored", StoredAt: time.Now().UTC(),
	}
	if len(data) == 0 {
		obj.Status = "failed"
		obj.ErrorCode = "empty_audio"
	} else if s.minio != nil {
		_, err = s.minio.PutObject(ctx, s.cfg.MinioBucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: obj.ContentType})
		if err != nil {
			obj.Status = "failed"
			obj.ErrorCode = "archive_write_failed"
			_, _ = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.conversation_records SET status='archive_failed',updated_at=now() WHERE id=$1`, quoteIdent(s.cfg.PostgresSchema)), rec.ID)
		}
	} else {
		obj.Status = "failed"
		obj.ErrorCode = "minio_disabled"
	}
	if err := s.insertArchiveObject(ctx, obj); err != nil {
		return obj, err
	}
	return obj, err
}

func (s *Store) conversationContextForCall(ctx context.Context, callID string) (string, string, time.Time, *time.Time, int) {
	if strings.TrimSpace(callID) == "" || s == nil || s.pg == nil {
		return "", "", time.Time{}, nil, 0
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var customerID, avatarID string
	var startedAt time.Time
	var endedAt *time.Time
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT customer_id, avatar_id, started_at, ended_at
FROM %s.call_sessions WHERE id=$1`, schema), callID).Scan(&customerID, &avatarID, &startedAt, &endedAt)
	if err != nil {
		return "", "", time.Time{}, nil, 0
	}
	duration := 0
	if endedAt != nil {
		duration = int(endedAt.Sub(startedAt).Seconds())
		if duration < 0 {
			duration = 0
		}
	}
	return customerID, avatarID, startedAt, endedAt, duration
}

func (s *Store) insertArchiveObject(ctx context.Context, obj ConversationArchiveObject) error {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.conversation_archive_objects
(id,tenant_id,conversation_record_id,object_key,object_type,content_type,size_bytes,checksum_sha256,protection_mode,status,error_code,stored_at,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$13)
ON CONFLICT (id) DO NOTHING`, schema),
		obj.ID, obj.TenantID, obj.ConversationRecordID, obj.ObjectKey, obj.ObjectType, obj.ContentType, obj.SizeBytes,
		obj.ChecksumSHA256, obj.ProtectionMode, obj.Status, obj.ErrorCode, nullableTime(obj.StoredAt), actor)
	return err
}

func (s *Store) CreateKnowledgeGapCandidate(ctx context.Context, in KnowledgeGapInput) (KnowledgeGapCandidate, error) {
	if in.GapReason == "" {
		in.GapReason = "no_source"
	}
	id := "kgap_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var out KnowledgeGapCandidate
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`INSERT INTO %s.knowledge_gap_candidates
(id,tenant_id,conversation_record_id,avatar_id,customer_id,source_turn_id,question,answer_excerpt,gap_reason,confidence,status,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'open',$11,$11)
RETURNING id,tenant_id,conversation_record_id,avatar_id,customer_id,source_turn_id,question,answer_excerpt,gap_reason,confidence,status,reviewer_note,snoozed_until,resolved_at,created_at,updated_at`, schema),
		id, in.TenantID, in.ConversationRecordID, in.AvatarID, in.CustomerID, in.SourceTurnID, trimBounded(in.Question, 1000), trimBounded(in.AnswerExcerpt, 1000), in.GapReason, in.Confidence, actor).
		Scan(&out.ID, &out.TenantID, &out.ConversationRecordID, &out.AvatarID, &out.CustomerID, &out.SourceTurnID, &out.Question, &out.AnswerExcerpt,
			&out.GapReason, &out.Confidence, &out.Status, &out.ReviewerNote, &out.SnoozedUntil, &out.ResolvedAt, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (s *Store) ListKnowledgeGapCandidates(ctx context.Context, tenantID, status string) ([]KnowledgeGapCandidate, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := fmt.Sprintf(`SELECT id,tenant_id,conversation_record_id,avatar_id,customer_id,source_turn_id,question,answer_excerpt,gap_reason,confidence,status,reviewer_note,snoozed_until,resolved_at,created_at,updated_at
FROM %s.knowledge_gap_candidates WHERE tenant_id=$1`, schema)
	args := []any{tenantID}
	if strings.TrimSpace(status) != "" {
		q += " AND status=$2"
		args = append(args, strings.TrimSpace(status))
	}
	q += " ORDER BY created_at DESC LIMIT 100"
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KnowledgeGapCandidate
	for rows.Next() {
		var g KnowledgeGapCandidate
		if err := rows.Scan(&g.ID, &g.TenantID, &g.ConversationRecordID, &g.AvatarID, &g.CustomerID, &g.SourceTurnID, &g.Question, &g.AnswerExcerpt,
			&g.GapReason, &g.Confidence, &g.Status, &g.ReviewerNote, &g.SnoozedUntil, &g.ResolvedAt, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) PatchKnowledgeGapCandidate(ctx context.Context, tenantID, id, status, note string) (KnowledgeGapCandidate, error) {
	status = strings.TrimSpace(status)
	if status != "open" && status != "snoozed" && status != "resolved" && status != "ignored" {
		return KnowledgeGapCandidate{}, fmt.Errorf("invalid gap status")
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var out KnowledgeGapCandidate
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`UPDATE %s.knowledge_gap_candidates
SET status=$3, reviewer_note=$4, resolved_at=CASE WHEN $3='resolved' THEN now() ELSE resolved_at END, updated_by=$5, updated_at=now()
WHERE tenant_id=$1 AND id=$2
RETURNING id,tenant_id,conversation_record_id,avatar_id,customer_id,source_turn_id,question,answer_excerpt,gap_reason,confidence,status,reviewer_note,snoozed_until,resolved_at,created_at,updated_at`, schema),
		tenantID, id, status, trimBounded(note, 1000), actor).
		Scan(&out.ID, &out.TenantID, &out.ConversationRecordID, &out.AvatarID, &out.CustomerID, &out.SourceTurnID, &out.Question, &out.AnswerExcerpt,
			&out.GapReason, &out.Confidence, &out.Status, &out.ReviewerNote, &out.SnoozedUntil, &out.ResolvedAt, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func normalizeProtectionMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "sse-s3", "sse-kms", "client":
		return value
	default:
		return "none"
	}
}

func safeObjectPart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	value = strings.ReplaceAll(value, "..", "_")
	return value
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func trimBounded(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if maxLen > 0 && len(value) > maxLen {
		return value[:maxLen]
	}
	return value
}
