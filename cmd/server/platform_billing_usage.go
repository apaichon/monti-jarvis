package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/store"
)

const (
	defaultPlatformBillingUsageLimit = 50
	maxPlatformBillingUsageLimit     = 100
	platformBillingUsageTimeout      = 10 * time.Second
)

type platformBillingUsageResponse struct {
	Range          platformBillingUsageRange     `json:"range"`
	Freshness      platformBillingUsageFreshness `json:"freshness"`
	Billing        platformBillingSummary        `json:"billing"`
	Quota          platformBillingQuotaSummary   `json:"quota"`
	AICost         platformAICostSummary         `json:"ai_cost"`
	Reconciliation platformBillingReconciliation `json:"reconciliation"`
	Tenants        []platformBillingTenant       `json:"tenants"`
	Pagination     platformBillingPagination     `json:"pagination"`
}

type platformBillingUsageRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Timezone  string `json:"timezone"`
}

type platformBillingUsageFreshness struct {
	Status                string     `json:"status"`
	GeneratedAt           time.Time  `json:"generated_at"`
	ActivityLastProjected *time.Time `json:"activity_last_projected_at,omitempty"`
	AIUsageLastProjected  *time.Time `json:"ai_usage_last_projected_at,omitempty"`
}

type platformBillingSummary struct {
	PaidOrders      int    `json:"paid_orders"`
	PaidAmountMinor int64  `json:"paid_amount_minor"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
}

type platformBillingQuotaSummary struct {
	ReportingMinutes int                      `json:"reporting_minutes"`
	Enforcement      platformQuotaEnforcement `json:"enforcement"`
}

type platformQuotaEnforcement struct {
	Status       string `json:"status"`
	MonthlyUsed  int    `json:"monthly_used"`
	MonthlyLimit int    `json:"monthly_limit"`
	DailyUsed    int    `json:"daily_used"`
	DailyLimit   int    `json:"daily_limit"`
	DailyStatus  string `json:"daily_status"`
}

type platformAICostSummary struct {
	RateVersion         string  `json:"rate_version"`
	PricingAsOf         string  `json:"pricing_as_of,omitempty"`
	Currency            string  `json:"currency"`
	ObservedCostMicros  int64   `json:"observed_cost_microunits"`
	EstimatedCostMicros int64   `json:"estimated_cost_microunits"`
	ObservedEvents      int     `json:"observed_events"`
	EstimatedEvents     int     `json:"estimated_events"`
	UnavailableEvents   int     `json:"unavailable_events"`
	CoveragePercent     float64 `json:"coverage_percent"`
	Status              string  `json:"status"`
}

type platformBillingReconciliation struct {
	ActivityQuota      string `json:"activity_quota"`
	OrdersEntitlements string `json:"orders_entitlements"`
	AICoverage         string `json:"ai_coverage"`
}

type platformBillingTenant struct {
	TenantID              string                   `json:"tenant_id"`
	Slug                  string                   `json:"slug"`
	Name                  string                   `json:"name"`
	Package               platformBillingPackage   `json:"package"`
	PaidOrders            int                      `json:"paid_orders"`
	PaidAmountMinor       int64                    `json:"paid_amount_minor"`
	Currency              string                   `json:"currency"`
	ReportingMinutes      int                      `json:"reporting_minutes"`
	Quota                 platformQuotaEnforcement `json:"quota"`
	AIObservedCostMicros  int64                    `json:"ai_observed_cost_microunits"`
	AIEstimatedCostMicros int64                    `json:"ai_estimated_cost_microunits"`
	AIObservedEvents      int                      `json:"ai_observed_events"`
	AIEstimatedEvents     int                      `json:"ai_estimated_events"`
	AIUnavailableEvents   int                      `json:"ai_unavailable_events"`
	AICoveragePercent     float64                  `json:"ai_coverage_percent"`
	Status                string                   `json:"status"`
}

type platformBillingPackage struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type platformBillingPagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (s *server) getPlatformBillingUsage(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, tenantFilter, limit, offset, err := parsePlatformBillingUsageQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "validation_error"})
		return
	}
	if s.store == nil || s.ch == nil || !s.ch.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "usage unavailable", "code": "usage_unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformBillingUsageTimeout)
	defer cancel()

	activity, err := s.ch.QueryPlatformCallCenterStats(ctx, tenantFilter, startDate, endDate)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "usage unavailable", "code": "usage_unavailable"})
		return
	}
	aiStats, aiErr := s.ch.QueryAIUsageStats(ctx, tenantFilter, startDate, endDate)
	billing, billingErr := s.store.GetPlatformBillingUsage(ctx, tenantFilter, startDate, endDate)
	if billingErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "usage unavailable", "code": "usage_unavailable"})
		return
	}
	items, total, err := s.listBillingUsageTenants(ctx, tenantFilter, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "usage unavailable", "code": "usage_unavailable"})
		return
	}

	response := platformBillingUsageResponse{
		Range:          platformBillingUsageRange{StartDate: startDate, EndDate: endDate, Timezone: platformStatisticsTimezone()},
		Freshness:      platformBillingUsageFreshness{Status: platformCallCenterFreshnessStatus(activity.LastProjectedAt), GeneratedAt: time.Now().UTC()},
		Billing:        platformBillingSummary{PaidOrders: billing.PaidOrders, PaidAmountMinor: billing.PaidAmountCents, Currency: billing.Currency, Status: "current"},
		Quota:          platformBillingQuotaSummary{ReportingMinutes: rangeCallMinutes(activity.TotalDurationSeconds), Enforcement: platformQuotaEnforcement{Status: "unavailable"}},
		AICost:         platformAICostSummary{RateVersion: s.cfg.AIUsageRateVersion, PricingAsOf: s.cfg.AIUsagePricingAsOf, Currency: s.cfg.AIUsageCurrency, Status: "unavailable"},
		Reconciliation: platformBillingReconciliation{ActivityQuota: "not_comparable", OrdersEntitlements: "ok", AICoverage: "unavailable"},
		Tenants:        make([]platformBillingTenant, 0, len(items)),
		Pagination:     platformBillingPagination{Total: total, Limit: limit, Offset: offset},
	}
	if !activity.LastProjectedAt.IsZero() {
		last := activity.LastProjectedAt.UTC()
		response.Freshness.ActivityLastProjected = &last
	}
	if aiErr == nil {
		response.AICost = platformAICostResponse(aiStats, s.cfg.AIUsageRateVersion, s.cfg.AIUsagePricingAsOf, s.cfg.AIUsageCurrency)
		if !aiStats.LastProjectedAt.IsZero() {
			last := aiStats.LastProjectedAt.UTC()
			response.Freshness.AIUsageLastProjected = &last
		}
		response.Reconciliation.AICoverage = aiCoverageStatus(aiStats)
	} else {
		response.Reconciliation.AICoverage = "unavailable"
	}

	for _, item := range items {
		row := platformBillingTenant{TenantID: item.ID, Slug: item.Slug, Name: item.Name, Status: "current"}
		if paid, ok := billing.ByTenant[item.ID]; ok {
			row.PaidOrders, row.PaidAmountMinor, row.Currency = paid.PaidOrders, paid.PaidAmountCents, paid.Currency
		}
		activityBucket := platformTenantBucket(activity.ByTenant, item.ID)
		row.ReportingMinutes = rangeCallMinutes(activityBucket.TotalDurationSeconds)
		if aiErr == nil {
			if bucket := platformAITenant(aiStats.ByTenant, item.ID); bucket != nil {
				row.AIObservedCostMicros = bucket.ObservedCostMicros
				row.AIEstimatedCostMicros = bucket.EstimatedCostMicros
				row.AIObservedEvents = bucket.ObservedEvents
				row.AIEstimatedEvents = bucket.EstimatedEvents
				row.AIUnavailableEvents = bucket.UnavailableEvents
				row.AICoveragePercent = aiBucketCoverage(*bucket)
			}
		}
		row.Package = platformBillingPackage{Name: "Unassigned", Status: "unassigned"}
		if s.quota != nil {
			if snap, snapErr := s.quota.Snapshot(ctx, item.ID); snapErr == nil {
				row.Quota = quotaUsageResponse(snap)
				if snap.Package != nil {
					row.Package = platformBillingPackage{Name: snap.Package.Name, Status: snap.Status}
				}
			} else {
				row.Quota.Status = "unavailable"
				row.Status = "degraded"
			}
		} else {
			row.Quota.Status = "disabled"
		}
		response.Tenants = append(response.Tenants, row)
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *server) listBillingUsageTenants(ctx context.Context, tenantFilter string, limit, offset int) ([]store.TenantListItem, int, error) {
	if tenantFilter != "" {
		tenant, err := s.store.GetTenant(ctx, tenantFilter)
		if err != nil {
			return nil, 0, err
		}
		if tenant.Status != "active" {
			return []store.TenantListItem{}, 0, nil
		}
		return []store.TenantListItem{{ID: tenant.ID, Slug: tenant.Slug, Name: tenant.Name, Status: tenant.Status}}, 1, nil
	}
	return s.store.ListTenants(ctx, "active", "", limit, offset)
}

func quotaUsageResponse(snapshot *quota.Snapshot) platformQuotaEnforcement {
	if snapshot == nil {
		return platformQuotaEnforcement{Status: "unavailable"}
	}
	status := "current"
	if snapshot.Status != "active" || snapshot.Limits == nil {
		status = "unavailable"
	}
	monthlyLimit := 0
	if snapshot.Limits != nil {
		monthlyLimit = snapshot.Limits.MaxMonthlyCallMinutes
	}
	return platformQuotaEnforcement{Status: status, MonthlyUsed: snapshot.Usage.MonthlyCallMinutes, MonthlyLimit: monthlyLimit, DailyStatus: "unavailable"}
}

func parsePlatformBillingUsageQuery(r *http.Request) (startDate, endDate, tenantID string, limit, offset int, err error) {
	query := r.URL.Query()
	today := time.Now().In(platformStatisticsLocation()).Format("2006-01-02")
	startDate, endDate = strings.TrimSpace(query.Get("start_date")), strings.TrimSpace(query.Get("end_date"))
	if startDate == "" {
		startDate = today
	}
	if endDate == "" {
		endDate = today
	}
	start, startErr := time.Parse("2006-01-02", startDate)
	end, endErr := time.Parse("2006-01-02", endDate)
	if startErr != nil || endErr != nil {
		return "", "", "", 0, 0, errors.New("start_date and end_date must be YYYY-MM-DD")
	}
	if start.After(end) || end.Sub(start) > 366*24*time.Hour {
		return "", "", "", 0, 0, errors.New("date range must be 366 days or less and start_date must be on or before end_date")
	}
	tenantID = strings.TrimSpace(query.Get("tenant_id"))
	if len(tenantID) > 128 {
		return "", "", "", 0, 0, errors.New("tenant_id must be at most 128 characters")
	}
	limit = defaultPlatformBillingUsageLimit
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > maxPlatformBillingUsageLimit {
			return "", "", "", 0, 0, errors.New("limit must be between 1 and 100")
		}
	}
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 || offset > 1_000_000 {
			return "", "", "", 0, 0, errors.New("offset must be a non-negative bounded integer")
		}
	}
	return startDate, endDate, tenantID, limit, offset, nil
}

func platformAICostResponse(stats clickhouse.AIUsageStats, rateVersion, pricingAsOf, currency string) platformAICostSummary {
	status := aiCoverageStatus(stats)
	return platformAICostSummary{RateVersion: rateVersion, PricingAsOf: pricingAsOf, Currency: currency, ObservedCostMicros: stats.ObservedCostMicros, EstimatedCostMicros: stats.EstimatedCostMicros, ObservedEvents: stats.ObservedEvents, EstimatedEvents: stats.EstimatedEvents, UnavailableEvents: stats.UnavailableEvents, CoveragePercent: aiCoveragePercent(stats), Status: status}
}

func aiCoverageStatus(stats clickhouse.AIUsageStats) string {
	if stats.Events == 0 {
		return "empty"
	}
	if stats.UnavailableEvents > 0 || stats.EstimatedEvents > 0 {
		return "warning"
	}
	return "current"
}

func aiCoveragePercent(stats clickhouse.AIUsageStats) float64 {
	if stats.Events == 0 {
		return 0
	}
	return float64(stats.ObservedEvents+stats.EstimatedEvents) * 100 / float64(stats.Events)
}

func aiBucketCoverage(bucket clickhouse.AIUsageBucket) float64 {
	if bucket.Events == 0 {
		return 0
	}
	return float64(bucket.ObservedEvents+bucket.EstimatedEvents) * 100 / float64(bucket.Events)
}

func platformAITenant(rows []clickhouse.AIUsageBucket, tenantID string) *clickhouse.AIUsageBucket {
	for i := range rows {
		if rows[i].TenantID == tenantID {
			return &rows[i]
		}
	}
	return nil
}
