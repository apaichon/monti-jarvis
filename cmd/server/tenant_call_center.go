package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
)

type tenantCallCenterStatistics struct {
	Range                  map[string]string `json:"range"`
	Timezone               string            `json:"timezone"`
	Freshness              *time.Time        `json:"freshness,omitempty"`
	TotalCompleted         int               `json:"total_completed_conversations"`
	TotalDurationSeconds   int               `json:"total_duration_seconds"`
	AverageDurationSeconds float64           `json:"average_duration_seconds"`
	Topic                  string            `json:"topic"`
	ByAvatar               []map[string]any  `json:"by_avatar"`
	ByChannel              []map[string]any  `json:"by_channel"`
	ByTopic                []map[string]any  `json:"by_topic"`
	Quota                  any               `json:"quota"`
	DailyUsage             map[string]any    `json:"daily_usage"`
	CallLimits             any               `json:"call_limits"`
}

func (s *server) getTenantCallCenterStatistics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tz := "Asia/Bangkok"
	if s.store != nil {
		tz = s.store.TenantTimezone(r.Context(), tenantID)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
		tz = "UTC"
	}
	today := time.Now().In(loc).Format("2006-01-02")
	startDate := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endDate := strings.TrimSpace(r.URL.Query().Get("end_date"))
	if startDate == "" {
		startDate = today
	}
	if endDate == "" {
		endDate = today
	}
	start, startErr := time.Parse("2006-01-02", startDate)
	end, endErr := time.Parse("2006-01-02", endDate)
	if startErr != nil || endErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid date range", "code": "validation_error"})
		return
	}
	if start.After(end) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "start_date must be on or before end_date", "code": "validation_error"})
		return
	}
	topicFilter, topicErr := parseCallCenterTopicFilter(r.URL.Query().Get("topic"))
	if topicErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "topic must contain only lowercase letters, numbers, and underscores", "code": "validation_error"})
		return
	}
	if s.ch == nil || !s.ch.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "analytics unavailable", "code": "analytics_unavailable"})
		return
	}
	stats, err := s.ch.QueryCallCenterStats(r.Context(), tenantID, startDate, endDate, topicFilter)
	if err != nil {
		log.Printf("call center statistics tenant=%s: %v", tenantID, err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "analytics unavailable", "code": "analytics_unavailable"})
		return
	}
	avatarNames := map[string]string{}
	if s.store != nil {
		assignments, assignmentErr := s.store.ListTenantAvatarAssignments(r.Context(), tenantID)
		if assignmentErr == nil {
			for _, assignment := range assignments {
				if assignment.Avatar != nil {
					avatarNames[assignment.AvatarID] = assignment.Avatar.Name
				}
			}
		}
	}
	byAvatar := make([]map[string]any, 0, len(stats.ByAvatar))
	for _, bucket := range stats.ByAvatar {
		name := avatarNames[bucket.ID]
		if name == "" {
			name = bucket.ID
		}
		byAvatar = append(byAvatar, map[string]any{"id": bucket.ID, "name": name, "completed": bucket.CompletedConversations, "total_duration_seconds": bucket.TotalDurationSeconds, "average_duration_seconds": bucket.AverageDurationSeconds})
	}
	byChannel := make([]map[string]any, 0, len(stats.ByChannel))
	for _, bucket := range stats.ByChannel {
		byChannel = append(byChannel, map[string]any{"channel": bucket.Channel, "completed": bucket.CompletedConversations, "total_duration_seconds": bucket.TotalDurationSeconds, "average_duration_seconds": bucket.AverageDurationSeconds})
	}
	byTopic := make([]map[string]any, 0, len(stats.ByTopic))
	for _, bucket := range stats.ByTopic {
		topic := bucket.Topic
		if strings.TrimSpace(topic) == "" {
			topic = "unknown"
		}
		byTopic = append(byTopic, map[string]any{"topic": topic, "label": callCenterTopicLabel(topic), "completed": bucket.CompletedConversations, "total_duration_seconds": bucket.TotalDurationSeconds, "average_duration_seconds": bucket.AverageDurationSeconds})
	}
	selectedTopic := topicFilter
	if selectedTopic == "" {
		selectedTopic = "all"
	}
	out := tenantCallCenterStatistics{
		Range: map[string]string{"start_date": startDate, "end_date": endDate}, Timezone: tz, Topic: selectedTopic,
		TotalCompleted: stats.CompletedConversations, TotalDurationSeconds: stats.TotalDurationSeconds,
		AverageDurationSeconds: stats.AverageDurationSeconds, ByAvatar: byAvatar, ByChannel: byChannel, ByTopic: byTopic,
		DailyUsage: map[string]any{"call_minutes": 0, "timezone": tz},
	}
	if !stats.Freshness.IsZero() {
		freshness := stats.Freshness.UTC()
		out.Freshness = &freshness
	}
	if s.quota != nil {
		snapshot, quotaErr := s.quota.Snapshot(r.Context(), tenantID)
		if quotaErr != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "quota unavailable", "code": "quota_unavailable"})
			return
		}
		out.Quota = snapshot
		daily, _ := s.quota.GetDailyCallMinutes(r.Context(), tenantID, tz)
		out.DailyUsage["call_minutes"] = daily
	}
	if s.store != nil {
		if limits, limitErr := s.store.GetOrCreateTenantCallLimits(r.Context(), tenantID); limitErr == nil {
			out.CallLimits = limits
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) projectCallCenterRecord(ctx context.Context, tenantID, recordID string) {
	if s == nil || s.ch == nil || !s.ch.Enabled() || s.store == nil {
		return
	}
	analytics, err := s.store.GetConversationAnalyticsContext(ctx, tenantID, recordID)
	if err != nil || analytics.Record.EndedAt == nil {
		if err != nil {
			log.Printf("load call center record tenant=%s record=%s: %v", tenantID, recordID, err)
		}
		return
	}
	tz := s.store.TenantTimezone(ctx, tenantID)
	loc, locErr := time.LoadLocation(tz)
	if locErr != nil {
		loc = time.UTC
	}
	endedAt := analytics.Record.EndedAt.UTC()
	updatedAt := analytics.SourceUpdatedAt
	if updatedAt.IsZero() {
		updatedAt = endedAt
	}
	source := strings.TrimSpace(analytics.Source)
	if source == "" {
		source = "production"
	}
	fact := clickhouse.CallCenterUsageFact{
		FactID: "ccf_" + analytics.Record.ID, TenantID: analytics.Record.TenantID, CallID: analytics.Record.CallID,
		ConversationID: analytics.Record.ID, AvatarID: analytics.Record.AvatarID, Channel: analytics.Record.Channel,
		Topic: callCenterTopicFromSummary(analytics.Record.Summary), Source: source, Status: analytics.Record.Status, StartedAt: formatCHTime(analytics.Record.StartedAt),
		EndedAt: formatCHTime(endedAt), UsageDate: endedAt.In(loc).Format("2006-01-02"),
		DurationSeconds: analytics.Record.DurationSeconds, SourceUpdatedAt: formatCHTime(updatedAt),
		CreatedAt: formatCHTime(time.Now().UTC()), UpdatedAt: formatCHTime(updatedAt), CreatedBy: "system", UpdatedBy: "system",
	}
	if err := s.ch.UpsertCallCenterFact(ctx, fact); err != nil {
		log.Printf("project call center record tenant=%s record=%s: %v", tenantID, recordID, err)
	}
}

func (s *server) backfillCallCenterAnalytics(ctx context.Context) {
	if s == nil || s.ch == nil || !s.ch.Enabled() || s.store == nil {
		return
	}
	rows, err := s.store.ListCompletedConversationAnalytics(ctx)
	if err != nil {
		log.Printf("call center analytics backfill: %v", err)
		return
	}
	for _, item := range rows {
		s.projectCallCenterRecord(ctx, item.Record.TenantID, item.Record.ID)
	}
	if len(rows) > 0 {
		log.Printf("call center analytics backfill: projected %d conversation records", len(rows))
	}
}

func formatCHTime(value time.Time) string {
	if value.IsZero() {
		value = time.Now().UTC()
	}
	return value.UTC().Format("2006-01-02 15:04:05")
}

func callCenterTopicFromSummary(summary map[string]any) string {
	if summary == nil {
		return "unknown"
	}
	topic, _ := summary["topic"]
	return normalizeCallCenterTopic(topic)
}

func normalizeCallCenterTopic(value any) string {
	text, ok := value.(string)
	if !ok {
		return "unknown"
	}
	normalized, valid := normalizeCallCenterTopicText(text)
	if !valid || normalized == "" {
		return "unknown"
	}
	return normalized
}

func parseCallCenterTopicFilter(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	normalized, valid := normalizeCallCenterTopicText(value)
	if !valid || normalized == "" {
		return "", errInvalidCallCenterTopic
	}
	return normalized, nil
}

var errInvalidCallCenterTopic = errValidation("invalid call center topic")

type errValidation string

func (e errValidation) Error() string { return string(e) }

func normalizeCallCenterTopicText(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == ' ' || r == '\t' || r == '\n':
			if builder.Len() > 0 && !lastUnderscore {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		default:
			return "", false
		}
		if builder.Len() > 48 {
			return "", false
		}
	}
	normalized := strings.Trim(builder.String(), "_")
	if normalized == "" || len(normalized) > 48 {
		return "", false
	}
	return normalized, true
}

func callCenterTopicLabel(topic string) string {
	topic = normalizeCallCenterTopic(topic)
	if topic == "unknown" {
		return "Unknown"
	}
	parts := strings.Split(topic, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
