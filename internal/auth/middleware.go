package auth

import (
	"context"
	"net/http"

	"github.com/libra/monti-jarvis/internal/auditctx"
)

type TenantStatusChecker interface {
	IsTenantActive(ctx context.Context, tenantID string) (bool, error)
}

type HTTPGuard struct {
	svc          *Service
	tenants      TenantStatusChecker
	authDisabled bool
}

func NewHTTPGuard(svc *Service, tenants TenantStatusChecker, authDisabled bool) *HTTPGuard {
	return &HTTPGuard{svc: svc, tenants: tenants, authDisabled: authDisabled}
}

func (g *HTTPGuard) OptionalBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.authDisabled || g.svc == nil || !g.svc.Enabled() {
			next.ServeHTTP(w, r)
			return
		}
		if ac, err := g.svc.ParseBearer(r.Header.Get("Authorization")); err == nil {
			ctx := WithContext(r.Context(), ac)
			r = r.WithContext(auditctx.WithActor(ctx, ac.UserID))
		}
		next.ServeHTTP(w, r)
	})
}

func (g *HTTPGuard) RequireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.authDisabled {
			next.ServeHTTP(w, r)
			return
		}
		if g.svc == nil || !g.svc.Enabled() {
			writeAuthError(w, ErrNotConfigured)
			return
		}
		ac, err := g.svc.ParseBearer(r.Header.Get("Authorization"))
		if err != nil {
			writeAuthError(w, ErrUnauthorized)
			return
		}
		ctx := auditctx.WithActor(WithContext(r.Context(), ac), ac.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *HTTPGuard) RequireKMWrite(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.authDisabled {
			next.ServeHTTP(w, r)
			return
		}
		if g.svc == nil || !g.svc.Enabled() {
			writeAuthError(w, ErrNotConfigured)
			return
		}
		ac, err := g.svc.ParseBearer(r.Header.Get("Authorization"))
		if err != nil {
			writeAuthError(w, ErrUnauthorized)
			return
		}
		if !CanWriteKM(ac.Role) {
			writeAuthError(w, ErrForbidden)
			return
		}
		if ac.Role == RoleTenantAdmin && g.tenants != nil && ac.TenantID != "" {
			active, err := g.tenants.IsTenantActive(r.Context(), ac.TenantID)
			if err != nil {
				writeAuthError(w, ErrNotConfigured)
				return
			}
			if !active {
				writeAuthError(w, ErrForbidden)
				return
			}
		}
		ctx := auditctx.WithActor(WithContext(r.Context(), ac), ac.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *HTTPGuard) RequireTenantAdminOrPlatform(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.authDisabled {
			next.ServeHTTP(w, r)
			return
		}
		if g.svc == nil || !g.svc.Enabled() {
			writeAuthError(w, ErrNotConfigured)
			return
		}
		ac, err := g.svc.ParseBearer(r.Header.Get("Authorization"))
		if err != nil {
			writeAuthError(w, ErrUnauthorized)
			return
		}
		if ac.Role != RolePlatformAdmin && ac.Role != RoleTenantAdmin {
			writeAuthError(w, ErrForbidden)
			return
		}
		ctx := auditctx.WithActor(WithContext(r.Context(), ac), ac.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *HTTPGuard) RequirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.authDisabled {
			next.ServeHTTP(w, r)
			return
		}
		if g.svc == nil || !g.svc.Enabled() {
			writeAuthError(w, ErrNotConfigured)
			return
		}
		ac, err := g.svc.ParseBearer(r.Header.Get("Authorization"))
		if err != nil {
			writeAuthError(w, ErrUnauthorized)
			return
		}
		if ac.Role != RolePlatformAdmin {
			writeAuthError(w, ErrForbidden)
			return
		}
		ctx := auditctx.WithActor(WithContext(r.Context(), ac), ac.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusUnauthorized
	msg := "unauthorized"
	switch err {
	case ErrForbidden:
		status = http.StatusForbidden
		msg = "forbidden"
	case ErrNotConfigured:
		status = http.StatusServiceUnavailable
		msg = "auth is not configured"
	case ErrInvalidCredentials:
		msg = "invalid credentials"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}