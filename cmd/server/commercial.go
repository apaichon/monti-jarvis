package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/store"
)

type commercialCalculateBody struct {
	PackageID       string `json:"package_id"`
	BillingInterval string `json:"billing_interval"`
}

type dedicatedQuoteBody struct {
	PackageID           string `json:"package_id"`
	CompanyLegalName    string `json:"company_legal_name"`
	ContactName         string `json:"contact_name"`
	ContactEmail        string `json:"contact_email"`
	ContactPhone        string `json:"contact_phone"`
	TaxRegistrationID   string `json:"tax_registration_id"`
	CompanySize         string `json:"company_size"`
	ExpectedConcurrency int    `json:"expected_concurrency"`
	PreferredRegion     string `json:"preferred_region"`
	Notes               string `json:"notes"`
	IdempotencyKey      string `json:"idempotency_key"`
}

type quoteTransitionBody struct {
	Status            string         `json:"status"`
	QuotedAmountCents *int           `json:"quoted_amount_cents"`
	Currency          string         `json:"currency"`
	QuoteExpiresAt    *time.Time     `json:"quote_expires_at"`
	CapacitySnapshot  map[string]any `json:"capacity_snapshot"`
}

func (s *server) commercialTenantID(r *http.Request) (string, bool) {
	if tenantID, ok := s.tenantIDFromAuth(r); ok {
		return tenantID, true
	}
	if s.cfg.AuthDisabled {
		tenantID := auth.ResolveTenant(
			r.Context(),
			strings.TrimSpace(r.Header.Get("X-Tenant-ID")),
			true,
			s.cfg.DemoTenantID,
		)
		return tenantID, tenantID != ""
	}
	return "", false
}

func writeCommercialError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrPackageNotFound), errors.Is(err, store.ErrCatalogVersionNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "package not found", "code": "PACKAGE_NOT_FOUND"})
	case errors.Is(err, store.ErrPackageRequiresQuote):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "PACKAGE_REQUIRES_QUOTE"})
	case errors.Is(err, store.ErrQuoteNotFound), errors.Is(err, store.ErrBillingCycleNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error(), "code": "NOT_FOUND"})
	case errors.Is(err, store.ErrQuoteStateConflict):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "QUOTE_STATE_CONFLICT"})
	case errors.Is(err, store.ErrBillingCycleConflict):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "BILLING_CYCLE_CONFLICT"})
	case errors.Is(err, store.ErrIdempotencyConflict):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "IDEMPOTENCY_CONFLICT"})
	case errors.Is(err, store.ErrInvalidCommercialRequest):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "INVALID_COMMERCIAL_REQUEST"})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "commercial operation unavailable", "code": "COMMERCIAL_UNAVAILABLE"})
	}
}

// POST /api/tenant/commercial/calculate
func (s *server) calculateTenantCommercialPlan(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.commercialTenantID(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body commercialCalculateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "INVALID_COMMERCIAL_REQUEST"})
		return
	}
	calculation, err := s.store.CalculatePackagePrice(
		r.Context(),
		strings.TrimSpace(body.PackageID),
		strings.TrimSpace(body.BillingInterval),
		time.Now().UTC(),
	)
	if err != nil {
		writeCommercialError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, calculation)
}

// POST /api/tenant/commercial/quotes
func (s *server) createTenantDedicatedQuote(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.commercialTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body dedicatedQuoteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "INVALID_COMMERCIAL_REQUEST"})
		return
	}
	idempotencyKey := strings.TrimSpace(body.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	}
	quote, replayed, err := s.store.CreateDedicatedQuote(r.Context(), store.CreateDedicatedQuoteInput{
		TenantID: tenantID, PackageID: body.PackageID,
		CompanyLegalName: body.CompanyLegalName, ContactName: body.ContactName,
		ContactEmail: body.ContactEmail, ContactPhone: body.ContactPhone,
		TaxRegistrationID: body.TaxRegistrationID, CompanySize: body.CompanySize,
		ExpectedConcurrency: body.ExpectedConcurrency, PreferredRegion: body.PreferredRegion,
		Notes: body.Notes, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		writeCommercialError(w, err)
		return
	}
	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{"quote": quote, "replayed": replayed})
}

