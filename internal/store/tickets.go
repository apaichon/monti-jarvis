package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrTicketConflict    = errors.New("ticket already exists")
	ErrTicketIdempotency = errors.New("idempotency key conflicts with an existing request")
	ErrInvalidTicket     = errors.New("invalid ticket")
	ErrInvalidTransition = errors.New("invalid ticket status transition")
)

type Ticket struct {
	ID                   string         `json:"id"`
	TenantID             string         `json:"tenant_id,omitempty"`
	ConversationRecordID string         `json:"conversation_record_id,omitempty"`
	CallID               string         `json:"call_id,omitempty"`
	CustomerID           string         `json:"customer_id,omitempty"`
	AvatarID             string         `json:"avatar_id,omitempty"`
	AvatarName           string         `json:"avatar_name,omitempty"`
	Subject              string         `json:"subject"`
	Description          string         `json:"description,omitempty"`
	SourceSummary        map[string]any `json:"source_summary,omitempty"`
	Category             string         `json:"category"`
	Priority             string         `json:"priority"`
	Status               string         `json:"status"`
	Source               string         `json:"source"`
	AssigneeUserID       string         `json:"assignee_user_id,omitempty"`
	ContactName          string         `json:"contact_name,omitempty"`
	ContactEmail         string         `json:"contact_email,omitempty"`
	ResolvedAt           *time.Time     `json:"resolved_at,omitempty"`
	ClosedAt             *time.Time     `json:"closed_at,omitempty"`
	LastActivityAt       time.Time      `json:"last_activity_at"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type TicketEvent struct {
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id,omitempty"`
	TicketID  string         `json:"ticket_id"`
	EventType string         `json:"event_type"`
	ActorType string         `json:"actor_type"`
	ActorID   string         `json:"actor_id,omitempty"`
	Note      string         `json:"note,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type TicketInput struct {
	TenantID             string
	ConversationRecordID string
	CallID               string
	CustomerID           string
	AvatarID             string
	Subject              string
	Description          string
	SourceSummary        map[string]any
	Category             string
	Priority             string
	Source               string
	AssigneeUserID       string
	ContactName          string
	ContactEmail         string
	IdempotencyKey       string
	ActorType            string
	ActorID              string
}

type TicketFilters struct {
	StartDate      string
	EndDate        string
	Status         string
	Priority       string
	Category       string
	AvatarID       string
	CustomerID     string
	AssigneeUserID string
}

func (s *Store) ensureTicketsSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tickets (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  conversation_record_id text REFERENCES %s.conversation_records(id) ON DELETE SET NULL,
  call_id text NOT NULL DEFAULT '',
  customer_id text REFERENCES %s.customers(id) ON DELETE SET NULL,
  avatar_id text REFERENCES %s.ai_avatars(id) ON DELETE SET NULL,
  subject text NOT NULL,
  description text NOT NULL DEFAULT '',
  source_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  category text NOT NULL DEFAULT 'general' CHECK (category IN ('general','billing','technical','other')),
  priority text NOT NULL DEFAULT 'normal' CHECK (priority IN ('low','normal','high','urgent')),
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','in_progress','waiting_customer','resolved','closed')),
  source text NOT NULL DEFAULT 'customer_request' CHECK (source IN ('customer_request','agent_escalation','tenant_created')),
  assignee_user_id text,
  contact_name text NOT NULL DEFAULT '',
  contact_email text NOT NULL DEFAULT '',
  idempotency_key text NOT NULL DEFAULT '',
  resolved_at timestamptz,
  closed_at timestamptz,
  last_activity_at timestamptz NOT NULL DEFAULT now(),%s
)`, schema, schema, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS tickets_tenant_idempotency_idx
ON %s.tickets (tenant_id, idempotency_key) WHERE idempotency_key <> ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tickets ADD COLUMN IF NOT EXISTS source_summary jsonb NOT NULL DEFAULT '{}'::jsonb`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS tickets_tenant_queue_idx
ON %s.tickets (tenant_id, status, priority, last_activity_at DESC)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS tickets_tenant_started_idx
ON %s.tickets (tenant_id, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.ticket_events (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  ticket_id text NOT NULL REFERENCES %s.tickets(id) ON DELETE CASCADE,
  event_type text NOT NULL CHECK (event_type IN ('created','status_changed','priority_changed','assigned','note_added','customer_confirmed')),
  actor_type text NOT NULL CHECK (actor_type IN ('system','customer','tenant_user')),
  actor_id text NOT NULL DEFAULT '',
  note text NOT NULL DEFAULT '',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS ticket_events_ticket_idx
ON %s.ticket_events (tenant_id, ticket_id, created_at ASC)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("tickets schema: %w", err)
		}
	}
	return nil
}

func (s *Store) CreateTicket(ctx context.Context, in TicketInput) (Ticket, bool, error) {
	if s == nil || s.pg == nil {
		return Ticket{}, false, fmt.Errorf("postgres is not available")
	}
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.Subject = trimBounded(strings.TrimSpace(in.Subject), 160)
	in.Description = trimBounded(strings.TrimSpace(in.Description), 2000)
	if in.SourceSummary == nil {
		in.SourceSummary = map[string]any{}
	}
	in.Category = strings.TrimSpace(in.Category)
	in.Priority = strings.TrimSpace(in.Priority)
	in.Source = strings.TrimSpace(in.Source)
	in.IdempotencyKey = trimBounded(strings.TrimSpace(in.IdempotencyKey), 160)
	if in.Category == "" {
		in.Category = "general"
	}
	if in.Priority == "" {
		in.Priority = "normal"
	}
	if in.Source == "" {
		in.Source = "customer_request"
	}
	if in.ActorType == "" {
		in.ActorType = "system"
	}
	if err := validateTicketFields(in.Category, in.Priority, in.Source, in.ActorType); err != nil {
		return Ticket{}, false, err
	}
	if in.TenantID == "" || in.Subject == "" || in.Description == "" {
		return Ticket{}, false, fmt.Errorf("tenant, subject, and description are required")
	}

	schema := quoteIdent(s.cfg.PostgresSchema)
	if in.IdempotencyKey != "" {
		existing, err := s.getTicketByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey)
		if err == nil {
			if existing.Subject != in.Subject || existing.Description != in.Description {
				return Ticket{}, false, ErrTicketIdempotency
			}
			return existing, true, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Ticket{}, false, err
		}
	}
	if in.CallID != "" {
		var openID string
		err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s.tickets
WHERE tenant_id=$1 AND call_id=$2 AND status IN ('open','in_progress','waiting_customer')
ORDER BY created_at DESC LIMIT 1`, schema), in.TenantID, strings.TrimSpace(in.CallID)).Scan(&openID)
		if err == nil {
			return Ticket{}, false, ErrTicketConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Ticket{}, false, err
		}
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return Ticket{}, false, err
	}
	defer tx.Rollback(ctx)

	id := "tick_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	if in.ActorID != "" {
		actor = in.ActorID
	}
	var recID any = nilIfBlank(in.ConversationRecordID)
	var customerID any = nilIfBlank(in.CustomerID)
	var avatarID any = nilIfBlank(in.AvatarID)
	var assigneeID any = nilIfBlank(in.AssigneeUserID)
	sourceSummary, _ := json.Marshal(in.SourceSummary)
	_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tickets
