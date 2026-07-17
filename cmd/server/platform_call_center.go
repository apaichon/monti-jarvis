package main

import (
	"context"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/store"
)

const (
	defaultPlatformCallCenterLimit = 50
	maxPlatformCallCenterLimit     = 100
	platformCallCenterTimeout      = 10 * time.Second
	platformCallCenterStaleAfter   = 5 * time.Minute
)

type platformCallCenterResponse struct {
	Range        platformCallCenterRange        `json:"range"`
	Freshness    platformCallCenterFreshness    `json:"freshness"`
	Totals       platformCallCenterTotals       `json:"totals"`
	ByChannel    []platformCallCenterDimension  `json:"by_channel"`
	ByAvatar     []platformCallCenterDimension  `json:"by_avatar"`
	PackageUsage platformCallCenterPackageUsage `json:"package_usage"`
	Enrichment   platformCallCenterEnrichment   `json:"enrichment"`
	Tenants      []platformCallCenterTenant     `json:"tenants"`
	Pagination   platformCallCenterPagination   `json:"pagination"`
}

type platformCallCenterRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Timezone  string `json:"timezone"`
}

type platformCallCenterFreshness struct {
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	GeneratedAt     time.Time  `json:"generated_at"`
	LastProjectedAt *time.Time `json:"last_projected_at,omitempty"`
}

type platformCallCenterTotals struct {
	CompletedConversations   int            `json:"completed_conversations"`
	TotalDurationSeconds     int            `json:"total_duration_seconds"`
	AverageDurationSeconds   float64        `json:"average_duration_seconds"`
	ChatConversations        int            `json:"chat_conversations"`
	VoiceConversations       int            `json:"voice_conversations"`
	RangeCallMinutes         int            `json:"range_call_minutes"`
	ReviewedConversations    int            `json:"reviewed_conversations"`
	AverageSatisfaction      float64        `json:"average_satisfaction"`
	ReviewCompletionRate     float64        `json:"review_completion_rate"`
	SatisfactionDistribution map[string]int `json:"satisfaction_distribution"`
}

type platformCallCenterDimension struct {
	ID                     string  `json:"id,omitempty"`
	Name                   string  `json:"name,omitempty"`
	Channel                string  `json:"channel,omitempty"`
	Completed              int     `json:"completed"`
	TotalDurationSeconds   int     `json:"total_duration_seconds"`
	AverageDurationSeconds float64 `json:"average_duration_seconds"`
}

type platformCallCenterPackageUsage struct {
	ActivePackageTenants int    `json:"active_package_tenants"`
	RangeCallMinutes     int    `json:"range_call_minutes"`
	EnforcementCounters  string `json:"enforcement_counters"`
}

type platformCallCenterEnrichment struct {
	Satisfaction string `json:"satisfaction"`
	Packages     string `json:"packages"`
}

type platformCallCenterTenant struct {
	TenantID               string                    `json:"tenant_id"`
	Slug                   string                    `json:"slug"`
	Name                   string                    `json:"name"`
	LogoURL                string                    `json:"logo_url,omitempty"`
	Package                platformCallCenterPackage `json:"package"`
	AnalyticsStatus        string                    `json:"analytics_status"`
	CompletedConversations int                       `json:"completed_conversations"`
	TotalDurationSeconds   int                       `json:"total_duration_seconds"`
	AverageDurationSeconds float64                   `json:"average_duration_seconds"`
	RangeCallMinutes       int                       `json:"range_call_minutes"`
	ReviewedConversations  int                       `json:"reviewed_conversations"`
	AverageSatisfaction    float64                   `json:"average_satisfaction"`
	ReviewCompletionRate   float64                   `json:"review_completion_rate"`
}

