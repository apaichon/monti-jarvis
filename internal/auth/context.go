package auth

import (
	"context"
	"strings"
)

type ctxKey struct{}

type AuthContext struct {
	UserID   string
	Email    string
	Role     Role
	TenantID string
	JTI      string
}

func WithContext(ctx context.Context, ac AuthContext) context.Context {
	return context.WithValue(ctx, ctxKey{}, ac)
}

func FromContext(ctx context.Context) (AuthContext, bool) {
	ac, ok := ctx.Value(ctxKey{}).(AuthContext)
	return ac, ok
}

// ResolveTenant picks the effective tenant for KM/calls data access.
func ResolveTenant(ctx context.Context, headerTenant string, authDisabled bool, demoTenantID string) string {
	headerTenant = strings.TrimSpace(headerTenant)
	demoTenantID = strings.TrimSpace(demoTenantID)
	if demoTenantID == "" {
		demoTenantID = "demo"
	}

	if authDisabled {
		if headerTenant != "" {
			return headerTenant
		}
		return demoTenantID
	}

	ac, ok := FromContext(ctx)
	if !ok {
		if headerTenant != "" {
			return headerTenant
		}
		return demoTenantID
	}

	switch ac.Role {
	case RolePlatformAdmin:
		if headerTenant != "" {
			return headerTenant
		}
		if ac.TenantID != "" {
			return ac.TenantID
		}
		return demoTenantID
	default:
		if ac.TenantID != "" {
			return ac.TenantID
		}
		return demoTenantID
	}
}