package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libra/monti-jarvis/internal/audit"
	"github.com/libra/monti-jarvis/internal/observability"
	"github.com/libra/monti-jarvis/internal/store"
)

const (
	defaultPlatformMonitoringLimit = 50
	maxPlatformMonitoringLimit     = 100
	maxPlatformMonitoringTenants   = 1000
	platformMonitoringPageSize     = 100
	platformMonitoringConcurrency  = 16
)

type platformPerformanceResponse struct {
	OverallStatus string                      `json:"overall_status"`
	CheckedAt     time.Time                   `json:"checked_at"`
	Summary       platformPerformanceSummary  `json:"summary"`
	Components    []observability.Component   `json:"components"`
	Audit         platformPerformanceAudit    `json:"audit"`
	Tenants       []platformPerformanceTenant `json:"tenants"`
	Total         int                         `json:"total"`
	Limit         int                         `json:"limit"`
	Offset        int                         `json:"offset"`
}

type platformPerformanceSummary struct {
	TenantsTotal int `json:"tenants_total"`
	Operational  int `json:"operational"`
	Degraded     int `json:"degraded"`
	Unavailable  int `json:"unavailable"`
}

type platformPerformanceAudit struct {
	Mode                        string  `json:"mode"`
	Status                      string  `json:"status"`
	PendingFiles                int     `json:"pending_files"`
	FailedFiles                 int     `json:"failed_files"`
	OldestPendingFileAgeSeconds int64   `json:"oldest_pending_file_age_seconds"`
	LastSuccessfulTransfer      *string `json:"last_successful_transfer"`
}

type platformPerformanceTenant struct {
	TenantID    string                  `json:"tenant_id"`
	Slug        string                  `json:"slug"`
	Name        string                  `json:"name"`
	Status      string                  `json:"status"`
	Analytics   observability.Analytics `json:"analytics"`
	AuditStatus string                  `json:"audit_status"`
}

func (s *server) getPlatformSystemPerformance(w http.ResponseWriter, r *http.Request) {
	if s.store == nil || s.monitoring == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "monitoring unavailable",
			"code":  "monitoring_unavailable",
		})
		return
	}

	tenantFilter, statusFilter, limit, offset, err := parsePlatformMonitoringQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "validation_error"})
		return
	}

	timeout := s.cfg.MonitoringProbeTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	shared := s.monitoring.Snapshot(ctx, nil)
	items, err := s.listActiveMonitoringTenants(ctx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "monitoring unavailable",
			"code":  "monitoring_unavailable",
		})
		return
	}

	filteredItems := make([]store.TenantListItem, 0, len(items))
	for _, item := range items {
		if tenantFilter != "" && item.ID != tenantFilter {
			continue
		}
		filteredItems = append(filteredItems, item)
	}

	analytics := make([]observability.Analytics, len(filteredItems))
	reader := s.analyticsHealthReader
	sem := make(chan struct{}, platformMonitoringConcurrency)
	var wg sync.WaitGroup
	for i, item := range filteredItems {
		i, tenantID := i, item.ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				analytics[i] = observability.Analytics{Status: observability.AnalyticsUnavailable}
				return
			}
			defer func() { <-sem }()
			analytics[i] = reader(tenantID)(ctx)
		}()
	}
	wg.Wait()

	clickhouseStatus := componentStatus(shared.Components, "clickhouse")
	auditSummary := s.platformPerformanceAuditSummary(clickhouseStatus)
	rows := make([]platformPerformanceTenant, 0, len(filteredItems))
	for i, item := range filteredItems {
		status := platformTenantStatus(shared.Components, analytics[i].Status, auditSummary.Status)
		if statusFilter != "" && status != statusFilter {
			continue
		}
		rows = append(rows, platformPerformanceTenant{
			TenantID:    item.ID,
			Slug:        item.Slug,
			Name:        item.Name,
			Status:      status,
			Analytics:   analytics[i],
			AuditStatus: auditSummary.Status,
		})
	}

	response := platformPerformanceResponse{
		OverallStatus: platformOverallStatus(shared.Components, auditSummary.Status, rows),
		CheckedAt:     shared.CheckedAt,
		Components:    shared.Components,
		Audit:         auditSummary,
		Limit:         limit,
		Offset:        offset,
	}
	response.Summary.TenantsTotal = len(rows)
	for _, row := range rows {
		switch row.Status {
		case observability.StatusOperational:
			response.Summary.Operational++
		case observability.StatusDegraded:
			response.Summary.Degraded++
		case observability.StatusUnavailable:
			response.Summary.Unavailable++
		}
	}
	response.Total = len(rows)
	if offset < len(rows) {
		end := offset + limit
		if end > len(rows) {
			end = len(rows)
		}
		response.Tenants = rows[offset:end]
	}
	if response.Tenants == nil {
		response.Tenants = []platformPerformanceTenant{}
	}
	writeJSON(w, http.StatusOK, response)
}

