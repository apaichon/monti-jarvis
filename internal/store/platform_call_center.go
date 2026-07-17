package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// PlatformCallCenterSatisfaction is a redacted, tenant-level rating summary.
// It deliberately excludes review text and all customer dimensions.
type PlatformCallCenterSatisfaction struct {
	TenantID     string
	Reviewed     int
	AverageScore float64
	Distribution map[string]int
}

// CountActivePlatformEntitlements counts active packages for active tenants.
// It is a reporting value only; this method does not read or mutate quota
// counters.
func (s *Store) CountActivePlatformEntitlements(ctx context.Context, tenantID string) (int, error) {
	if s == nil || s.pg == nil {
		return 0, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	filter := ""
	args := []any{}
	if strings.TrimSpace(tenantID) != "" {
		filter = " AND e.tenant_id = $1"
		args = append(args, tenantID)
	}
	var count int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*)
FROM %s.tenant_entitlements e
JOIN %s.tenants t ON t.id = e.tenant_id
WHERE e.status = 'active' AND t.status = 'active'%s`, schema, schema, filter), args...).Scan(&count)
	return count, err
}

// GetPlatformCallCenterSatisfaction returns bounded aggregate enrichment for
// the requested range. The query groups by tenant and score only; it never
// selects customer identifiers or review content.
func (s *Store) GetPlatformCallCenterSatisfaction(ctx context.Context, tenantID, startDate, endDate string) (map[string]PlatformCallCenterSatisfaction, error) {
	if s == nil || s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	filter := ""
	args := []any{startDate, endDate}
	if strings.TrimSpace(tenantID) != "" {
		args = append(args, tenantID)
		filter = " AND r.tenant_id = $3"
	}
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT r.tenant_id, cr.score, COUNT(*)
FROM %s.conversation_records r
LEFT JOIN %s.conversation_ratings cr
  ON cr.tenant_id = r.tenant_id AND cr.call_id = r.call_id
WHERE r.status = 'archived'
  AND r.ended_at >= $1::date
  AND r.ended_at < ($2::date + interval '1 day')%s
GROUP BY r.tenant_id, cr.score`, schema, schema, filter), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]PlatformCallCenterSatisfaction)
	for rows.Next() {
		var id string
		var score *int
		var count int
		if err := rows.Scan(&id, &score, &count); err != nil {
			return nil, err
		}
		stats := out[id]
		if stats.Distribution == nil {
			stats.Distribution = map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
		}
		if score != nil && *score >= 1 && *score <= 5 {
			stats.Reviewed += count
			stats.AverageScore += float64(*score * count)
			stats.Distribution[strconv.Itoa(*score)] += count
		}
		out[id] = stats
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for id, stats := range out {
		if stats.Reviewed > 0 {
			stats.AverageScore = roundSatisfaction(stats.AverageScore / float64(stats.Reviewed))
		}
		// Completed count is not returned by this method, so the API computes
		// completion rate against the ClickHouse tenant activity count.
		out[id] = stats
	}
	return out, nil
}
