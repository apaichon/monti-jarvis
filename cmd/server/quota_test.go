package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/quota"
)

func TestWriteQuotaError_LimitExceeded(t *testing.T) {
	rr := httptest.NewRecorder()
	err := &quota.Error{
		Code:      "quota_exceeded",
		Dimension: "max_km_documents",
		Limit:     50,
		Usage:     50,
		Message:   "max_km_documents limit exceeded (50/50)",
	}
	// Wrap cause for errors.Is
	type causer interface{ Unwrap() error }
	_ = causer(err)

	writeQuotaError(rr, err)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "quota_exceeded" {
		t.Fatalf("body %#v", body)
	}
	if body["dimension"] != "max_km_documents" {
		t.Fatalf("dimension %#v", body)
	}
}

func TestWriteQuotaError_RateLimited(t *testing.T) {
	rr := httptest.NewRecorder()
	writeQuotaError(rr, &quota.Error{
		Code: "rate_limited", Dimension: "chat", Limit: 60, Usage: 61, Message: "rate limit",
	})
	if rr.Code != 429 {
		t.Fatalf("status %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After")
	}
}

func TestWriteQuotaError_FeatureDisabled(t *testing.T) {
	rr := httptest.NewRecorder()
	writeQuotaError(rr, &quota.Error{
		Code: "feature_disabled", Dimension: "voice_enabled", Message: "off",
	})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestWriteQuotaError_PlainLimit(t *testing.T) {
	rr := httptest.NewRecorder()
	writeQuotaError(rr, quota.ErrLimitExceeded)
	if rr.Code != 429 {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestWriteQuotaError_Unknown(t *testing.T) {
	rr := httptest.NewRecorder()
	writeQuotaError(rr, errors.New("boom"))
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status %d", rr.Code)
	}
}