// GET /api/tenant/commercial/quotes
func (s *server) listTenantDedicatedQuotes(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.commercialTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := s.store.ListDedicatedQuotes(r.Context(), tenantID, strings.TrimSpace(r.URL.Query().Get("status")), 100)
	if err != nil {
		writeCommercialError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"quotes": items})
}

// GET /api/platform/commercial/quotes
func (s *server) listPlatformDedicatedQuotes(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.store.ListDedicatedQuotes(
		r.Context(),
		strings.TrimSpace(r.URL.Query().Get("tenant_id")),
		strings.TrimSpace(r.URL.Query().Get("status")),
		limit,
	)
	if err != nil {
		writeCommercialError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"quotes": items})
}

// PATCH /api/platform/commercial/quotes/{id}
func (s *server) transitionPlatformDedicatedQuote(w http.ResponseWriter, r *http.Request) {
	var body quoteTransitionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "INVALID_COMMERCIAL_REQUEST"})
		return
	}
	quote, err := s.store.TransitionDedicatedQuote(r.Context(), strings.TrimSpace(r.PathValue("id")), store.QuoteTransitionInput{
		Status: body.Status, QuotedAmountCents: body.QuotedAmountCents, Currency: body.Currency,
		QuoteExpiresAt: body.QuoteExpiresAt, CapacitySnapshot: body.CapacitySnapshot,
	})
	if err != nil {
		writeCommercialError(w, err)
		return
	}
	if quote.Status == store.QuoteActive && s.entitlements != nil {
		s.entitlements.Invalidate(r.Context(), quote.TenantID)
	}
	writeJSON(w, http.StatusOK, quote)
}

// POST /api/platform/billing/cycles/{id}/retry
func (s *server) retryPlatformBillingCycle(w http.ResponseWriter, r *http.Request) {
	cycle, err := s.store.RetryBillingCycle(
		r.Context(),
		strings.TrimSpace(r.PathValue("id")),
		time.Now().UTC(),
		s.cfg.BillingGracePeriod,
		s.cfg.BillingRetryDelays,
	)
	if err != nil {
		writeCommercialError(w, err)
		return
	}
	if s.entitlements != nil {
		s.entitlements.Invalidate(r.Context(), cycle.TenantID)
	}
	writeJSON(w, http.StatusOK, cycle)
}

type currentPlanQuotaDimension struct {
	Dimension   string   `json:"dimension"`
	Unit        string   `json:"unit"`
	Period      string   `json:"period"`
	Unlimited   bool     `json:"unlimited"`
	Limit       *int     `json:"limit"`
	Used        *int     `json:"used"`
	Remaining   *int     `json:"remaining"`
	Utilization *float64 `json:"utilization"`
	Source      string   `json:"source"`
	Freshness   string   `json:"freshness"`
}

func currentPlanQuotaDimensions(snapshot *quota.Snapshot) ([]currentPlanQuotaDimension, *float64) {
	if snapshot == nil {
		return []currentPlanQuotaDimension{}, nil
	}
	out := make([]currentPlanQuotaDimension, 0, len(snapshot.Dimensions))
	var compact *float64
	for _, dimension := range snapshot.Dimensions {
		limitValue := dimension.TotalLimit
		unlimited := commercialDimensionUnlimited(dimension.Dimension, limitValue)
		var limit *int
		var utilization *float64
		if !unlimited {
			value := limitValue
			limit = &value
			if dimension.Consumed != nil && limitValue > 0 && dimension.Freshness != "unavailable" {
				ratio := float64(*dimension.Consumed) / float64(limitValue)
				if ratio < 0 {
					ratio = 0
				}
				utilization = &ratio
				if compact == nil || ratio > *compact {
					copyValue := ratio
					compact = &copyValue
				}
			}
		}
		remaining := dimension.Remaining
		if unlimited {
			remaining = nil
		}
		out = append(out, currentPlanQuotaDimension{
			Dimension: dimension.Dimension, Unit: dimension.Unit, Period: dimension.Period,
			Unlimited: unlimited, Limit: limit, Used: dimension.Consumed, Remaining: remaining,
			Utilization: utilization, Source: dimension.Source, Freshness: dimension.Freshness,
		})
	}
	return out, compact
}

