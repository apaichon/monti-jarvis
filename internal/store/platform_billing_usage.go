package store

import (
	"context"
	"fmt"
	"strings"
)

type PlatformBillingUsage struct {
	PaidOrders      int
	PaidAmountCents int64
	Currency        string
	ByTenant        map[string]PlatformTenantBillingUsage
}

type PlatformTenantBillingUsage struct {
	PaidOrders      int
	PaidAmountCents int64
	Currency        string
}

// GetPlatformBillingUsage reads paid orders for reporting. It does not infer
// payment from entitlements and never changes payment or fulfillment state.
func (s *Store) GetPlatformBillingUsage(ctx context.Context, tenantID, startDate, endDate string) (PlatformBillingUsage, error) {
	if s == nil || s.pg == nil {
		return PlatformBillingUsage{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	filter := ""
	args := []any{startDate, endDate}
	if strings.TrimSpace(tenantID) != "" {
		args = append(args, tenantID)
		filter = " AND o.tenant_id = $3"
	}
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT o.tenant_id, COUNT(*), COALESCE(SUM(o.amount_cents), 0), COALESCE(MAX(o.currency), '')
FROM %s.payment_orders o
JOIN %s.tenants t ON t.id = o.tenant_id AND t.status = 'active'
WHERE o.status = 'paid'
  AND o.created_at >= $1::date
  AND o.created_at < ($2::date + interval '1 day')%s
GROUP BY o.tenant_id`, schema, schema, filter), args...)
	if err != nil {
		return PlatformBillingUsage{}, err
	}
	defer rows.Close()
	out := PlatformBillingUsage{ByTenant: make(map[string]PlatformTenantBillingUsage)}
	for rows.Next() {
		var tenantID, currency string
		var orders int
		var amount int64
		if err := rows.Scan(&tenantID, &orders, &amount, &currency); err != nil {
			return PlatformBillingUsage{}, err
		}
		out.PaidOrders += orders
		out.PaidAmountCents += amount
		out.ByTenant[tenantID] = PlatformTenantBillingUsage{PaidOrders: orders, PaidAmountCents: amount, Currency: currency}
		if out.Currency == "" {
			out.Currency = currency
		} else if out.Currency != currency {
			out.Currency = "MIXED"
		}
	}
	return out, rows.Err()
}
