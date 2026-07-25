package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/store"
)

const maxUsageReconciliationDays = 31

type usageReconciliationRequest struct {
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	TenantID       string `json:"tenant_id,omitempty"`
	DryRun         bool   `json:"dry_run"`
	IdempotencyKey string `json:"idempotency_key"`
}

func parseUsageReconciliationRequest(body []byte) (store.UsageReconciliationInput, error) {
	var req usageReconciliationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return store.UsageReconciliationInput{}, store.ErrUsageValidation
	}
	start, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return store.UsageReconciliationInput{}, store.ErrUsageValidation
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return store.UsageReconciliationInput{}, store.ErrUsageValidation
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" || end.Before(start) || end.Sub(start) >= maxUsageReconciliationDays*24*time.Hour {
		return store.UsageReconciliationInput{}, store.ErrUsageValidation
	}
	return store.UsageReconciliationInput{TenantID: strings.TrimSpace(req.TenantID), IdempotencyKey: strings.TrimSpace(req.IdempotencyKey), StartDate: start, EndDate: end, DryRun: req.DryRun}, nil
}

func (s *server) startUsageReconciliation(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "usage_source_unavailable"})
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil && !errors.Is(err, context.Canceled) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_reconciliation_request"})
		return
	}
	input, err := parseUsageReconciliationRequest(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_reconciliation_request"})
		return
	}
	run, duplicate, err := s.store.CreateUsageReconciliationRun(r.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrUsageValidation) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_reconciliation_request"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "usage_source_unavailable"})
		return
	}
	if !duplicate {
		go func(runID string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			_ = s.store.RunUsageReconciliation(ctx, runID)
		}(run.ID)
	}
	status := http.StatusAccepted
	if duplicate {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{"run_id": run.ID, "status": run.Status, "duplicate": duplicate})
}

func (s *server) getUsageReconciliation(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "usage_source_unavailable"})
		return
	}
	run, err := s.store.GetUsageReconciliationRun(r.Context(), r.PathValue("run_id"))
	if errors.Is(err, store.ErrUsageReconciliationNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"code": "reconciliation_not_found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "usage_source_unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, run)
}