(id,tenant_id,conversation_record_id,call_id,customer_id,avatar_id,subject,description,source_summary,category,priority,status,source,assignee_user_id,contact_name,contact_email,idempotency_key,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'open',$12,$13,$14,$15,$16,$17,$17)`, schema),
		id, in.TenantID, recID, strings.TrimSpace(in.CallID), customerID, avatarID, in.Subject, in.Description,
		sourceSummary, in.Category, in.Priority, in.Source, assigneeID, trimBounded(in.ContactName, 160), trimBounded(in.ContactEmail, 320), in.IdempotencyKey, actor)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tickets_tenant_idempotency_idx") {
			existing, lookupErr := s.getTicketByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey)
			if lookupErr == nil {
				return existing, true, nil
			}
		}
		return Ticket{}, false, err
	}
	payload, _ := json.Marshal(map[string]any{"source": in.Source, "conversation_record_id": in.ConversationRecordID, "call_id": in.CallID})
	_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.ticket_events
(id,tenant_id,ticket_id,event_type,actor_type,actor_id,payload,created_by,updated_by)
VALUES ($1,$2,$3,'created',$4,$5,$6,$7,$7)`, schema),
		"tev_"+newStoreID(), in.TenantID, id, in.ActorType, in.ActorID, payload, actor)
	if err != nil {
		return Ticket{}, false, err
	}
	if in.ActorType == "customer" {
		_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.ticket_events
(id,tenant_id,ticket_id,event_type,actor_type,actor_id,payload,created_by,updated_by)
VALUES ($1,$2,$3,'customer_confirmed','customer',$4,'{}'::jsonb,$5,$5)`, schema),
			"tev_"+newStoreID(), in.TenantID, id, in.ActorID, actor)
		if err != nil {
			return Ticket{}, false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Ticket{}, false, err
	}
	created, err := s.GetTicket(ctx, in.TenantID, id)
	return created, false, err
}

func (s *Store) getTicketByIdempotencyKey(ctx context.Context, tenantID, key string) (Ticket, error) {
	return s.getTicket(ctx, `idempotency_key=$2`, tenantID, key)
}

func (s *Store) GetTicket(ctx context.Context, tenantID, id string) (Ticket, error) {
	return s.getTicket(ctx, `id=$2`, tenantID, id)
}

func (s *Store) getTicket(ctx context.Context, predicate string, tenantID, value string) (Ticket, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT t.id,t.tenant_id,COALESCE(t.conversation_record_id,''),t.call_id,COALESCE(t.customer_id,''),COALESCE(t.avatar_id,''),COALESCE(a.name,''),t.subject,t.description,COALESCE(t.source_summary,'{}'::jsonb),t.category,t.priority,t.status,t.source,COALESCE(t.assignee_user_id,''),t.contact_name,t.contact_email,t.resolved_at,t.closed_at,t.last_activity_at,t.created_at,t.updated_at
FROM %s.tickets t LEFT JOIN %s.ai_avatars a ON a.id=t.avatar_id
WHERE t.tenant_id=$1 AND t.%s`, schema, schema, predicate), tenantID, value)
	return scanTicket(row)
}