func parsePlatformMonitoringQuery(r *http.Request) (tenantID, status string, limit, offset int, err error) {
	query := r.URL.Query()
	tenantID = strings.TrimSpace(query.Get("tenant_id"))
	if len(tenantID) > 128 {
		return "", "", 0, 0, strconv.ErrSyntax
	}
	status = strings.TrimSpace(query.Get("status"))
	switch status {
	case "", observability.StatusOperational, observability.StatusDegraded, observability.StatusUnavailable:
	default:
		return "", "", 0, 0, errInvalidPlatformMonitoringStatus
	}

	limit = defaultPlatformMonitoringLimit
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > maxPlatformMonitoringLimit {
			return "", "", 0, 0, errInvalidPlatformMonitoringLimit
		}
	}
	offset = 0
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 || offset > 1_000_000 {
			return "", "", 0, 0, errInvalidPlatformMonitoringOffset
		}
	}
	return tenantID, status, limit, offset, nil
}

var (
	errInvalidPlatformMonitoringStatus = &platformMonitoringQueryError{"status is invalid"}
	errInvalidPlatformMonitoringLimit  = &platformMonitoringQueryError{"limit must be between 1 and 100"}
	errInvalidPlatformMonitoringOffset = &platformMonitoringQueryError{"offset must be a non-negative bounded integer"}
)

type platformMonitoringQueryError struct{ message string }

func (e *platformMonitoringQueryError) Error() string { return e.message }

func (s *server) listActiveMonitoringTenants(ctx context.Context) ([]store.TenantListItem, error) {
	items := make([]store.TenantListItem, 0)
	total := 0
	for offset := 0; offset < maxPlatformMonitoringTenants; offset += platformMonitoringPageSize {
		page, pageTotal, err := s.store.ListTenants(ctx, "active", "", platformMonitoringPageSize, offset)
		if err != nil {
			return nil, err
		}
		if offset == 0 {
			total = pageTotal
		}
		items = append(items, page...)
		if len(page) == 0 || offset+len(page) >= total || len(items) >= maxPlatformMonitoringTenants {
			break
		}
	}
	if len(items) > maxPlatformMonitoringTenants {
		items = items[:maxPlatformMonitoringTenants]
	}
	return items, nil
}

func (s *server) platformPerformanceAuditSummary(clickhouseStatus string) platformPerformanceAudit {
	health := audit.Health{Mode: s.cfg.AuditLogMode}
	if s.audit != nil {
		health = s.audit.Health()
	}
	status := clickhouseStatus
	if status == observability.StatusOperational && (health.PendingFiles > 0 || health.FailedFiles > 0) {
		status = observability.StatusDegraded
	}
	var lastTransfer *string
	if health.LastSuccessfulTransfer != nil {
		value := health.LastSuccessfulTransfer.UTC().Format(time.RFC3339)
		lastTransfer = &value
	}
	return platformPerformanceAudit{
		Mode:                        health.Mode,
		Status:                      status,
		PendingFiles:                health.PendingFiles,
		FailedFiles:                 health.FailedFiles,
		OldestPendingFileAgeSeconds: int64(health.OldestPendingFileAge.Seconds()),
		LastSuccessfulTransfer:      lastTransfer,
	}
}

func componentStatus(components []observability.Component, name string) string {
	for _, component := range components {
		if component.Name == name {
			return component.Status
		}
	}
	return observability.StatusDisabled
}

func platformTenantStatus(components []observability.Component, analyticsStatus, auditStatus string) string {
	for _, component := range components {
		if component.Status == observability.StatusUnavailable {
			return observability.StatusUnavailable
		}
	}
	if analyticsStatus == observability.AnalyticsUnavailable {
		return observability.StatusUnavailable
	}
	if analyticsStatus == observability.AnalyticsStale || analyticsStatus == observability.AnalyticsDisabled || auditStatus != observability.StatusOperational {
		return observability.StatusDegraded
	}
	for _, component := range components {
		if component.Status != observability.StatusOperational {
			return observability.StatusDegraded
		}
	}
	return observability.StatusOperational
}

func platformOverallStatus(components []observability.Component, auditStatus string, rows []platformPerformanceTenant) string {
	for _, component := range components {
		if component.Status == observability.StatusUnavailable {
			return observability.StatusUnavailable
		}
	}
	if auditStatus == observability.StatusUnavailable {
		return observability.StatusUnavailable
	}
	if auditStatus != observability.StatusOperational {
		return observability.StatusDegraded
	}
	for _, component := range components {
		if component.Status != observability.StatusOperational {
			return observability.StatusDegraded
		}
	}
	for _, row := range rows {
		if row.Status == observability.StatusUnavailable {
			return observability.StatusUnavailable
		}
		if row.Status != observability.StatusOperational {
			return observability.StatusDegraded
		}
	}
	return observability.StatusOperational
}
