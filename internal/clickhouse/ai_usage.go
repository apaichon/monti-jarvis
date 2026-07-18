package clickhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type AIUsageFact struct {
	FactID           string `json:"fact_id"`
	TenantID         string `json:"tenant_id"`
	CallID           string `json:"call_id"`
	ConversationID   string `json:"conversation_record_id"`
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	Modality         string `json:"modality"`
	MeasurementState string `json:"measurement_state"`
	InputUnits       uint64 `json:"input_units"`
	OutputUnits      uint64 `json:"output_units"`
	AudioSeconds     uint32 `json:"audio_seconds"`
	RateVersion      string `json:"rate_version"`
	CostMicrounits   int64  `json:"cost_microunits"`
	Currency         string `json:"currency"`
	UsageDate        string `json:"usage_date"`
	SourceUpdatedAt  string `json:"source_updated_at"`
	UpdatedAt        string `json:"updated_at"`
}

type AIUsageStats struct {
	Events              int
	ObservedEvents      int
	EstimatedEvents     int
	UnavailableEvents   int
	ObservedCostMicros  int64
	EstimatedCostMicros int64
	InputUnits          uint64
	OutputUnits         uint64
	AudioSeconds        uint64
	LastProjectedAt     time.Time
	ByTenant            []AIUsageBucket
}

type AIUsageBucket struct {
	TenantID            string
	Events              int
	ObservedEvents      int
	EstimatedEvents     int
	UnavailableEvents   int
	ObservedCostMicros  int64
	EstimatedCostMicros int64
	InputUnits          uint64
	OutputUnits         uint64
	AudioSeconds        uint64
	LastProjectedAt     time.Time
}

type aiUsageQueryRow struct {
	TenantID            string          `json:"tenant_id"`
	Events              clickhouseInt   `json:"events"`
	ObservedEvents      clickhouseInt   `json:"observed_events"`
	EstimatedEvents     clickhouseInt   `json:"estimated_events"`
	UnavailableEvents   clickhouseInt   `json:"unavailable_events"`
	ObservedCostMicros  clickhouseInt64 `json:"observed_cost_microunits"`
	EstimatedCostMicros clickhouseInt64 `json:"estimated_cost_microunits"`
	InputUnits          clickhouseInt64 `json:"input_units"`
	OutputUnits         clickhouseInt64 `json:"output_units"`
	AudioSeconds        clickhouseInt64 `json:"audio_seconds"`
	Freshness           string          `json:"freshness"`
}

type clickhouseInt64 int64

