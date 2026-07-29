package main

import (
	"testing"

	"github.com/libra/monti-jarvis/internal/quota"
)

func intPointer(value int) *int { return &value }

func TestCurrentPlanQuotaDimensionsUsesHighestReliableFiniteDimension(t *testing.T) {
	snapshot := &quota.Snapshot{
		Period: "2026-08",
		Dimensions: []quota.Dimension{
			{
				Dimension: "ai_employees", Unit: "assignments", Period: "2026-08",
				TotalLimit: 10, Consumed: intPointer(3), Remaining: intPointer(7),
				Source: "postgres", Freshness: "current",
			},
			{
				Dimension: "concurrent_calls", Unit: "calls", Period: "2026-08",
				TotalLimit: 4, Consumed: intPointer(3), Remaining: intPointer(1),
				Source: "redis", Freshness: "current",
			},
			{
				Dimension: "monthly_call_minutes", Unit: "minutes", Period: "2026-08",
				TotalLimit: 99_999_999, Consumed: intPointer(100), Remaining: intPointer(99_999_899),
				Source: "postgres", Freshness: "current",
			},
		},
	}
	rows, compact := currentPlanQuotaDimensions(snapshot)
	if len(rows) != 3 {
		t.Fatalf("expected three rows, got %d", len(rows))
	}
	if compact == nil || *compact != 0.75 {
		t.Fatalf("expected highest utilization 0.75, got %v", compact)
	}
	if !rows[2].Unlimited || rows[2].Limit != nil || rows[2].Utilization != nil {
		t.Fatalf("unlimited quota must not affect utilization: %+v", rows[2])
	}
}

func TestCurrentPlanQuotaDimensionsDoesNotRenderUnavailableAsZero(t *testing.T) {
	snapshot := &quota.Snapshot{
		Period: "2026-08",
		Dimensions: []quota.Dimension{{
			Dimension: "storage_bytes", Unit: "bytes", Period: "2026-08",
			TotalLimit: 1024, Consumed: nil, Remaining: nil,
			Source: "unavailable", Freshness: "unavailable",
		}},
	}
	rows, compact := currentPlanQuotaDimensions(snapshot)
	if len(rows) != 1 || rows[0].Used != nil || rows[0].Remaining != nil || rows[0].Utilization != nil {
		t.Fatalf("unavailable dimension must remain nullable: %+v", rows)
	}
	if compact != nil {
		t.Fatalf("compact utilization must be nil, got %v", *compact)
	}
}
