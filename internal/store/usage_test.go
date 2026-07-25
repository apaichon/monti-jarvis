package store

import (
	"errors"
	"testing"
	"time"
)

func TestNormalizeUsageEvent(t *testing.T) {
	in, err := normalizeUsageEvent(UsageEventInput{
		TenantID: " tenant-a ", IdempotencyKey: " call-1 ", Dimension: "monthly_call_minutes", Unit: "minutes",
		Amount: 3, SourceType: "call", SourceID: "call-1",
		PeriodStart: time.Date(2026, 7, 25, 18, 30, 0, 0, time.FixedZone("ICT", 7*60*60)),
	})
	if err != nil {
		t.Fatalf("normalize usage event: %v", err)
	}
	if in.TenantID != "tenant-a" || in.IdempotencyKey != "call-1" || in.State != UsageStateApplied {
		t.Fatalf("unexpected normalized event: %+v", in)
	}
	if in.PeriodStart.Hour() != 0 || in.PeriodStart.Location() != time.UTC {
		t.Fatalf("period was not normalized to UTC date: %v", in.PeriodStart)
	}
}

func TestNormalizeUsageEventRejectsInvalidDimensionAndCorrection(t *testing.T) {
	base := UsageEventInput{TenantID: "tenant-a", IdempotencyKey: "k", Dimension: "monthly_call_minutes", Unit: "bytes", Amount: 1, SourceType: "call"}
	if !errors.Is(func() error { _, err := normalizeUsageEvent(base); return err }(), ErrUsageValidation) {
		t.Fatal("expected invalid dimension/unit to be rejected")
	}
	base.Unit = "minutes"
	base.State = UsageStateCorrection
	if !errors.Is(func() error { _, err := normalizeUsageEvent(base); return err }(), ErrUsageValidation) {
		t.Fatal("expected correction without correction_of to be rejected")
	}
}
