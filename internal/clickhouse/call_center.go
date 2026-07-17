package clickhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CallCenterUsageFact is the privacy-minimized analytics projection of one
// completed conversation. The fact key is stable across archive retries.
type CallCenterUsageFact struct {
	FactID          string `json:"fact_id"`
	TenantID        string `json:"tenant_id"`
	CallID          string `json:"call_id"`
	ConversationID  string `json:"conversation_record_id"`
	AvatarID        string `json:"avatar_id"`
	Channel         string `json:"channel"`
	Source          string `json:"source"`
	Status          string `json:"status"`
	StartedAt       string `json:"started_at"`
	EndedAt         string `json:"ended_at"`
	UsageDate       string `json:"usage_date"`
	DurationSeconds int    `json:"duration_seconds"`
	SourceUpdatedAt string `json:"source_updated_at"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	CreatedBy       string `json:"created_by"`
	UpdatedBy       string `json:"updated_by"`
}

type CallCenterStats struct {
	CompletedConversations int
	TotalDurationSeconds   int
	AverageDurationSeconds float64
	ByAvatar               []CallCenterBucket
	ByChannel              []CallCenterBucket
	Freshness              time.Time
}

type CallCenterBucket struct {
	ID                     string
	Channel                string
	CompletedConversations int
	TotalDurationSeconds   int
	AverageDurationSeconds float64
}

// PlatformCallCenterStats is the aggregate-only projection used by the
// platform dashboard. It intentionally contains no conversation or customer
// identifiers beyond the tenant and avatar dimensions needed for grouping.
type PlatformCallCenterStats struct {
	CompletedConversations int
	TotalDurationSeconds   int
	ByTenant               []PlatformCallCenterBucket
	ByAvatar               []PlatformCallCenterBucket
	ByChannel              []PlatformCallCenterBucket
	LastProjectedAt        time.Time
}

type PlatformCallCenterBucket struct {
	TenantID               string
	AvatarID               string
	Channel                string
	CompletedConversations int
	TotalDurationSeconds   int
	AverageDurationSeconds float64
}

type callCenterQueryRow struct {
	AvatarID     string        `json:"avatar_id"`
	Channel      string        `json:"channel"`
	Sessions     clickhouseInt `json:"sessions"`
	TotalSeconds clickhouseInt `json:"total_duration_seconds"`
	Freshness    string        `json:"freshness"`
}

type platformCallCenterQueryRow struct {
	TenantID     string        `json:"tenant_id"`
	AvatarID     string        `json:"avatar_id"`
	Channel      string        `json:"channel"`
	Completed    clickhouseInt `json:"completed_conversations"`
	TotalSeconds clickhouseInt `json:"total_duration_seconds"`
	Freshness    string        `json:"freshness"`
}

// ClickHouse JSON FORMAT JSON serializes UInt values as strings by default.
// Accept numbers too so this remains compatible with test and proxy servers.
type clickhouseInt int

func (value *clickhouseInt) UnmarshalJSON(data []byte) error {
	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*value = clickhouseInt(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return err
	}
	*value = clickhouseInt(parsed)
	return nil
}

func (c *Client) EnsureCallCenterSchema(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.call_center_usage_facts (
  fact_id String,
  tenant_id String,
  call_id String,
  conversation_record_id String,
  avatar_id String,
  channel String,
  source String,
  status String,
  started_at DateTime,
  ended_at DateTime,
  usage_date Date,
  duration_seconds UInt32,
  source_updated_at DateTime,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now(),
  created_by String DEFAULT 'system',
  updated_by String DEFAULT 'system'
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (tenant_id, usage_date, call_id, fact_id)`, quoteIdent(c.db))
	return c.exec(ctx, stmt)
}