func (s *Store) ListTickets(ctx context.Context, tenantID string, f TicketFilters) ([]Ticket, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := fmt.Sprintf(`SELECT t.id,t.tenant_id,COALESCE(t.conversation_record_id,''),t.call_id,COALESCE(t.customer_id,''),COALESCE(t.avatar_id,''),COALESCE(a.name,''),t.subject,t.description,COALESCE(t.source_summary,'{}'::jsonb),t.category,t.priority,t.status,t.source,COALESCE(t.assignee_user_id,''),t.contact_name,t.contact_email,t.resolved_at,t.closed_at,t.last_activity_at,t.created_at,t.updated_at
FROM %s.tickets t LEFT JOIN %s.ai_avatars a ON a.id=t.avatar_id WHERE t.tenant_id=$1`, schema, schema)
	args := []any{tenantID}
	add := func(condition string, value any) {
		args = append(args, value)
		q += fmt.Sprintf(" AND %s=$%d", condition, len(args))
	}
	if f.StartDate != "" {
		args = append(args, f.StartDate)
		q += fmt.Sprintf(" AND t.created_at >= $%d::date", len(args))
	}
	if f.EndDate != "" {
		args = append(args, f.EndDate)
		q += fmt.Sprintf(" AND t.created_at < ($%d::date + INTERVAL '1 day')", len(args))
	}
	if f.Status != "" {
		add("t.status", f.Status)
	}
	if f.Priority != "" {
		add("t.priority", f.Priority)
	}
	if f.Category != "" {
		add("t.category", f.Category)
	}
	if f.AvatarID != "" {
		add("t.avatar_id", f.AvatarID)
	}
	if f.CustomerID != "" {
		add("t.customer_id", f.CustomerID)
	}
	if f.AssigneeUserID != "" {
		add("t.assignee_user_id", f.AssigneeUserID)
	}
	q += " ORDER BY t.last_activity_at DESC LIMIT 100"
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) UpdateTicket(ctx context.Context, tenantID, id string, status, priority, assignee *string) (Ticket, TicketEvent, error) {
	current, err := s.GetTicket(ctx, tenantID, id)
	if err != nil {
		return Ticket{}, TicketEvent{}, ErrTicketNotFound
	}
	nextStatus := current.Status
	if status != nil {
		nextStatus = strings.TrimSpace(*status)
	}
	nextPriority := current.Priority
	if priority != nil {
		nextPriority = strings.TrimSpace(*priority)
	}
	nextAssignee := current.AssigneeUserID
	if assignee != nil {
		nextAssignee = strings.TrimSpace(*assignee)
	}
	if err := validateTicketFields(current.Category, nextPriority, current.Source, "tenant_user"); err != nil {
		return Ticket{}, TicketEvent{}, err
	}
	if !validTicketTransition(current.Status, nextStatus) {
		return Ticket{}, TicketEvent{}, fmt.Errorf("%w: %s to %s", ErrInvalidTransition, current.Status, nextStatus)
	}
	if status == nil && priority == nil && assignee == nil {
		return current, TicketEvent{}, nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	if nextAssignee != "" {
		var assigned bool
		if err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (
SELECT 1 FROM %s.user_roles r JOIN %s.users u ON u.id=r.user_id
WHERE r.user_id=$1 AND r.tenant_id=$2 AND r.role='tenant_admin' AND u.status='active'
)`, schema, schema), nextAssignee, tenantID).Scan(&assigned); err != nil {
			return Ticket{}, TicketEvent{}, err
		}
		if !assigned {
			return Ticket{}, TicketEvent{}, fmt.Errorf("assignee must be an active tenant admin: %w", ErrInvalidTicket)
		}
	}
	actor := auditctx.ActorID(ctx)
	var resolvedAt any = current.ResolvedAt
	var closedAt any = current.ClosedAt
	if nextStatus == "resolved" && current.Status != "resolved" {
		resolvedAt = time.Now().UTC()
	}
	if nextStatus == "closed" && current.Status != "closed" {
		closedAt = time.Now().UTC()
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.tickets
SET status=$3,priority=$4,assignee_user_id=$5,resolved_at=$6,closed_at=$7,last_activity_at=now(),updated_by=$8,updated_at=now()
WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id, nextStatus, nextPriority, nilIfBlank(nextAssignee), resolvedAt, closedAt, actor)
	if err != nil {
		return Ticket{}, TicketEvent{}, err
	}
	eventType := "assigned"
	if current.Status != nextStatus {
		eventType = "status_changed"
	} else if current.Priority != nextPriority {
		eventType = "priority_changed"
	}
	payload := map[string]any{"status": nextStatus, "priority": nextPriority, "assignee_user_id": nextAssignee}
	event, err := s.addTicketEvent(ctx, tenantID, id, eventType, "tenant_user", actor, "", payload)
	if err != nil {
		return Ticket{}, TicketEvent{}, err
	}
	updated, err := s.GetTicket(ctx, tenantID, id)
	return updated, event, err
}

func (s *Store) AddTicketNote(ctx context.Context, tenantID, id, note string) (Ticket, TicketEvent, error) {
	note = trimBounded(strings.TrimSpace(note), 2000)
	if note == "" {
		return Ticket{}, TicketEvent{}, fmt.Errorf("note is required")
	}
	if _, err := s.GetTicket(ctx, tenantID, id); err != nil {
		return Ticket{}, TicketEvent{}, ErrTicketNotFound
	}
	event, err := s.addTicketEvent(ctx, tenantID, id, "note_added", "tenant_user", auditctx.ActorID(ctx), note, nil)
	if err != nil {
		return Ticket{}, TicketEvent{}, err
	}
	updated, err := s.GetTicket(ctx, tenantID, id)
	return updated, event, err
}

func (s *Store) addTicketEvent(ctx context.Context, tenantID, ticketID, eventType, actorType, actorID, note string, payload map[string]any) (TicketEvent, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	if payload == nil {
		payload = map[string]any{}
	}
	raw, _ := json.Marshal(payload)
	actor := auditctx.ActorID(ctx)
	if actorID != "" {
		actor = actorID
	}
	event := TicketEvent{ID: "tev_" + newStoreID(), TenantID: tenantID, TicketID: ticketID, EventType: eventType, ActorType: actorType, ActorID: actorID, Note: note, Payload: payload, CreatedAt: time.Now().UTC()}
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.ticket_events
(id,tenant_id,ticket_id,event_type,actor_type,actor_id,note,payload,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)`, schema), event.ID, tenantID, ticketID, eventType, actorType, actorID, note, raw, actor)
	if err != nil {
		return TicketEvent{}, err
	}
	_, _ = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.tickets SET last_activity_at=now(),updated_at=now(),updated_by=$3 WHERE tenant_id=$1 AND id=$2`, schema), tenantID, ticketID, actor)
	return event, nil
}

func (s *Store) ListTicketEvents(ctx context.Context, tenantID, ticketID string) ([]TicketEvent, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id,tenant_id,ticket_id,event_type,actor_type,actor_id,note,payload,created_at
FROM %s.ticket_events WHERE tenant_id=$1 AND ticket_id=$2 ORDER BY created_at ASC`, schema), tenantID, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TicketEvent
	for rows.Next() {
		var event TicketEvent
		var raw []byte
		if err := rows.Scan(&event.ID, &event.TenantID, &event.TicketID, &event.EventType, &event.ActorType, &event.ActorID, &event.Note, &raw, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.Payload = map[string]any{}
		_ = json.Unmarshal(raw, &event.Payload)
		out = append(out, event)
	}
	return out, rows.Err()
}