func commercialDimensionUnlimited(dimension string, limit int) bool {
	switch dimension {
	case "monthly_call_minutes", "mobile_call_minutes":
		return limit >= 99_999_999
	case "km_documents":
		return limit >= 1_000_000
	default:
		return false
	}
}

func unavailableQuotaDimensions(entitlement *store.TenantEntitlement, period string) []currentPlanQuotaDimension {
	if entitlement == nil {
		return []currentPlanQuotaDimension{}
	}
	if period == "" {
		period = time.Now().UTC().Format("2006-01")
	}
	type definition struct {
		rule string
		name string
		unit string
	}
	definitions := []definition{
		{"max_ai_employees", "ai_employees", "assignments"},
		{"max_monthly_call_minutes", "monthly_call_minutes", "minutes"},
		{"max_mobile_call_minutes", "mobile_call_minutes", "minutes"},
		{"max_km_documents", "km_documents", "documents"},
		{"max_storage_bytes", "storage_bytes", "bytes"},
		{"max_concurrent_calls", "concurrent_calls", "calls"},
	}
	out := make([]currentPlanQuotaDimension, 0, len(definitions))
	for _, definition := range definitions {
		value, ok := numericRuleValue(entitlement.RulesSnapshot[definition.rule])
		if !ok {
			continue
		}
		unlimited := commercialDimensionUnlimited(definition.name, value)
		var limit *int
		if !unlimited {
			copyValue := value
			limit = &copyValue
		}
		out = append(out, currentPlanQuotaDimension{
			Dimension: definition.name, Unit: definition.unit, Period: period,
			Unlimited: unlimited, Limit: limit, Source: "unavailable", Freshness: "unavailable",
		})
	}
	return out
}

func numericRuleValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		if typed > float64(math.MaxInt) || typed < 0 {
			return 0, false
		}
		return int(typed), true
	default:
		return 0, false
	}
}

