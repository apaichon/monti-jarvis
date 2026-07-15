package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/observability"
	"github.com/libra/monti-jarvis/internal/store"
)

const monitoringAnalyticsStaleAfter = 5 * time.Minute

func newMonitoringService(st *store.Store, ch *clickhouse.Client, bus *natsbus.Bus, ai *gemini.Client, cfg env.Config) *observability.Service {
	storeProbe := func(name string) observability.Probe {
		return func(ctx context.Context) (bool, error) {
			if st == nil {
				return false, nil
			}
			return st.ProbeDependency(ctx, name)
		}
	}
	dependencies := []observability.Dependency{
		{Name: "postgres", Probe: storeProbe("postgres")},
		{Name: "redis", Probe: storeProbe("redis")},
		{Name: "minio", Probe: storeProbe("minio")},
		{Name: "clickhouse", Probe: func(ctx context.Context) (bool, error) {
			if ch == nil || !ch.Enabled() {
				return false, nil
			}
			return true, ch.Ping(ctx)
		}},
		{Name: "nats", Probe: func(context.Context) (bool, error) {
			if strings.TrimSpace(cfg.NATSURL) == "" {
				return false, nil
			}
			if bus == nil || !bus.Enabled() {
				return true, fmt.Errorf("nats unavailable")
			}
			return true, nil
		}},
		{Name: "livekit", Probe: func(context.Context) (bool, error) {
			if strings.TrimSpace(cfg.LiveKitAPIKey) == "" || strings.TrimSpace(cfg.LiveKitAPISecret) == "" {
				return false, nil
			}
			return true, nil
		}},
		{Name: "gemini", Probe: func(context.Context) (bool, error) {
			if ai == nil || !ai.Enabled() {
				return false, nil
			}
			return true, nil
		}},
	}
	return observability.New(dependencies, cfg.MonitoringProbeTimeout)
}

func (s *server) getTenantSystemPerformance(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.monitoring == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "monitoring unavailable", "code": "monitoring_unavailable"})
		return
	}
	snapshot := s.monitoring.Snapshot(r.Context(), s.analyticsHealthReader(tenantID))
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *server) analyticsHealthReader(tenantID string) observability.AnalyticsReader {
	return func(ctx context.Context) observability.Analytics {
		generated := time.Now().UTC()
		if s.ch == nil || !s.ch.Enabled() {
			return observability.Analytics{Status: observability.AnalyticsDisabled, GeneratedAt: &generated}
		}
		tz := "Asia/Bangkok"
		if s.store != nil {
			tz = s.store.TenantTimezone(ctx, tenantID)
		}
		loc, err := time.LoadLocation(strings.TrimSpace(tz))
		if err != nil {
			loc = time.UTC
		}
		today := time.Now().In(loc).Format("2006-01-02")
		stats, err := s.ch.QueryCallCenterStats(ctx, tenantID, today, today)
		if err != nil {
			return observability.Analytics{Status: observability.AnalyticsUnavailable, GeneratedAt: &generated}
		}
		analytics := observability.Analytics{Status: observability.AnalyticsCurrent, GeneratedAt: &generated}
		if !stats.Freshness.IsZero() {
			last := stats.Freshness.UTC()
			analytics.LastProjectedAt = &last
			if generated.Sub(last) > monitoringAnalyticsStaleAfter {
				analytics.Status = observability.AnalyticsStale
			}
		}
		return analytics
	}
}