func (c *Client) UpsertCallCenterFact(ctx context.Context, fact CallCenterUsageFact) error {
	if !c.Enabled() {
		return fmt.Errorf("clickhouse is not configured")
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(fact); err != nil {
		return fmt.Errorf("encode call center fact: %w", err)
	}
	query := fmt.Sprintf("INSERT INTO %s.call_center_usage_facts (fact_id, tenant_id, call_id, conversation_record_id, avatar_id, channel, source, status, started_at, ended_at, usage_date, duration_seconds, source_updated_at, created_at, updated_at, created_by, updated_by) FORMAT JSONEachRow", quoteIdent(c.db))
	return c.execInsert(ctx, query, body.Bytes())
}

func (c *Client) QueryCallCenterStats(ctx context.Context, tenantID, startDate, endDate string) (CallCenterStats, error) {
	if !c.Enabled() {
		return CallCenterStats{}, fmt.Errorf("clickhouse is not configured")
	}
	query := fmt.Sprintf(`
SELECT avatar_id, channel, count() AS sessions,
       sum(duration_seconds) AS total_duration_seconds,
       formatDateTime(max(updated_at), '%%Y-%%m-%%d %%H:%%i:%%s') AS freshness
FROM %s.call_center_usage_facts FINAL
WHERE tenant_id = '%s'
  AND usage_date BETWEEN toDate('%s') AND toDate('%s')
  AND status = 'archived'
GROUP BY avatar_id, channel
ORDER BY sessions DESC
FORMAT JSON`, quoteIdent(c.db), escape(tenantID), escape(startDate), escape(endDate))
	body, err := c.query(ctx, query)
	if err != nil {
		return CallCenterStats{}, err
	}
	var parsed struct {
		Data []callCenterQueryRow `json:"data"`
	}
	if err := jsonUnmarshal(body, &parsed); err != nil {
		return CallCenterStats{}, err
	}
	stats := CallCenterStats{}
	avatar := make(map[string]*CallCenterBucket)
	channel := make(map[string]*CallCenterBucket)
	for _, row := range parsed.Data {
		sessions := int(row.Sessions)
		totalSeconds := int(row.TotalSeconds)
		stats.CompletedConversations += sessions
		stats.TotalDurationSeconds += totalSeconds
		if sessions > 0 {
			bucket := avatar[row.AvatarID]
			if bucket == nil {
				bucket = &CallCenterBucket{ID: row.AvatarID}
				avatar[row.AvatarID] = bucket
			}
			bucket.CompletedConversations += sessions
			bucket.TotalDurationSeconds += totalSeconds
			bucket = channel[row.Channel]
			if bucket == nil {
				bucket = &CallCenterBucket{Channel: row.Channel}
				channel[row.Channel] = bucket
			}
			bucket.CompletedConversations += sessions
			bucket.TotalDurationSeconds += totalSeconds
		}
		if parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", row.Freshness); parseErr == nil && parsedTime.After(stats.Freshness) {
			stats.Freshness = parsedTime
		}
	}
	stats.AverageDurationSeconds = averageSeconds(stats.TotalDurationSeconds, stats.CompletedConversations)
	for _, bucket := range avatar {
		bucket.AverageDurationSeconds = averageSeconds(bucket.TotalDurationSeconds, bucket.CompletedConversations)
		stats.ByAvatar = append(stats.ByAvatar, *bucket)
	}
	for _, bucket := range channel {
		bucket.AverageDurationSeconds = averageSeconds(bucket.TotalDurationSeconds, bucket.CompletedConversations)
		stats.ByChannel = append(stats.ByChannel, *bucket)
	}
	sort.Slice(stats.ByAvatar, func(i, j int) bool {
		return stats.ByAvatar[i].CompletedConversations > stats.ByAvatar[j].CompletedConversations
	})
	sort.Slice(stats.ByChannel, func(i, j int) bool {
		return stats.ByChannel[i].CompletedConversations > stats.ByChannel[j].CompletedConversations
	})
	return stats, nil
}