func (s *Store) ValidateCallReference(ctx context.Context, tenantID, callID string) error {
	if strings.TrimSpace(callID) == "" || s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var exists bool
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (
  SELECT 1 FROM %s.call_sessions WHERE tenant_id=$1 AND id=$2
  UNION ALL
  SELECT 1 FROM %s.conversation_records WHERE tenant_id=$1 AND call_id=$2
)`, schema, schema), tenantID, callID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTicketNotFound
	}
	return nil
}

func scanTicket(row pgx.Row) (Ticket, error) {
	var t Ticket
	var rawSummary []byte
	err := row.Scan(&t.ID, &t.TenantID, &t.ConversationRecordID, &t.CallID, &t.CustomerID, &t.AvatarID, &t.AvatarName,
		&t.Subject, &t.Description, &rawSummary, &t.Category, &t.Priority, &t.Status, &t.Source, &t.AssigneeUserID, &t.ContactName,
		&t.ContactEmail, &t.ResolvedAt, &t.ClosedAt, &t.LastActivityAt, &t.CreatedAt, &t.UpdatedAt)
	t.SourceSummary = map[string]any{}
	_ = json.Unmarshal(rawSummary, &t.SourceSummary)
	return t, err
}

func validateTicketFields(category, priority, source, actorType string) error {
	if !oneOf(category, "general", "billing", "technical", "other") ||
		!oneOf(priority, "low", "normal", "high", "urgent") ||
		!oneOf(source, "customer_request", "agent_escalation", "tenant_created") ||
		!oneOf(actorType, "system", "customer", "tenant_user") {
		return ErrInvalidTicket
	}
	return nil
}

func validTicketTransition(current, next string) bool {
	if current == next {
		return true
	}
	switch current {
	case "open":
		return oneOf(next, "in_progress", "closed")
	case "in_progress":
		return oneOf(next, "waiting_customer", "resolved", "closed")
	case "waiting_customer":
		return oneOf(next, "in_progress", "closed")
	case "resolved":
		return oneOf(next, "closed", "in_progress")
	default:
		return false
	}
}

func oneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

func nilIfBlank(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
