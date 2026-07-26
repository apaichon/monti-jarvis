package km

import (
	"context"
	"strings"
)

type tenantContextKey struct{}

// WithTenantContext binds the tenant used by database RLS for a KM request.
// The store layer consumes this without depending on HTTP/auth packages.
func WithTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, strings.TrimSpace(tenantID))
}

func TenantFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	tenantID, _ := ctx.Value(tenantContextKey{}).(string)
	return strings.TrimSpace(tenantID)
}