// QueryPlatformCallCenterStats reads the same archived usage facts as the
// tenant endpoint, but groups them by tenant as well as avatar and channel.
// The query is aggregate-only by design and supports an exact tenant filter
// without changing the caller's platform-level authorization.
func (c *Client) QueryPlatformCallCenterStats(ctx context.Context, tenantID, startDate, endDate string) (PlatformCallCenterStats, error) {
	if !c.Enabled() {
		return PlatformCallCenterStats{}, fmt.Errorf("clickhouse is not configured")
	}
	tenantFilter := ""
	if strings.TrimSpace(tenantID) != "" {
		tenantFilter = "\n  AND tenant_id = '" + escape(tenantID) + "'"
	}
	query := fmt.Sprintf(`
SELECT tenant_id, avatar_id, channel, count() AS completed_conversations,
       sum(duration_seconds) AS total_duration_seconds,
       formatDateTime(max(updated_at), '%%Y-%%m-%%d %%H:%%i:%%s', 'UTC') AS freshness
FROM %s.call_center_usage_facts FINAL
WHERE usage_date BETWEEN toDate('%s') AND toDate('%s')
  AND status = 'archived'%s
GROUP BY tenant_id, avatar_id, channel
ORDER BY completed_conversations DESC, tenant_id, avatar_id, channel
FORMAT JSON`, quoteIdent(c.db), escape(startDate), escape(endDate), tenantFilter)
	body, err := c.query(ctx, query)
	if err != nil {
		return PlatformCallCenterStats{}, err
	}
	var parsed struct {
		Data []platformCallCenterQueryRow `json:"data"`
	}
	if err := jsonUnmarshal(body, &parsed); err != nil {
		return PlatformCallCenterStats{}, err
	}

	stats := PlatformCallCenterStats{}
	tenant := make(map[string]*PlatformCallCenterBucket)
	avatar := make(map[string]*PlatformCallCenterBucket)
	channel := make(map[string]*PlatformCallCenterBucket)
	for _, row := range parsed.Data {
		completed := int(row.Completed)
		totalSeconds := int(row.TotalSeconds)
		stats.CompletedConversations += completed
		stats.TotalDurationSeconds += totalSeconds

		tenantBucket := tenant[row.TenantID]
		if tenantBucket == nil {
			tenantBucket = &PlatformCallCenterBucket{TenantID: row.TenantID}
			tenant[row.TenantID] = tenantBucket
		}
		tenantBucket.CompletedConversations += completed
		tenantBucket.TotalDurationSeconds += totalSeconds

		avatarBucket := avatar[row.AvatarID]
		if avatarBucket == nil {
			avatarBucket = &PlatformCallCenterBucket{AvatarID: row.AvatarID}
			avatar[row.AvatarID] = avatarBucket
		}
		avatarBucket.CompletedConversations += completed
		avatarBucket.TotalDurationSeconds += totalSeconds

		channelBucket := channel[row.Channel]
		if channelBucket == nil {
			channelBucket = &PlatformCallCenterBucket{Channel: row.Channel}
			channel[row.Channel] = channelBucket
		}
		channelBucket.CompletedConversations += completed
		channelBucket.TotalDurationSeconds += totalSeconds

		if parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", row.Freshness); parseErr == nil && parsedTime.After(stats.LastProjectedAt) {
			stats.LastProjectedAt = parsedTime
		}
	}
	for _, bucket := range tenant {
		bucket.AverageDurationSeconds = averageSeconds(bucket.TotalDurationSeconds, bucket.CompletedConversations)
		stats.ByTenant = append(stats.ByTenant, *bucket)
	}
	for _, bucket := range avatar {
		bucket.AverageDurationSeconds = averageSeconds(bucket.TotalDurationSeconds, bucket.CompletedConversations)
		stats.ByAvatar = append(stats.ByAvatar, *bucket)
	}
	for _, bucket := range channel {
		bucket.AverageDurationSeconds = averageSeconds(bucket.TotalDurationSeconds, bucket.CompletedConversations)
		stats.ByChannel = append(stats.ByChannel, *bucket)
	}
	byActivity := func(a, b PlatformCallCenterBucket) bool {
		if a.CompletedConversations != b.CompletedConversations {
			return a.CompletedConversations > b.CompletedConversations
		}
		return platformBucketKey(a) < platformBucketKey(b)
	}
	sort.Slice(stats.ByTenant, func(i, j int) bool { return byActivity(stats.ByTenant[i], stats.ByTenant[j]) })
	sort.Slice(stats.ByAvatar, func(i, j int) bool { return byActivity(stats.ByAvatar[i], stats.ByAvatar[j]) })
	sort.Slice(stats.ByChannel, func(i, j int) bool { return byActivity(stats.ByChannel[i], stats.ByChannel[j]) })
	return stats, nil
}

func platformBucketKey(bucket PlatformCallCenterBucket) string {
	if bucket.TenantID != "" {
		return bucket.TenantID
	}
	if bucket.AvatarID != "" {
		return bucket.AvatarID
	}
	return bucket.Channel
}

func averageSeconds(total, count int) float64 {
	if count <= 0 {
		return 0
	}
	return float64(total) / float64(count)
}