// GET /api/tenant/commercial/current-plan
func (s *server) getTenantCurrentCommercialPlan(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.commercialTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	out := map[string]any{
		"tenant_id": tenantID, "billing_state": "no_plan", "package": nil,
		"subscription": nil, "next_bill": nil, "quota": []currentPlanQuotaDimension{},
		"compact_utilization": nil, "concurrent_queue": nil, "documents": []map[string]any{}, "quote": nil,
	}

	entitlement, entitlementErr := s.store.GetActiveEntitlement(r.Context(), tenantID)
	if entitlementErr != nil && !errors.Is(entitlementErr, store.ErrEntitlementNotFound) {
		writeCommercialError(w, entitlementErr)
		return
	}
	if entitlement != nil && entitlement.Package != nil {
		pkg := entitlement.Package
		out["package"] = map[string]any{
			"id": pkg.ID, "slug": pkg.Slug, "name": pkg.Name, "status": entitlement.Status,
			"deployment_mode": store.PackageDeployment(*pkg), "purchase_mode": store.PackagePurchaseMode(*pkg),
			"entitlement_id": entitlement.ID, "valid_from": entitlement.ValidFrom, "valid_until": entitlement.ValidUntil,
		}
		out["billing_state"] = "no_scheduled_bill"
		if subscription, err := s.store.GetLiveTenantSubscription(r.Context(), tenantID); err == nil {
			out["subscription"] = map[string]any{
				"id": subscription.ID, "status": subscription.Status, "billing_interval": subscription.BillingInterval,
				"current_period_start": subscription.CurrentPeriodStart,
				"current_period_end":   subscription.CurrentPeriodEnd,
				"billing_anchor":       subscription.BillingAnchor, "grace_until": subscription.GraceUntil,
			}
			state := "scheduled"
			if subscription.Status != store.SubscriptionActive {
				state = subscription.Status
			}
			out["billing_state"] = state
			if subscription.NextBillAt != nil {
				out["next_bill"] = map[string]any{
					"at":           subscription.NextBillAt,
					"amount_cents": subscription.CalculationSnapshot.AmountDueCents,
					"currency":     subscription.CalculationSnapshot.Currency,
					"state":        state,
				}
			}
		} else if !errors.Is(err, store.ErrSubscriptionNotFound) {
			writeCommercialError(w, err)
			return
		}

		if s.quota != nil {
			if snapshot, err := s.quota.Snapshot(r.Context(), tenantID); err == nil {
				dimensions, compact := currentPlanQuotaDimensions(snapshot)
				out["quota"] = dimensions
				out["compact_utilization"] = compact
				out["concurrent_queue"] = snapshot.ConcurrentQueue
			} else {
				out["quota"] = unavailableQuotaDimensions(entitlement, time.Now().UTC().Format("2006-01"))
			}
		} else {
			out["quota"] = unavailableQuotaDimensions(entitlement, time.Now().UTC().Format("2006-01"))
		}
	}

	quotes, quoteErr := s.store.ListDedicatedQuotes(r.Context(), tenantID, "", 20)
	if quoteErr == nil {
		for _, quote := range quotes {
			switch quote.Status {
			case store.QuoteSubmitted, store.QuoteUnderReview, store.QuoteCapacityConfirmed,
				store.QuoteQuoted, store.QuoteAccepted, store.QuoteProvisioning:
				out["quote"] = quote
				if entitlement == nil {
					out["billing_state"] = "quote_pending"
				}
				goto quotesDone
			}
		}
	}
quotesDone:

	if docs, err := s.store.ListTenantPaymentDocuments(r.Context(), tenantID); err == nil {
		documentRows := make([]map[string]any, 0, len(docs))
		for _, doc := range docs {
			if doc.Status != store.PaymentDocStatusIssued {
				continue
			}
			documentRows = append(documentRows, map[string]any{
				"id": doc.ID, "type": doc.DocType, "number": doc.DocNumber, "issued_at": doc.IssuedAt,
				"amount_cents": doc.AmountCents, "currency": doc.Currency,
				"href": "/api/tenant/billing/documents/" + doc.ID + "?format=html",
			})
			if len(documentRows) == 10 {
				break
			}
		}
		out["documents"] = documentRows
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) startBillingScheduler(ctx context.Context) {
	if !s.cfg.BillingSchedulerEnabled || s.store == nil {
		return
	}
	interval := s.cfg.BillingSchedulerPoll
	if interval <= 0 {
		interval = time.Minute
	}
	run := func() {
		cycles, err := s.store.ScheduleDueBillingCycles(
			ctx, time.Now().UTC(), 50, s.cfg.BillingGracePeriod, s.cfg.BillingRetryDelays,
		)
		if err != nil {
			log.Printf("billing scheduler: schedule due cycles: %v", err)
			return
		}
		retried, err := s.store.RetryDueBillingCycles(
			ctx, time.Now().UTC(), 50, s.cfg.BillingGracePeriod, s.cfg.BillingRetryDelays,
		)
		if err != nil {
			log.Printf("billing scheduler: retry due cycles: %v", err)
			return
		}
		settled, err := s.store.SettlePaidBillingCycles(ctx, 50)
		if err != nil {
			log.Printf("billing scheduler: settle paid cycles: %v", err)
			return
		}
		if len(cycles) > 0 || retried > 0 || settled > 0 {
			log.Printf("billing scheduler: scheduled=%d retried=%d settled=%d", len(cycles), retried, settled)
		}
	}
	go func() {
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
