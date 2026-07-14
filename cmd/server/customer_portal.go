package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type customerPortalContext struct {
	TenantID string
	Settings store.CustomerAuthSettings
	Customer *store.Customer
	Auth     *auth.AuthContext
}

func (s *server) customerPortalPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := s.publicCustomerTenantID(r)
	settings, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	summary, _ := s.customerQuotaSummary(r, tenantID, settings, nil)
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id": tenantID,
		"customer_auth": map[string]any{
			"enabled":                    settings.Enabled,
			"mode":                       settings.AuthMode,
			"require_auth_for_workforce": settings.RequireAuthForWorkforce,
			"allow_public_no_auth":       !settings.RequireAuthForWorkforce,
		},
		"quota": summary,
	})
}

func (s *server) customerWorkforce(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.resolveCustomerPortalContext(w, r, true)
	if !ok {
		return
	}
	agents, err := s.customerWorkforceAgents(r, ctx.TenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		out = append(out, map[string]any{
			"id": agent.ID, "name": agent.Name, "role": agent.Role, "trait": agent.Trait,
			"color": agent.Color, "image": agent.Image, "greeting": agent.Greeting,
			"voice": agent.Voice, "popular": agent.Popular, "robot": agent.Robot,
			"skin": agent.Skin, "hair": agent.Hair, "quota_state": "available",
		})
	}
	selected := ""
	if len(agents) > 0 {
		selected = agents[0].ID
	}
	writeJSON(w, http.StatusOK, map[string]any{"avatars": out, "agents": out, "selected_avatar_id": selected})
}

func (s *server) customerQuota(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.resolveCustomerPortalContext(w, r, true)
	if !ok {
		return
	}
	summary, err := s.customerQuotaSummary(r, ctx.TenantID, ctx.Settings, ctx.Customer)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *server) resolveCustomerPortalContext(w http.ResponseWriter, r *http.Request, enforceWorkforceAuth bool) (customerPortalContext, bool) {
	tenantID := s.publicCustomerTenantID(r)
	settings, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return customerPortalContext{}, false
	}
	var ac *auth.AuthContext
	var customer *store.Customer
	if parsed, ok := auth.FromContext(r.Context()); ok && parsed.Role == auth.RoleCustomer && parsed.TenantID == tenantID {
		ac = &parsed
		if c, err := s.store.GetCustomer(r.Context(), parsed.TenantID, parsed.UserID); err == nil && c.Status == "active" {
			customer = c
		}
	}
	if enforceWorkforceAuth && settings.Enabled && settings.RequireAuthForWorkforce && customer == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "customer auth required", "code": "customer_auth_required"})
		return customerPortalContext{}, false
	}
	return customerPortalContext{TenantID: tenantID, Settings: settings, Customer: customer, Auth: ac}, true
}

func (s *server) customerQuotaSummary(r *http.Request, tenantID string, settings store.CustomerAuthSettings, customer *store.Customer) (store.CustomerUsageSummary, error) {
	customerID := ""
	if customer != nil {
		customerID = customer.ID
	}
	return s.store.CustomerUsageSummary(r.Context(), tenantID, customerID, settings.CustomerDailyCallSeconds, settings.CustomerMaxCallSeconds, time.Now())
}

func (s *server) enforceCustomerPortalAccess(w http.ResponseWriter, r *http.Request, tenantID, agentID, usageType string) (*store.Customer, store.CustomerAuthSettings, bool) {
	settings, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return nil, settings, false
	}
	var customer *store.Customer
	if ac, ok := auth.FromContext(r.Context()); ok && ac.Role == auth.RoleCustomer && ac.TenantID == tenantID {
		if c, err := s.store.GetCustomer(r.Context(), ac.TenantID, ac.UserID); err == nil && c.Status == "active" {
			customer = c
		}
	}
	if settings.Enabled && settings.RequireAuthForWorkforce && customer == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "customer auth required", "code": "customer_auth_required"})
		return nil, settings, false
	}
	if !s.agentAvailableForTenant(r, tenantID, agentID) {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "avatar unavailable", "code": "avatar_unavailable"})
		return nil, settings, false
	}
	if customer != nil && settings.CustomerDailyCallSeconds > 0 {
		summary, err := s.store.CustomerUsageSummary(r.Context(), tenantID, customer.ID, settings.CustomerDailyCallSeconds, settings.CustomerMaxCallSeconds, time.Now())
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return nil, settings, false
		}
		if summary.State == "quota_exhausted" {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "customer quota exhausted", "code": "customer_quota_exhausted", "quota": summary})
			_ = s.store.RecordCustomerUsage(r.Context(), tenantID, customer.ID, "", agentID, usageType, 0, "denied", "customer_quota_exhausted")
			return nil, settings, false
		}
	}
	return customer, settings, true
}

func (s *server) agentAvailableForTenant(r *http.Request, tenantID, agentID string) bool {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return true
	}
	agents, err := s.customerWorkforceAgents(r, tenantID)
	if err != nil {
		return false
	}
	for _, agent := range agents {
		if agent.ID == agentID {
			return true
		}
	}
	return false
}

func (s *server) customerWorkforceAgents(r *http.Request, tenantID string) ([]workforce.Agent, error) {
	if s.store == nil {
		return nil, errors.New("store is not available")
	}
	if strings.TrimSpace(tenantID) == "" {
		return nil, errors.New("tenant_id is required")
	}
	agents, err := s.store.ListWorkforceAgents(r.Context(), tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]workforce.Agent, 0, len(agents))
	for _, agent := range agents {
		out = append(out, workforce.FromWorkforceAgent(agent))
	}
	return out, nil
}