type platformCallCenterPackage struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type platformCallCenterPagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (s *server) getPlatformCallCenterStatistics(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, tenantFilter, limit, offset, err := parsePlatformCallCenterQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "validation_error"})
		return
	}
	if s.ch == nil || !s.ch.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "analytics unavailable", "code": "analytics_unavailable"})
		return
	}
	if s.store == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "statistics unavailable", "code": "statistics_unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), platformCallCenterTimeout)
	defer cancel()
	stats, err := s.ch.QueryPlatformCallCenterStats(ctx, tenantFilter, startDate, endDate)
	if err != nil {
		// Keep provider details in server logs only. The client receives a
		// stable state that is distinguishable from an empty range.
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "analytics unavailable", "code": "analytics_unavailable"})
		return
	}

	items, err := s.listActiveMonitoringTenants(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "statistics unavailable", "code": "statistics_unavailable"})
		return
	}
	if tenantFilter != "" {
		filtered := items[:0]
		for _, item := range items {
			if item.ID == tenantFilter {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	totalTenants := len(items)
	page := items
	if offset >= len(page) {
		page = nil
	} else {
		end := offset + limit
		if end > len(page) {
			end = len(page)
		}
		page = page[offset:end]
	}

	satisfaction, satisfactionErr := s.store.GetPlatformCallCenterSatisfaction(ctx, tenantFilter, startDate, endDate)
	packageCount, packageCountErr := s.store.CountActivePlatformEntitlements(ctx, tenantFilter)
	avatarNames := platformAvatarNames(ctx, s.store)

	response := platformCallCenterResponse{
		Range: platformCallCenterRange{StartDate: startDate, EndDate: endDate, Timezone: platformStatisticsTimezone()},
		Freshness: platformCallCenterFreshness{
			Source:      "clickhouse",
			Status:      platformCallCenterFreshnessStatus(stats.LastProjectedAt),
			GeneratedAt: time.Now().UTC(),
		},
		Totals: platformCallCenterTotals{
			CompletedConversations:   stats.CompletedConversations,
			TotalDurationSeconds:     stats.TotalDurationSeconds,
			AverageDurationSeconds:   averageSecondsFloat(stats.TotalDurationSeconds, stats.CompletedConversations),
			RangeCallMinutes:         rangeCallMinutes(stats.TotalDurationSeconds),
			SatisfactionDistribution: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
		},
		ByChannel: make([]platformCallCenterDimension, 0, len(stats.ByChannel)),
		ByAvatar:  make([]platformCallCenterDimension, 0, len(stats.ByAvatar)),
		PackageUsage: platformCallCenterPackageUsage{
			RangeCallMinutes:     rangeCallMinutes(stats.TotalDurationSeconds),
			EnforcementCounters:  "not_included",
			ActivePackageTenants: packageCount,
		},
		Enrichment: platformCallCenterEnrichment{Satisfaction: "available", Packages: "available"},
		Tenants:    make([]platformCallCenterTenant, 0, len(page)),
		Pagination: platformCallCenterPagination{Total: totalTenants, Limit: limit, Offset: offset},
	}
	if !stats.LastProjectedAt.IsZero() {
		last := stats.LastProjectedAt.UTC()
		response.Freshness.LastProjectedAt = &last
	}
	if satisfactionErr != nil {
		response.Enrichment.Satisfaction = "unavailable"
	}
	if packageCountErr != nil {
		response.Enrichment.Packages = "unavailable"
		response.PackageUsage.ActivePackageTenants = 0
	}

	for _, bucket := range stats.ByChannel {
		response.ByChannel = append(response.ByChannel, platformCallCenterDimension{
			Channel: bucket.Channel, Completed: bucket.CompletedConversations,
			TotalDurationSeconds: bucket.TotalDurationSeconds, AverageDurationSeconds: bucket.AverageDurationSeconds,
		})
		if bucket.Channel == "chat" {
			response.Totals.ChatConversations = bucket.CompletedConversations
		}
		if bucket.Channel == "voice" {
			response.Totals.VoiceConversations = bucket.CompletedConversations
		}
	}
	for _, bucket := range stats.ByAvatar {
		name := avatarNames[bucket.AvatarID]
		if name == "" {
			name = bucket.AvatarID
		}
		response.ByAvatar = append(response.ByAvatar, platformCallCenterDimension{
			ID: bucket.AvatarID, Name: name, Completed: bucket.CompletedConversations,
			TotalDurationSeconds: bucket.TotalDurationSeconds, AverageDurationSeconds: bucket.AverageDurationSeconds,
		})
	}
	for _, rating := range satisfaction {
		response.Totals.ReviewedConversations += rating.Reviewed
		response.Totals.AverageSatisfaction += rating.AverageScore * float64(rating.Reviewed)
		for score, count := range rating.Distribution {
			response.Totals.SatisfactionDistribution[score] += count
		}
	}
	if response.Totals.ReviewedConversations > 0 {
		response.Totals.AverageSatisfaction = roundPlatformSatisfaction(response.Totals.AverageSatisfaction / float64(response.Totals.ReviewedConversations))
	}
	if stats.CompletedConversations > 0 {
		response.Totals.ReviewCompletionRate = roundPlatformSatisfaction(float64(response.Totals.ReviewedConversations) * 100 / float64(stats.CompletedConversations))
	}
	for _, item := range page {
		bucket := platformTenantBucket(stats.ByTenant, item.ID)
		rating := satisfaction[item.ID]
		row := platformCallCenterTenant{
			TenantID: item.ID, Slug: item.Slug, Name: item.Name, LogoURL: item.LogoURL,
			Package:                platformCallCenterPackage{Name: "Unassigned", Status: "unassigned"},
			AnalyticsStatus:        platformCallCenterTenantStatus(bucket, stats.LastProjectedAt),
			CompletedConversations: bucket.CompletedConversations,
			TotalDurationSeconds:   bucket.TotalDurationSeconds,
			AverageDurationSeconds: bucket.AverageDurationSeconds,
			RangeCallMinutes:       rangeCallMinutes(bucket.TotalDurationSeconds),
			ReviewedConversations:  rating.Reviewed, AverageSatisfaction: rating.AverageScore,
		}
		if bucket.CompletedConversations > 0 {
			row.ReviewCompletionRate = roundPlatformSatisfaction(float64(rating.Reviewed) * 100 / float64(bucket.CompletedConversations))
		}
		if entitlement, entitlementErr := s.store.GetActiveEntitlement(ctx, item.ID); entitlementErr == nil && entitlement != nil && entitlement.Package != nil {
			row.Package = platformCallCenterPackage{Name: entitlement.Package.Name, Status: entitlement.Status}
		} else if entitlementErr != nil && entitlementErr != store.ErrEntitlementNotFound {
			response.Enrichment.Packages = "unavailable"
		}
		response.Tenants = append(response.Tenants, row)
	}
	writeJSON(w, http.StatusOK, response)
}

func parsePlatformCallCenterQuery(r *http.Request) (startDate, endDate, tenantID string, limit, offset int, err error) {
	query := r.URL.Query()
	today := time.Now().In(platformStatisticsLocation()).Format("2006-01-02")
	startDate = strings.TrimSpace(query.Get("start_date"))
	endDate = strings.TrimSpace(query.Get("end_date"))
	if startDate == "" {
		startDate = today
	}
	if endDate == "" {
		endDate = today
	}
	start, startErr := time.Parse("2006-01-02", startDate)
	end, endErr := time.Parse("2006-01-02", endDate)
	if startErr != nil || endErr != nil {
		return "", "", "", 0, 0, &platformCallCenterQueryError{"start_date and end_date must be YYYY-MM-DD"}
	}
	if start.After(end) {
		return "", "", "", 0, 0, &platformCallCenterQueryError{"start_date must be on or before end_date"}
	}
	tenantID = strings.TrimSpace(query.Get("tenant_id"))
	if len(tenantID) > 128 {
		return "", "", "", 0, 0, &platformCallCenterQueryError{"tenant_id must be at most 128 characters"}
	}
	limit = defaultPlatformCallCenterLimit
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > maxPlatformCallCenterLimit {
			return "", "", "", 0, 0, &platformCallCenterQueryError{"limit must be between 1 and 100"}
		}
	}
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 || offset > 1_000_000 {
			return "", "", "", 0, 0, &platformCallCenterQueryError{"offset must be a non-negative bounded integer"}
		}
	}
	return startDate, endDate, tenantID, limit, offset, nil
}

