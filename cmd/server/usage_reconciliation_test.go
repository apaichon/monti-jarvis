package main

import (
	"errors"
	"testing"

	"github.com/libra/monti-jarvis/internal/store"
)

func TestParseUsageReconciliationRequest(t *testing.T) {
	run, err := parseUsageReconciliationRequest([]byte(`{"start_date":"2026-07-01","end_date":"2026-07-07","tenant_id":"tenant-a","dry_run":true,"idempotency_key":"reconcile-1"}`))
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	if run.TenantID != "tenant-a" || run.StartDate.Day() != 1 || run.EndDate.Day() != 7 {
		t.Fatalf("unexpected reconciliation input: %+v", run)
	}
}

func TestParseUsageReconciliationRequestBoundsPeriod(t *testing.T) {
	for _, body := range []string{
		`{"start_date":"2026-07-01","end_date":"2026-07-01"}`,
		`{"start_date":"2026-07-01","end_date":"2026-08-10","idempotency_key":"too-long"}`,
		`{"start_date":"2026-07-08","end_date":"2026-07-01","idempotency_key":"backwards"}`,
	} {
		_, err := parseUsageReconciliationRequest([]byte(body))
		if !errors.Is(err, store.ErrUsageValidation) {
			t.Fatalf("expected validation error for %s, got %v", body, err)
		}
	}
}

func TestParseUsageReconciliationRequestAcceptsBoundedThirtyOneDayWindow(t *testing.T) {
	_, err := parseUsageReconciliationRequest([]byte(`{"start_date":"2026-07-01","end_date":"2026-07-31","idempotency_key":"bounded"}`))
	if err != nil {
		t.Fatalf("31 calendar-day window should be accepted: %v", err)
	}
	_, err = parseUsageReconciliationRequest([]byte(`{"start_date":"2026-07-01","end_date":"2026-08-01","idempotency_key":"too-wide"}`))
	if !errors.Is(err, store.ErrUsageValidation) {
		t.Fatalf("expected 32-day window to be rejected, got %v", err)
	}
}
