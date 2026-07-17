package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/observability"
)

func TestParsePlatformMonitoringQueryDefaultsAndBounds(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/platform/system-performance", nil)
	tenantID, status, limit, offset, err := parsePlatformMonitoringQuery(req)
	if err != nil {
		t.Fatalf("parse default query: %v", err)
	}
	if tenantID != "" || status != "" || limit != defaultPlatformMonitoringLimit || offset != 0 {
		t.Fatalf("unexpected defaults: tenant=%q status=%q limit=%d offset=%d", tenantID, status, limit, offset)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/platform/system-performance?tenant_id=tenant-1&status=degraded&limit=100&offset=25", nil)
	tenantID, status, limit, offset, err = parsePlatformMonitoringQuery(req)
	if err != nil || tenantID != "tenant-1" || status != observability.StatusDegraded || limit != 100 || offset != 25 {
		t.Fatalf("unexpected bounded query: %q %q %d %d %v", tenantID, status, limit, offset, err)
	}
}

func TestParsePlatformMonitoringQueryRejectsInvalidValues(t *testing.T) {
	for name, path := range map[string]string{
		"status": "/api/platform/system-performance?status=stale",
		"limit":  "/api/platform/system-performance?limit=101",
		"offset": "/api/platform/system-performance?offset=-1",
	} {
		t.Run(name, func(t *testing.T) {
			_, _, _, _, err := parsePlatformMonitoringQuery(httptest.NewRequest(http.MethodGet, path, nil))
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestPlatformTenantStatusRedactsAndNormalizesStates(t *testing.T) {
	components := []observability.Component{
		{Name: "postgres", Status: observability.StatusOperational},
		{Name: "redis", Status: observability.StatusOperational},
	}
	if got := platformTenantStatus(components, observability.AnalyticsCurrent, observability.StatusOperational); got != observability.StatusOperational {
		t.Fatalf("healthy tenant status = %q", got)
	}
	if got := platformTenantStatus(components, observability.AnalyticsStale, observability.StatusOperational); got != observability.StatusDegraded {
		t.Fatalf("stale tenant status = %q", got)
	}
	components[0].Status = observability.StatusUnavailable
	if got := platformTenantStatus(components, observability.AnalyticsCurrent, observability.StatusOperational); got != observability.StatusUnavailable {
		t.Fatalf("unavailable tenant status = %q", got)
	}
}

func TestPlatformMonitoringUnavailableWithoutStore(t *testing.T) {
	s := &server{}
	rec := httptest.NewRecorder()
	s.getPlatformSystemPerformance(rec, httptest.NewRequest(http.MethodGet, "/api/platform/system-performance", nil).WithContext(context.Background()))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "monitoring unavailable" || body["code"] != "monitoring_unavailable" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
