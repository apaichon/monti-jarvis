package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/audit"
	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/clickhouse"
)

func (s *server) listPlatformAuditLogs(w http.ResponseWriter, r *http.Request) {
	if s.ch == nil || !s.ch.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "analytics unavailable", "code": "analytics_unavailable"})
		return
	}
	filter, err := parseAuditFilter(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "validation_error"})
		return
	}
	page, err := s.ch.ListAuditEvents(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "analytics unavailable", "code": "analytics_unavailable"})
		return
	}
	nextCursor := ""
	if page.HasMore {
		nextCursor = encodeAuditCursor(filter.Offset + filter.Limit)
	}
	startDate, endDate := filter.StartDate, filter.EndDate
	writeJSON(w, http.StatusOK, map[string]any{
		"events":      page.Events,
		"next_cursor": nextCursor,
		"range": map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
			"timezone":   "UTC",
		},
	})
}

func (s *server) platformAuditHealth(w http.ResponseWriter, r *http.Request) {
	health := audit.Health{Mode: s.cfg.AuditLogMode}
	if s.audit != nil {
		health = s.audit.Health()
	}
	clickhouseStatus := "disabled"
	if s.ch != nil && s.ch.Enabled() {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		err := s.ch.Ping(ctx)
		cancel()
		if err != nil {
			clickhouseStatus = "unavailable"
		} else if health.PendingFiles > 0 || health.FailedFiles > 0 {
			clickhouseStatus = "degraded"
		} else {
			clickhouseStatus = "operational"
		}
	}
	var lastTransfer any
	if health.LastSuccessfulTransfer != nil {
		lastTransfer = health.LastSuccessfulTransfer.UTC().Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"mode":                            health.Mode,
		"queue_depth":                     health.QueueDepth,
		"last_successful_transfer":        lastTransfer,
		"oldest_pending_file_age_seconds": int64(health.OldestPendingFileAge.Seconds()),
		"pending_files":                   health.PendingFiles,
		"failed_files":                    health.FailedFiles,
		"clickhouse":                      clickhouseStatus,
	})
}

func parseAuditFilter(r *http.Request) (clickhouse.AuditEventFilter, error) {
	query := r.URL.Query()
	startDate := strings.TrimSpace(query.Get("start_date"))
	endDate := strings.TrimSpace(query.Get("end_date"))
	today := time.Now().UTC().Format("2006-01-02")
	if endDate == "" {
		endDate = today
	}
	if startDate == "" {
		end, _ := time.Parse("2006-01-02", endDate)
		startDate = end.AddDate(0, 0, -7).Format("2006-01-02")
	}
	for name, value := range map[string]string{"start_date": startDate, "end_date": endDate} {
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return clickhouse.AuditEventFilter{}, fmt.Errorf("%s must be YYYY-MM-DD", name)
		}
	}
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	if start.After(end) {
		return clickhouse.AuditEventFilter{}, fmt.Errorf("start_date must be on or before end_date")
	}
	limit := 50
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 200 {
			return clickhouse.AuditEventFilter{}, fmt.Errorf("limit must be between 1 and 200")
		}
		limit = parsed
	}
	offset, err := decodeAuditCursor(query.Get("cursor"))
	if err != nil {
		return clickhouse.AuditEventFilter{}, fmt.Errorf("cursor is invalid")
	}
	filter := clickhouse.AuditEventFilter{
		TenantID: queryValue(query.Get("tenant_id"), 128), ActorID: queryValue(query.Get("actor_id"), 128),
		Action: queryValue(query.Get("action"), 128), ResourceType: queryValue(query.Get("resource_type"), 128),
		Outcome: queryValue(query.Get("outcome"), 32), StartDate: startDate, EndDate: endDate, Limit: limit, Offset: offset,
	}
	if filter.Outcome != "" && filter.Outcome != "success" && filter.Outcome != "denied" && filter.Outcome != "failure" {
		return clickhouse.AuditEventFilter{}, fmt.Errorf("outcome is invalid")
	}
	return filter, nil
}

func queryValue(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

func encodeAuditCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeAuditCursor(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return 0, err
	}
	offset, err := strconv.Atoi(string(b))
	if err != nil || offset < 0 || offset > 1_000_000 {
		return 0, err
	}
	return offset, nil
}

func (s *server) auditActor(r *http.Request) audit.Actor {
	actor := audit.Actor{ID: "anonymous", Type: "anonymous", TenantID: strings.TrimSpace(r.Header.Get("X-Tenant-Id"))}
	if s.auth != nil && s.auth.Enabled() {
		if ac, err := s.auth.ParseBearer(r.Header.Get("Authorization")); err == nil {
			actor.ID = ac.UserID
			actor.TenantID = ac.TenantID
			switch ac.Role {
			case auth.RolePlatformAdmin:
				actor.Type = "platform_admin"
			case auth.RoleTenantAdmin:
				actor.Type = "tenant_admin"
			case auth.RoleCustomer:
				actor.Type = "customer"
			}
		}
	}
	if actor.TenantID == "" {
		actor.TenantID = tenantIDFromPath(r.URL.Path)
	}
	if actor.TenantID == "" {
		actor.TenantID = strings.TrimSpace(s.cfg.DemoTenantID)
	}
	return actor
}

func tenantIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := range parts[:max(0, len(parts)-1)] {
		if parts[i] == "tenants" && i+1 < len(parts) {
			return strings.TrimSpace(parts[i+1])
		}
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
