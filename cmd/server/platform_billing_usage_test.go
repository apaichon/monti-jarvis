package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/quota"
)

func TestParsePlatformBillingUsageQueryDefaultsAndBounds(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/platform/billing/usage?limit=25&offset=10", nil)
	start, end, tenant, limit, offset, err := parsePlatformBillingUsageQuery(req)
	if err != nil {
		t.Fatalf("parse defaults: %v", err)
	}
	if start == "" || end == "" || tenant != "" || limit != 25 || offset != 10 {
		t.Fatalf("unexpected query values: %q %q %q %d %d", start, end, tenant, limit, offset)
	}
	if _, err := time.Parse("2006-01-02", start); err != nil {
		t.Fatalf("start date: %v", err)
	}
}

func TestParsePlatformBillingUsageQueryRejectsUnsafeRanges(t *testing.T) {
	cases := []string{
		"?start_date=2026-02-01&end_date=2026-01-01",
		"?start_date=2025-01-01&end_date=2026-01-03",
		"?limit=101",
		"?offset=-1",
		"?tenant_id=" + "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	for _, query := range cases {
		req := httptest.NewRequest("GET", "/api/platform/billing/usage"+query, nil)
		if _, _, _, _, _, err := parsePlatformBillingUsageQuery(req); err == nil {
			t.Errorf("expected validation error for %s", query)
		}
	}
}

func TestQuotaUsageResponseMakesDailyGapExplicit(t *testing.T) {
	got := quotaUsageResponse(&quota.Snapshot{
		Status: "active",
		Limits: &quota.Limits{MaxMonthlyCallMinutes: 500},
		Usage:  quota.Usage{MonthlyCallMinutes: 37},
	})
	if got.Status != "current" || got.MonthlyUsed != 37 || got.MonthlyLimit != 500 {
		t.Fatalf("unexpected monthly quota: %+v", got)
	}
	if got.DailyStatus != "unavailable" {
		t.Fatalf("expected explicit daily status, got %q", got.DailyStatus)
	}
}

func TestAICostCoverageResponse(t *testing.T) {
	got := platformAICostResponse(clickhouse.AIUsageStats{
		Events:             4,
		ObservedEvents:     2,
		EstimatedEvents:    1,
		UnavailableEvents:  1,
		ObservedCostMicros: 12,
	}, "rate-2026-07", "2026-07-01", "THB")
	if got.Currency != "THB" || got.RateVersion != "rate-2026-07" || got.PricingAsOf != "2026-07-01" || got.ObservedCostMicros != 12 || got.CoveragePercent != 75 || got.Status != "warning" {
		t.Fatalf("unexpected AI cost response: %+v", got)
	}
}
