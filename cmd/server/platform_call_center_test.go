package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParsePlatformCallCenterQueryDefaultsAndBounds(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/platform/call-center/statistics", nil)
	start, end, tenantID, limit, offset, err := parsePlatformCallCenterQuery(req)
	if err != nil {
		t.Fatalf("parse default query: %v", err)
	}
	if start != end || tenantID != "" || limit != defaultPlatformCallCenterLimit || offset != 0 {
		t.Fatalf("unexpected defaults: start=%q end=%q tenant=%q limit=%d offset=%d", start, end, tenantID, limit, offset)
	}
	if _, parseErr := time.Parse("2006-01-02", start); parseErr != nil {
		t.Fatalf("default date is invalid: %v", parseErr)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/platform/call-center/statistics?start_date=2026-07-01&end_date=2026-07-31&tenant_id=tenant-1&limit=100&offset=25", nil)
	start, end, tenantID, limit, offset, err = parsePlatformCallCenterQuery(req)
	if err != nil || start != "2026-07-01" || end != "2026-07-31" || tenantID != "tenant-1" || limit != 100 || offset != 25 {
		t.Fatalf("unexpected bounded query: %q %q %q %d %d %v", start, end, tenantID, limit, offset, err)
	}
}

func TestParsePlatformCallCenterQueryRejectsInvalidValues(t *testing.T) {
	for name, path := range map[string]string{
		"invalid date":   "/api/platform/call-center/statistics?start_date=2026-02-30",
		"reversed range": "/api/platform/call-center/statistics?start_date=2026-07-02&end_date=2026-07-01",
		"limit":          "/api/platform/call-center/statistics?limit=101",
		"offset":         "/api/platform/call-center/statistics?offset=-1",
		"tenant id":      "/api/platform/call-center/statistics?tenant_id=123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
	} {
		t.Run(name, func(t *testing.T) {
			_, _, _, _, _, err := parsePlatformCallCenterQuery(httptest.NewRequest(http.MethodGet, path, nil))
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestPlatformCallCenterFreshnessStatus(t *testing.T) {
	if got := platformCallCenterFreshnessStatus(time.Time{}); got != "empty" {
		t.Fatalf("empty status = %q", got)
	}
	if got := platformCallCenterFreshnessStatus(time.Now().Add(-6 * time.Minute)); got != "stale" {
		t.Fatalf("stale status = %q", got)
	}
	if got := platformCallCenterFreshnessStatus(time.Now().Add(-time.Minute)); got != "current" {
		t.Fatalf("current status = %q", got)
	}
}

func TestRangeCallMinutesRoundsUp(t *testing.T) {
	if got := rangeCallMinutes(0); got != 0 {
		t.Fatalf("zero duration = %d", got)
	}
	if got := rangeCallMinutes(61); got != 2 {
		t.Fatalf("61 seconds = %d minutes, want 2", got)
	}
}
