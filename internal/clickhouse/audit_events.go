package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/audit"
)

type AuditEventFilter struct {
	TenantID     string
	ActorID      string
	Action       string
	ResourceType string
	Outcome      string
	StartDate    string
	EndDate      string
	Limit        int
	Offset       int
}

type AuditEventPage struct {
	Events  []audit.Event
	HasMore bool
}

func (c *Client) EnsureAuditEventsSchema(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	return c.exec(ctx, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.audit_events (
  event_id String,
  occurred_at DateTime64(3, 'UTC'),
  tenant_id String,
  actor_id String,
  actor_type LowCardinality(String),
  action LowCardinality(String),
  resource_type LowCardinality(String),
  resource_id String,
  request_id String,
  source LowCardinality(String),
  outcome LowCardinality(String),
  metadata_json String,
  ingested_at DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = ReplacingMergeTree(ingested_at)
ORDER BY (tenant_id, occurred_at, event_id)`, quoteIdent(c.db)))
}

func (c *Client) InsertAuditEvents(ctx context.Context, events []audit.Event) error {
	if !c.Enabled() {
		return fmt.Errorf("clickhouse is not configured")
	}
	if len(events) == 0 {
		return nil
	}
	type payload struct {
		EventID      string `json:"event_id"`
		OccurredAt   string `json:"occurred_at"`
		TenantID     string `json:"tenant_id"`
		ActorID      string `json:"actor_id"`
		ActorType    string `json:"actor_type"`
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
		RequestID    string `json:"request_id"`
		Source       string `json:"source"`
		Outcome      string `json:"outcome"`
		MetadataJSON string `json:"metadata_json"`
	}
	var body strings.Builder
	encoder := json.NewEncoder(&body)
	for _, event := range events {
		metadata, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("encode audit metadata: %w", err)
		}
		if err := encoder.Encode(payload{
			EventID:      event.EventID,
			OccurredAt:   event.OccurredAt.UTC().Format("2006-01-02 15:04:05.000"),
			TenantID:     event.TenantID,
			ActorID:      event.ActorID,
			ActorType:    event.ActorType,
			Action:       event.Action,
			ResourceType: event.ResourceType,
			ResourceID:   event.ResourceID,
			RequestID:    event.RequestID,
			Source:       event.Source,
			Outcome:      event.Outcome,
			MetadataJSON: string(metadata),
		}); err != nil {
			return fmt.Errorf("encode audit event: %w", err)
		}
	}
	query := fmt.Sprintf(`INSERT INTO %s.audit_events
  (event_id, occurred_at, tenant_id, actor_id, actor_type, action, resource_type, resource_id, request_id, source, outcome, metadata_json)
FORMAT JSONEachRow`, quoteIdent(c.db))
	return c.execInsert(ctx, query, []byte(body.String()))
}

func (c *Client) ListAuditEvents(ctx context.Context, filter AuditEventFilter) (AuditEventPage, error) {
	if !c.Enabled() {
		return AuditEventPage{}, fmt.Errorf("clickhouse is not configured")
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	where := []string{"1 = 1"}
	addStringFilter := func(column, value string) {
		if strings.TrimSpace(value) != "" {
			where = append(where, column+" = '"+escape(value)+"'")
		}
	}
	addStringFilter("tenant_id", filter.TenantID)
	addStringFilter("actor_id", filter.ActorID)
	addStringFilter("action", filter.Action)
	addStringFilter("resource_type", filter.ResourceType)
	addStringFilter("outcome", filter.Outcome)
	if filter.StartDate != "" {
		where = append(where, "occurred_at >= toDateTime64('"+escape(filter.StartDate)+" 00:00:00', 3, 'UTC')")
	}
	if filter.EndDate != "" {
		end, err := time.Parse("2006-01-02", filter.EndDate)
		if err != nil {
			return AuditEventPage{}, fmt.Errorf("invalid end date")
		}
		where = append(where, "occurred_at < toDateTime64('"+end.AddDate(0, 0, 1).Format("2006-01-02")+" 00:00:00', 3, 'UTC')")
	}
	query := fmt.Sprintf(`SELECT event_id, occurred_at, tenant_id, actor_id, actor_type, action, resource_type, resource_id, request_id, source, outcome, metadata_json
FROM %s.audit_events FINAL
WHERE %s
ORDER BY occurred_at DESC, event_id DESC
LIMIT %d OFFSET %d FORMAT JSON`, quoteIdent(c.db), strings.Join(where, " AND "), filter.Limit+1, filter.Offset)
	body, err := c.query(ctx, query)
	if err != nil {
		return AuditEventPage{}, err
	}
	var parsed struct {
		Data []struct {
			EventID      string `json:"event_id"`
			OccurredAt   string `json:"occurred_at"`
			TenantID     string `json:"tenant_id"`
			ActorID      string `json:"actor_id"`
			ActorType    string `json:"actor_type"`
			Action       string `json:"action"`
			ResourceType string `json:"resource_type"`
			ResourceID   string `json:"resource_id"`
			RequestID    string `json:"request_id"`
			Source       string `json:"source"`
			Outcome      string `json:"outcome"`
			MetadataJSON string `json:"metadata_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return AuditEventPage{}, fmt.Errorf("decode audit events: %w", err)
	}
	page := AuditEventPage{Events: make([]audit.Event, 0, len(parsed.Data))}
	if len(parsed.Data) > filter.Limit {
		page.HasMore = true
		parsed.Data = parsed.Data[:filter.Limit]
	}
	for _, row := range parsed.Data {
		occurredAt, err := parseClickHouseTime(row.OccurredAt)
		if err != nil {
			return AuditEventPage{}, err
		}
		var metadata map[string]any
		if strings.TrimSpace(row.MetadataJSON) != "" {
			if err := json.Unmarshal([]byte(row.MetadataJSON), &metadata); err != nil {
				metadata = map[string]any{}
			}
		}
		page.Events = append(page.Events, audit.Event{
			EventID: row.EventID, OccurredAt: occurredAt, TenantID: row.TenantID, ActorID: row.ActorID,
			ActorType: row.ActorType, Action: row.Action, ResourceType: row.ResourceType, ResourceID: row.ResourceID,
			RequestID: row.RequestID, Source: row.Source, Outcome: row.Outcome, Metadata: metadata,
		})
	}
	return page, nil
}

func parseClickHouseTime(value string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", time.RFC3339Nano} {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid audit event timestamp")
}
