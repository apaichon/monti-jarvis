package main

import (
	"net/http/httptest"
	"testing"
)

func TestParseAuditFilterDefaultsAndCursor(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/platform/audit-logs?limit=20&cursor=MjA&outcome=denied", nil)
	filter, err := parseAuditFilter(req)
	if err != nil {
		t.Fatal(err)
	}
	if filter.Limit != 20 || filter.Offset != 20 || filter.Outcome != "denied" || filter.StartDate == "" || filter.EndDate == "" {
		t.Fatalf("unexpected filter: %#v", filter)
	}
}

func TestParseAuditFilterRejectsInvalidRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/platform/audit-logs?start_date=2026-07-17&end_date=2026-07-16", nil)
	if _, err := parseAuditFilter(req); err == nil {
		t.Fatal("expected invalid date range")
	}
}