func (value *clickhouseInt64) UnmarshalJSON(data []byte) error {
	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		*value = clickhouseInt64(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	var parsed int64
	if _, err := fmt.Sscan(strings.TrimSpace(text), &parsed); err != nil {
		return err
	}
	*value = clickhouseInt64(parsed)
	return nil
}

func (c *Client) EnsureAIUsageSchema(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.ai_cost_usage_facts (
  fact_id String,
  tenant_id String,
  call_id String,
  conversation_record_id String,
  provider LowCardinality(String),
  model LowCardinality(String),
  modality LowCardinality(String),
  measurement_state LowCardinality(String),
  input_units UInt64,
  output_units UInt64,
  audio_seconds UInt32,
  rate_version String,
  cost_microunits Int64,
  currency FixedString(3),
  usage_date Date,
  source_updated_at DateTime,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (tenant_id, usage_date, provider, model, fact_id)`, quoteIdent(c.db))
	return c.exec(ctx, stmt)
}

func (c *Client) UpsertAIUsageFact(ctx context.Context, fact AIUsageFact) error {
	if !c.Enabled() {
		return fmt.Errorf("clickhouse is not configured")
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(fact); err != nil {
		return fmt.Errorf("encode ai usage fact: %w", err)
	}
	query := fmt.Sprintf("INSERT INTO %s.ai_cost_usage_facts (fact_id, tenant_id, call_id, conversation_record_id, provider, model, modality, measurement_state, input_units, output_units, audio_seconds, rate_version, cost_microunits, currency, usage_date, source_updated_at, updated_at) FORMAT JSONEachRow", quoteIdent(c.db))
	return c.execInsert(ctx, query, body.Bytes())
}

func (c *Client) QueryAIUsageStats(ctx context.Context, tenantID, startDate, endDate string) (AIUsageStats, error) {
	if !c.Enabled() {
		return AIUsageStats{}, fmt.Errorf("clickhouse is not configured")
	}
	filter := ""
	if strings.TrimSpace(tenantID) != "" {
		filter = "\n  AND tenant_id = '" + escape(tenantID) + "'"
	}
	query := fmt.Sprintf(`
SELECT tenant_id,
       count() AS events,
       countIf(measurement_state = 'observed') AS observed_events,
       countIf(measurement_state = 'estimated') AS estimated_events,
       countIf(measurement_state = 'unavailable') AS unavailable_events,
       sumIf(cost_microunits, measurement_state = 'observed') AS observed_cost_microunits,
       sumIf(cost_microunits, measurement_state = 'estimated') AS estimated_cost_microunits,
       sum(input_units) AS input_units,
       sum(output_units) AS output_units,
       sum(audio_seconds) AS audio_seconds,
       formatDateTime(max(updated_at), '%%Y-%%m-%%d %%H:%%i:%%s', 'UTC') AS freshness
FROM %s.ai_cost_usage_facts FINAL
WHERE usage_date BETWEEN toDate('%s') AND toDate('%s')%s
GROUP BY tenant_id
ORDER BY events DESC, tenant_id
FORMAT JSON`, quoteIdent(c.db), escape(startDate), escape(endDate), filter)
	body, err := c.query(ctx, query)
	if err != nil {
		return AIUsageStats{}, err
	}
	var parsed struct {
		Data []aiUsageQueryRow `json:"data"`
	}
	if err := jsonUnmarshal(body, &parsed); err != nil {
		return AIUsageStats{}, err
	}
	stats := AIUsageStats{}
	for _, row := range parsed.Data {
		bucket := AIUsageBucket{
			TenantID:            row.TenantID,
			Events:              int(row.Events),
			ObservedEvents:      int(row.ObservedEvents),
			EstimatedEvents:     int(row.EstimatedEvents),
			UnavailableEvents:   int(row.UnavailableEvents),
			ObservedCostMicros:  int64(row.ObservedCostMicros),
			EstimatedCostMicros: int64(row.EstimatedCostMicros),
			InputUnits:          uint64(maxInt64(int64(row.InputUnits))),
			OutputUnits:         uint64(maxInt64(int64(row.OutputUnits))),
			AudioSeconds:        uint64(maxInt64(int64(row.AudioSeconds))),
		}
		if parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", row.Freshness); parseErr == nil {
			bucket.LastProjectedAt = parsedTime
			if parsedTime.After(stats.LastProjectedAt) {
				stats.LastProjectedAt = parsedTime
			}
		}
		stats.Events += bucket.Events
		stats.ObservedEvents += bucket.ObservedEvents
		stats.EstimatedEvents += bucket.EstimatedEvents
		stats.UnavailableEvents += bucket.UnavailableEvents
		stats.ObservedCostMicros += bucket.ObservedCostMicros
		stats.EstimatedCostMicros += bucket.EstimatedCostMicros
		stats.InputUnits += bucket.InputUnits
		stats.OutputUnits += bucket.OutputUnits
		stats.AudioSeconds += bucket.AudioSeconds
		stats.ByTenant = append(stats.ByTenant, bucket)
	}
	sort.Slice(stats.ByTenant, func(i, j int) bool {
		if stats.ByTenant[i].Events != stats.ByTenant[j].Events {
			return stats.ByTenant[i].Events > stats.ByTenant[j].Events
		}
		return stats.ByTenant[i].TenantID < stats.ByTenant[j].TenantID
	})
	return stats, nil
}

func maxInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
