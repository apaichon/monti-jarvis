package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/observability"
)

func TestTenantSystemPerformanceRequiresTenantContext(t *testing.T) {
	s := &server{monitoring: observability.New(nil, time.Second)}
	rec := httptest.NewRecorder()
	s.getTenantSystemPerformance(rec, httptest.NewRequest(http.MethodGet, "/api/tenant/system-performance", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTenantSystemPerformanceRedactsProbeErrors(t *testing.T) {
	s := &server{monitoring: observability.New([]observability.Dependency{
		{Name: "postgres", Probe: func(context.Context) (bool, error) {
			return true, errors.New("password=super-secret host=db.internal")
		}},
	}, time.Second)}
	ctx := auth.WithContext(context.Background(), auth.AuthContext{TenantID: "tenant-a", Role: auth.RoleTenantAdmin})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tenant/system-performance", nil).WithContext(ctx)
	s.getTenantSystemPerformance(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "super-secret") || strings.Contains(body, "db.internal") {
		t.Fatalf("response leaked probe detail: %s", body)
	}
	if !strings.Contains(body, `"status":"unavailable"`) || !strings.Contains(body, `"overall_status":"unavailable"`) {
		t.Fatalf("response did not normalize probe failure: %s", body)
	}
}