type platformCallCenterQueryError struct{ message string }

func (e *platformCallCenterQueryError) Error() string { return e.message }

func platformStatisticsTimezone() string { return platformStatisticsLocation().String() }

func platformStatisticsLocation() *time.Location {
	if tz := strings.TrimSpace(os.Getenv("TZ")); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	if time.Local != nil && time.Local.String() != "Local" {
		return time.Local
	}
	if loc, err := time.LoadLocation("Asia/Bangkok"); err == nil {
		return loc
	}
	return time.UTC
}

func platformCallCenterFreshnessStatus(last time.Time) string {
	if last.IsZero() {
		return "empty"
	}
	if time.Since(last) > platformCallCenterStaleAfter {
		return "stale"
	}
	return "current"
}

func platformCallCenterTenantStatus(bucket clickhouse.PlatformCallCenterBucket, last time.Time) string {
	if bucket.CompletedConversations == 0 {
		if last.IsZero() {
			return "empty"
		}
		return "current"
	}
	if last.IsZero() || time.Since(last) > platformCallCenterStaleAfter {
		return "stale"
	}
	return "current"
}

func platformTenantBucket(buckets []clickhouse.PlatformCallCenterBucket, tenantID string) clickhouse.PlatformCallCenterBucket {
	for _, bucket := range buckets {
		if bucket.TenantID == tenantID {
			return bucket
		}
	}
	return clickhouse.PlatformCallCenterBucket{TenantID: tenantID}
}

func platformAvatarNames(ctx context.Context, st *store.Store) map[string]string {
	result := map[string]string{}
	if st == nil {
		return result
	}
	avatars, err := st.ListAvatars(ctx, "active")
	if err != nil {
		return result
	}
	for _, avatar := range avatars {
		result[avatar.ID] = avatar.Name
	}
	return result
}

func averageSecondsFloat(total, count int) float64 {
	if count <= 0 {
		return 0
	}
	return float64(total) / float64(count)
}

func rangeCallMinutes(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return int(math.Ceil(float64(seconds) / 60))
}

func roundPlatformSatisfaction(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
