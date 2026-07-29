package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrCatalogVersionNotFound   = errors.New("catalog version not found")
	ErrInvalidCommercialRequest = errors.New("invalid commercial request")
	ErrPackageRequiresQuote     = errors.New("package requires quotation")
	ErrQuoteNotFound            = errors.New("dedicated quote request not found")
	ErrQuoteStateConflict       = errors.New("dedicated quote state conflict")
	ErrSubscriptionNotFound     = errors.New("tenant subscription not found")
	ErrBillingCycleNotFound     = errors.New("billing cycle not found")
	ErrBillingCycleConflict     = errors.New("billing cycle conflict")
)

const (
	BillingIntervalMonthly = "monthly"
	BillingIntervalAnnual  = "annual"

	SubscriptionPendingPayment = "pending_payment"
	SubscriptionActive         = "active"
	SubscriptionPastDue        = "past_due"
	SubscriptionGrace          = "grace"
	SubscriptionSuspended      = "suspended"
	SubscriptionCancelled      = "cancelled"
	SubscriptionEnded          = "ended"

	QuoteSubmitted         = "submitted"
	QuoteUnderReview       = "under_review"
	QuoteCapacityConfirmed = "capacity_confirmed"
	QuoteQuoted            = "quoted"
	QuoteAccepted          = "accepted"
	QuoteProvisioning      = "provisioning"
	QuoteActive            = "active"
	QuoteRejected          = "rejected"
	QuoteExpired           = "expired"
	QuoteWithdrawn         = "withdrawn"
)

// CatalogVersion is an immutable commercial snapshot. Existing subscriptions
// and documents retain this record even after a newer package version becomes
// active.
type CatalogVersion struct {
	ID                string         `json:"id"`
	PackageID         string         `json:"package_id"`
	Version           int            `json:"version"`
	MonthlyPriceCents int            `json:"monthly_price_cents"`
	AnnualPriceCents  int            `json:"annual_price_cents"`
	AnnualDiscountBps int            `json:"annual_discount_bps"`
	Currency          string         `json:"currency"`
	TaxRateBps        int            `json:"tax_rate_bps"`
	RulesSnapshot     map[string]any `json:"rules_snapshot"`
	EffectiveFrom     time.Time      `json:"effective_from"`
	EffectiveUntil    *time.Time     `json:"effective_until,omitempty"`
	Status            string         `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
}

// PriceCalculation is the server-authoritative amount snapshot used by
// checkout, subscriptions, quote indications, and billing cycles.
type PriceCalculation struct {
	PackageID          string    `json:"package_id"`
	PackageVersionID   string    `json:"package_version_id"`
	PackageName        string    `json:"package_name"`
	DeploymentMode     string    `json:"deployment_mode"`
	PurchaseMode       string    `json:"purchase_mode"`
	BillingInterval    string    `json:"billing_interval"`
	BasePriceCents     int       `json:"base_price_cents"`
	AddonsCents        int       `json:"addons_cents"`
	SetupFeesCents     int       `json:"setup_fees_cents"`
	ProrationCents     int       `json:"proration_cents"`
	SubtotalCents      int       `json:"subtotal_cents"`
	DiscountCents      int       `json:"discount_cents"`
	CreditsCents       int       `json:"credits_cents"`
	TaxableAmountCents int       `json:"taxable_amount_cents"`
	TaxCents           int       `json:"tax_cents"`
	AmountDueCents     int       `json:"amount_due_cents"`
	Currency           string    `json:"currency"`
	TaxRateBps         int       `json:"tax_rate_bps"`
	AnnualDiscountBps  int       `json:"annual_discount_bps"`
	CheckoutEligible   bool      `json:"checkout_eligible"`
	QuoteRequired      bool      `json:"quote_required"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

type DedicatedQuoteRequest struct {
	ID                  string         `json:"id"`
	TenantID            string         `json:"tenant_id"`
	PackageVersionID    string         `json:"package_version_id"`
	PackageID           string         `json:"package_id"`
	PackageName         string         `json:"package_name"`
	CompanyLegalName    string         `json:"company_legal_name"`
	ContactName         string         `json:"contact_name"`
	ContactEmail        string         `json:"contact_email"`
	ContactPhone        string         `json:"contact_phone"`
	TaxRegistrationID   string         `json:"tax_registration_id"`
	CompanySize         string         `json:"company_size"`
	ExpectedConcurrency int            `json:"expected_concurrency"`
	PreferredRegion     string         `json:"preferred_region"`
	Notes               string         `json:"notes"`
	Status              string         `json:"status"`
	QuotedAmountCents   *int           `json:"quoted_amount_cents,omitempty"`
	Currency            string         `json:"currency"`
	QuoteExpiresAt      *time.Time     `json:"quote_expires_at,omitempty"`
	CapacitySnapshot    map[string]any `json:"capacity_snapshot"`
	IdempotencyKey      string         `json:"idempotency_key,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	CreatedBy           string         `json:"created_by"`
	UpdatedBy           string         `json:"updated_by"`
}

type CreateDedicatedQuoteInput struct {
	TenantID            string
	PackageID           string
	CompanyLegalName    string
	ContactName         string
	ContactEmail        string
	ContactPhone        string
	TaxRegistrationID   string
	CompanySize         string
	ExpectedConcurrency int
	PreferredRegion     string
	Notes               string
	IdempotencyKey      string
}

type QuoteTransitionInput struct {
	Status            string
	QuotedAmountCents *int
	Currency          string
	QuoteExpiresAt    *time.Time
	CapacitySnapshot  map[string]any
}

type TenantSubscription struct {
	ID                  string           `json:"id"`
	TenantID            string           `json:"tenant_id"`
	PackageVersionID    string           `json:"package_version_id"`
	PackageID           string           `json:"package_id"`
	EntitlementID       string           `json:"entitlement_id,omitempty"`
	QuoteRequestID      string           `json:"quote_request_id,omitempty"`
	InitialOrderID      string           `json:"initial_order_id,omitempty"`
	BillingInterval     string           `json:"billing_interval"`
	Status              string           `json:"status"`
	BillingAnchor       time.Time        `json:"billing_anchor"`
	CurrentPeriodStart  time.Time        `json:"current_period_start"`
	CurrentPeriodEnd    time.Time        `json:"current_period_end"`
	NextBillAt          *time.Time       `json:"next_bill_at,omitempty"`
	GraceUntil          *time.Time       `json:"grace_until,omitempty"`
	CalculationSnapshot PriceCalculation `json:"calculation_snapshot"`
	Provider            string           `json:"provider"`
	PaymentMethod       string           `json:"payment_method"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type CreateCommercialOrderInput struct {
	TenantID        string
	PackageID       string
	BillingInterval string
	Provider        string
	PaymentMethod   string
	At              time.Time
}

func (s *Store) ensureCommercialSchema(ctx context.Context) error {
	if s == nil || s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.package_versions (
  id text PRIMARY KEY,
  package_id text NOT NULL REFERENCES %s.packages(id),
  version integer NOT NULL CHECK (version > 0),
  monthly_price_cents integer NOT NULL CHECK (monthly_price_cents >= 0),
  annual_price_cents integer NOT NULL CHECK (annual_price_cents >= 0),
  annual_discount_bps integer NOT NULL DEFAULT 2000 CHECK (annual_discount_bps BETWEEN 0 AND 10000),
  currency text NOT NULL,
  tax_rate_bps integer NOT NULL DEFAULT 700 CHECK (tax_rate_bps BETWEEN 0 AND 10000),
  rules_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb,
  effective_from timestamptz NOT NULL DEFAULT now(),
  effective_until timestamptz,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('draft','active','retired')),
  %s,
  UNIQUE (package_id, version),
  CHECK (effective_until IS NULL OR effective_until > effective_from)
);
CREATE UNIQUE INDEX IF NOT EXISTS package_versions_one_active_idx
  ON %s.package_versions (package_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS %s.dedicated_quote_requests (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  package_version_id text NOT NULL REFERENCES %s.package_versions(id),
  company_legal_name text NOT NULL,
  contact_name text NOT NULL,
  contact_email text NOT NULL,
  contact_phone text NOT NULL,
  tax_registration_id text NOT NULL DEFAULT '',
  company_size text NOT NULL DEFAULT '',
  expected_concurrency integer NOT NULL CHECK (expected_concurrency > 0),
  preferred_region text NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'submitted'
    CHECK (status IN ('submitted','under_review','capacity_confirmed','quoted','accepted','provisioning','active','rejected','expired','withdrawn')),
  quoted_amount_cents integer CHECK (quoted_amount_cents >= 0),
  currency text NOT NULL DEFAULT 'THB',
  quote_expires_at timestamptz,
  capacity_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb,
  idempotency_key text NOT NULL DEFAULT '',
  %s
);
CREATE INDEX IF NOT EXISTS dedicated_quotes_tenant_created_idx
  ON %s.dedicated_quote_requests (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS dedicated_quotes_status_created_idx
  ON %s.dedicated_quote_requests (status, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS dedicated_quotes_tenant_idem_idx
  ON %s.dedicated_quote_requests (tenant_id, idempotency_key)
  WHERE idempotency_key <> '';

CREATE TABLE IF NOT EXISTS %s.tenant_subscriptions (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  package_version_id text NOT NULL REFERENCES %s.package_versions(id),
  entitlement_id text NOT NULL DEFAULT '',
  quote_request_id text REFERENCES %s.dedicated_quote_requests(id),
  initial_order_id text UNIQUE REFERENCES %s.payment_orders(id),
  billing_interval text NOT NULL CHECK (billing_interval IN ('monthly','annual')),
  status text NOT NULL
    CHECK (status IN ('pending_payment','active','past_due','grace','suspended','cancelled','ended')),
  billing_anchor timestamptz NOT NULL,
  current_period_start timestamptz NOT NULL,
  current_period_end timestamptz NOT NULL,
  next_bill_at timestamptz,
  grace_until timestamptz,
  calculation_snapshot jsonb NOT NULL,
  provider text NOT NULL DEFAULT '',
  payment_method text NOT NULL DEFAULT '',
  %s,
  CHECK (current_period_end > current_period_start)
);
CREATE INDEX IF NOT EXISTS tenant_subscriptions_tenant_created_idx
  ON %s.tenant_subscriptions (tenant_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS tenant_subscriptions_one_live_idx
  ON %s.tenant_subscriptions (tenant_id)
  WHERE status IN ('active','past_due','grace','suspended');

CREATE TABLE IF NOT EXISTS %s.billing_cycles (
  id text PRIMARY KEY,
  subscription_id text NOT NULL REFERENCES %s.tenant_subscriptions(id) ON DELETE CASCADE,
  period_key text NOT NULL,
  period_start timestamptz NOT NULL,
  period_end timestamptz NOT NULL,
  status text NOT NULL
    CHECK (status IN ('scheduled','previewed','payment_pending','paid','documents_issued','settled','retry_wait','failed')),
  calculation_snapshot jsonb NOT NULL,
  order_id text REFERENCES %s.payment_orders(id),
  receipt_id text,
  tax_invoice_id text,
  attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
  next_attempt_at timestamptz,
  last_error_code text NOT NULL DEFAULT '',
  %s,
  UNIQUE (subscription_id, period_key),
  CHECK (period_end > period_start)
);
CREATE INDEX IF NOT EXISTS billing_cycles_retry_idx
  ON %s.billing_cycles (status, next_attempt_at);
`,
		schema, schema, auditColumnsDDL, schema,
		schema, schema, schema, auditColumnsDDL, schema, schema, schema,
		schema, schema, schema, schema, schema, auditColumnsDDL, schema, schema,
		schema, schema, schema, auditColumnsDDL, schema))
	if err != nil {
		return err
	}
	return s.seedCommercialCatalogVersions(ctx)
}

func (s *Store) seedCommercialCatalogVersions(ctx context.Context) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.package_versions (
  id, package_id, version, monthly_price_cents, annual_price_cents,
  annual_discount_bps, currency, tax_rate_bps, rules_snapshot,
  effective_from, status, created_by, updated_by
)
SELECT
  'pkgv_' || replace(p.id, '-', '_') || '_v1',
  p.id,
  1,
  p.price_cents,
  ((p.price_cents::bigint * 12 * 8000 + 5000) / 10000)::integer,
  2000,
  p.currency,
  700,
  pl.rules,
  now(),
  'active',
  'system',
  'system'
FROM %s.packages p
JOIN %s.package_limits pl ON pl.package_id = p.id
WHERE p.status = 'active'
  AND NOT EXISTS (
    SELECT 1 FROM %s.package_versions existing WHERE existing.package_id = p.id
  )
ON CONFLICT (package_id, version) DO NOTHING`, schema, schema, schema, schema))
	return err
}

// SyncPackageCatalogVersion creates a new immutable version when commercial
// price, currency, or quota rules change. Unchanged package metadata such as a
// display description does not churn versions.
func (s *Store) SyncPackageCatalogVersion(ctx context.Context, pkg Package) (*CatalogVersion, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if pkg.Status != "active" {
		return nil, ErrCatalogVersionNotFound
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var currentID, currentCurrency string
	var currentMonthly, currentVersion int
	var currentRulesRaw []byte
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, version, monthly_price_cents, currency, rules_snapshot
FROM %s.package_versions
WHERE package_id = $1 AND status = 'active'
ORDER BY version DESC
LIMIT 1
FOR UPDATE`, schema), pkg.ID).Scan(
		&currentID, &currentVersion, &currentMonthly, &currentCurrency, &currentRulesRaw,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		var currentRules map[string]any
		_ = json.Unmarshal(currentRulesRaw, &currentRules)
		if currentMonthly == pkg.PriceCents && currentCurrency == pkg.Currency && reflect.DeepEqual(currentRules, pkg.Rules) {
			if err := tx.Commit(ctx); err != nil {
				return nil, err
			}
			return s.GetEffectiveCatalogVersion(ctx, pkg.ID, time.Now().UTC())
		}
	}

	nextVersion := 1
	if currentVersion > 0 {
		nextVersion = currentVersion + 1
	}
	var maxVersion int
	if maxErr := tx.QueryRow(ctx, fmt.Sprintf(`
SELECT COALESCE(MAX(version), 0) FROM %s.package_versions WHERE package_id = $1`, schema), pkg.ID).Scan(&maxVersion); maxErr != nil {
		return nil, maxErr
	}
	if maxVersion >= nextVersion {
		nextVersion = maxVersion + 1
	}
	now := time.Now().UTC()
	if currentID != "" {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.package_versions
SET status = 'retired', effective_until = $2, updated_by = $3
WHERE id = $1 AND status = 'active'`, schema), currentID, now, actor)
		if err != nil {
			return nil, err
		}
	}
	rulesJSON, err := json.Marshal(pkg.Rules)
	if err != nil {
		return nil, err
	}
	id := "pkgv_" + newStoreID()
	annual := roundedBasisPoints(pkg.PriceCents*12, 8000)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.package_versions (
  id, package_id, version, monthly_price_cents, annual_price_cents,
  annual_discount_bps, currency, tax_rate_bps, rules_snapshot,
  effective_from, status, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,2000,$6,700,$7::jsonb,$8,'active',$9,$9)`, schema),
		id, pkg.ID, nextVersion, pkg.PriceCents, annual, pkg.Currency, string(rulesJSON), now, actor)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetEffectiveCatalogVersion(ctx, pkg.ID, time.Now().UTC())
}

func scanCatalogVersion(row pgx.Row) (*CatalogVersion, error) {
	var out CatalogVersion
	var rules []byte
	err := row.Scan(
		&out.ID, &out.PackageID, &out.Version, &out.MonthlyPriceCents, &out.AnnualPriceCents,
		&out.AnnualDiscountBps, &out.Currency, &out.TaxRateBps, &rules,
		&out.EffectiveFrom, &out.EffectiveUntil, &out.Status, &out.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCatalogVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rules, &out.RulesSnapshot)
	return &out, nil
}

const catalogVersionSelect = `id, package_id, version, monthly_price_cents, annual_price_cents,
  annual_discount_bps, currency, tax_rate_bps, rules_snapshot,
  effective_from, effective_until, status, created_at`

func (s *Store) GetEffectiveCatalogVersion(ctx context.Context, packageID string, at time.Time) (*CatalogVersion, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanCatalogVersion(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s FROM %s.package_versions
WHERE package_id = $1
  AND status = 'active'
  AND effective_from <= $2
  AND (effective_until IS NULL OR effective_until > $2)
ORDER BY version DESC
LIMIT 1`, catalogVersionSelect, schema), strings.TrimSpace(packageID), at.UTC()))
}

// CalculateCatalogPrice performs no I/O and ignores all browser-supplied
// amounts. Callers select a persisted catalog version before invoking it.
func CalculateCatalogPrice(pkg Package, version CatalogVersion, interval string, at time.Time) (PriceCalculation, error) {
	interval = strings.ToLower(strings.TrimSpace(interval))
	if interval == "" {
		interval = BillingIntervalMonthly
	}
	if interval != BillingIntervalMonthly && interval != BillingIntervalAnnual {
		return PriceCalculation{}, fmt.Errorf("%w: billing_interval must be monthly or annual", ErrInvalidCommercialRequest)
	}
	if version.PackageID != pkg.ID || version.Status != "active" {
		return PriceCalculation{}, ErrCatalogVersionNotFound
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	base := version.MonthlyPriceCents
	discount := 0
	if interval == BillingIntervalAnnual {
		base = version.MonthlyPriceCents * 12
		discount = base - version.AnnualPriceCents
		if discount < 0 {
			discount = 0
		}
	}
	subtotal := base
	taxable := subtotal - discount
	if taxable < 0 {
		taxable = 0
	}
	tax := roundedBasisPoints(taxable, version.TaxRateBps)
	mode := PackagePurchaseMode(pkg)
	return PriceCalculation{
		PackageID: pkg.ID, PackageVersionID: version.ID, PackageName: pkg.Name,
		DeploymentMode: PackageDeployment(pkg), PurchaseMode: mode, BillingInterval: interval,
		BasePriceCents: base, SubtotalCents: subtotal, DiscountCents: discount,
		TaxableAmountCents: taxable, TaxCents: tax, AmountDueCents: taxable + tax,
		Currency: version.Currency, TaxRateBps: version.TaxRateBps,
		AnnualDiscountBps: version.AnnualDiscountBps,
		CheckoutEligible:  mode == PurchaseModeSelfServe, QuoteRequired: mode == PurchaseModeQuote,
		CalculatedAt: at.UTC(),
	}, nil
}

func roundedBasisPoints(amount, basisPoints int) int {
	if amount <= 0 || basisPoints <= 0 {
		return 0
	}
	return (amount*basisPoints + 5000) / 10000
}

func (s *Store) CalculatePackagePrice(ctx context.Context, packageID, interval string, at time.Time) (PriceCalculation, error) {
	pkg, err := s.GetPackage(ctx, strings.TrimSpace(packageID))
	if err != nil {
		return PriceCalculation{}, err
	}
	if pkg.Status != "active" {
		return PriceCalculation{}, ErrPackageNotFound
	}
	version, err := s.GetEffectiveCatalogVersion(ctx, pkg.ID, at)
	if err != nil {
		return PriceCalculation{}, err
	}
	return CalculateCatalogPrice(*pkg, *version, interval, at)
}

func normalizeDedicatedQuoteInput(in CreateDedicatedQuoteInput) (CreateDedicatedQuoteInput, error) {
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.PackageID = strings.TrimSpace(in.PackageID)
	in.CompanyLegalName = strings.TrimSpace(in.CompanyLegalName)
	in.ContactName = strings.TrimSpace(in.ContactName)
	in.ContactEmail = strings.ToLower(strings.TrimSpace(in.ContactEmail))
	in.ContactPhone = strings.TrimSpace(in.ContactPhone)
	in.TaxRegistrationID = strings.TrimSpace(in.TaxRegistrationID)
	in.CompanySize = strings.TrimSpace(in.CompanySize)
	in.PreferredRegion = strings.TrimSpace(in.PreferredRegion)
	in.Notes = strings.TrimSpace(in.Notes)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)

	if in.TenantID == "" || in.PackageID == "" || in.CompanyLegalName == "" ||
		in.ContactName == "" || in.ContactEmail == "" || in.ContactPhone == "" {
		return in, fmt.Errorf("%w: company, contact, email, phone, and package are required", ErrInvalidCommercialRequest)
	}
	if _, err := mail.ParseAddress(in.ContactEmail); err != nil {
		return in, fmt.Errorf("%w: contact_email is invalid", ErrInvalidCommercialRequest)
	}
	if in.ExpectedConcurrency <= 0 || in.ExpectedConcurrency > 1_000_000 {
		return in, fmt.Errorf("%w: expected_concurrency must be between 1 and 1000000", ErrInvalidCommercialRequest)
	}
	if utf8.RuneCountInString(in.CompanyLegalName) > 200 || utf8.RuneCountInString(in.ContactName) > 120 ||
		utf8.RuneCountInString(in.ContactPhone) > 40 || utf8.RuneCountInString(in.TaxRegistrationID) > 80 ||
		utf8.RuneCountInString(in.Notes) > 2000 || utf8.RuneCountInString(in.IdempotencyKey) > 160 {
		return in, fmt.Errorf("%w: one or more fields exceed the allowed length", ErrInvalidCommercialRequest)
	}
	if in.CompanySize != "" {
		switch in.CompanySize {
		case "1-10", "11-50", "51-200", "201-500", "501+":
		default:
			return in, fmt.Errorf("%w: company_size is invalid", ErrInvalidCommercialRequest)
		}
	}
	if in.PreferredRegion != "" {
		switch in.PreferredRegion {
		case "th-bangkok", "sg-singapore", "jp-tokyo", "eu-frankfurt", "other":
		default:
			return in, fmt.Errorf("%w: preferred_region is invalid", ErrInvalidCommercialRequest)
		}
	}
	return in, nil
}

func (s *Store) CreateDedicatedQuote(ctx context.Context, input CreateDedicatedQuoteInput) (*DedicatedQuoteRequest, bool, error) {
	if s == nil || s.pg == nil {
		return nil, false, errors.New("postgres unavailable")
	}
	input, err := normalizeDedicatedQuoteInput(input)
	if err != nil {
		return nil, false, err
	}
	pkg, err := s.GetPackage(ctx, input.PackageID)
	if err != nil {
		return nil, false, err
	}
	if pkg.Status != "active" || PackagePurchaseMode(*pkg) != PurchaseModeQuote || PackageDeployment(*pkg) != DeploymentDedicatedVM {
		return nil, false, fmt.Errorf("%w: package is not a dedicated quotation plan", ErrInvalidCommercialRequest)
	}
	version, err := s.GetEffectiveCatalogVersion(ctx, pkg.ID, time.Now().UTC())
	if err != nil {
		return nil, false, err
	}
	if input.IdempotencyKey != "" {
		prior, priorErr := s.GetDedicatedQuoteByIdempotency(ctx, input.TenantID, input.IdempotencyKey)
		if priorErr == nil {
			if !dedicatedQuoteMatchesInput(prior, input) {
				return nil, false, ErrIdempotencyConflict
			}
			return prior, true, nil
		}
		if !errors.Is(priorErr, ErrQuoteNotFound) {
			return nil, false, priorErr
		}
	}

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	id := "dqr_" + newStoreID()
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.dedicated_quote_requests (
  id, tenant_id, package_version_id, company_legal_name,
  contact_name, contact_email, contact_phone, tax_registration_id,
  company_size, expected_concurrency, preferred_region, notes,
  status, currency, idempotency_key, created_by, updated_by
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,
  'submitted',$13,$14,$15,$15
)`, schema),
		id, input.TenantID, version.ID, input.CompanyLegalName,
		input.ContactName, input.ContactEmail, input.ContactPhone, input.TaxRegistrationID,
		input.CompanySize, input.ExpectedConcurrency, input.PreferredRegion, input.Notes,
		version.Currency, input.IdempotencyKey, actor)
	if err != nil {
		if input.IdempotencyKey != "" && strings.Contains(strings.ToLower(err.Error()), "unique") {
			prior, priorErr := s.GetDedicatedQuoteByIdempotency(ctx, input.TenantID, input.IdempotencyKey)
			if priorErr == nil {
				if !dedicatedQuoteMatchesInput(prior, input) {
					return nil, false, ErrIdempotencyConflict
				}
				return prior, true, nil
			}
		}
		return nil, false, err
	}
	out, err := s.GetDedicatedQuote(ctx, id)
	return out, false, err
}

func dedicatedQuoteMatchesInput(prior *DedicatedQuoteRequest, input CreateDedicatedQuoteInput) bool {
	return prior != nil &&
		prior.PackageID == input.PackageID &&
		prior.CompanyLegalName == input.CompanyLegalName &&
		prior.ContactName == input.ContactName &&
		prior.ContactEmail == input.ContactEmail &&
		prior.ContactPhone == input.ContactPhone &&
		prior.TaxRegistrationID == input.TaxRegistrationID &&
		prior.CompanySize == input.CompanySize &&
		prior.ExpectedConcurrency == input.ExpectedConcurrency &&
		prior.PreferredRegion == input.PreferredRegion &&
		prior.Notes == input.Notes
}

func scanDedicatedQuote(row pgx.Row) (*DedicatedQuoteRequest, error) {
	var out DedicatedQuoteRequest
	var capacity []byte
	err := row.Scan(
		&out.ID, &out.TenantID, &out.PackageVersionID, &out.PackageID, &out.PackageName,
		&out.CompanyLegalName, &out.ContactName, &out.ContactEmail, &out.ContactPhone,
		&out.TaxRegistrationID, &out.CompanySize, &out.ExpectedConcurrency, &out.PreferredRegion,
		&out.Notes, &out.Status, &out.QuotedAmountCents, &out.Currency, &out.QuoteExpiresAt,
		&capacity, &out.IdempotencyKey, &out.CreatedAt, &out.UpdatedAt, &out.CreatedBy, &out.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrQuoteNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(capacity, &out.CapacitySnapshot)
	if out.CapacitySnapshot == nil {
		out.CapacitySnapshot = map[string]any{}
	}
	return &out, nil
}

const dedicatedQuoteSelect = `q.id, q.tenant_id, q.package_version_id, pv.package_id, p.name,
  q.company_legal_name, q.contact_name, q.contact_email, q.contact_phone,
  q.tax_registration_id, q.company_size, q.expected_concurrency, q.preferred_region,
  q.notes, q.status, q.quoted_amount_cents, q.currency, q.quote_expires_at,
  q.capacity_snapshot, q.idempotency_key, q.created_at, q.updated_at, q.created_by, q.updated_by`

func (s *Store) GetDedicatedQuote(ctx context.Context, id string) (*DedicatedQuoteRequest, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanDedicatedQuote(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.dedicated_quote_requests q
JOIN %s.package_versions pv ON pv.id = q.package_version_id
JOIN %s.packages p ON p.id = pv.package_id
WHERE q.id = $1`, dedicatedQuoteSelect, schema, schema, schema), strings.TrimSpace(id)))
}

func (s *Store) GetDedicatedQuoteForTenant(ctx context.Context, tenantID, id string) (*DedicatedQuoteRequest, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanDedicatedQuote(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.dedicated_quote_requests q
JOIN %s.package_versions pv ON pv.id = q.package_version_id
JOIN %s.packages p ON p.id = pv.package_id
WHERE q.tenant_id = $1 AND q.id = $2`, dedicatedQuoteSelect, schema, schema, schema),
		strings.TrimSpace(tenantID), strings.TrimSpace(id)))
}

func (s *Store) GetDedicatedQuoteByIdempotency(ctx context.Context, tenantID, key string) (*DedicatedQuoteRequest, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanDedicatedQuote(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.dedicated_quote_requests q
JOIN %s.package_versions pv ON pv.id = q.package_version_id
JOIN %s.packages p ON p.id = pv.package_id
WHERE q.tenant_id = $1 AND q.idempotency_key = $2`, dedicatedQuoteSelect, schema, schema, schema),
		strings.TrimSpace(tenantID), strings.TrimSpace(key)))
}

func (s *Store) ListDedicatedQuotes(ctx context.Context, tenantID, status string, limit int) ([]DedicatedQuoteRequest, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	query := fmt.Sprintf(`
SELECT %s
FROM %s.dedicated_quote_requests q
JOIN %s.package_versions pv ON pv.id = q.package_version_id
JOIN %s.packages p ON p.id = pv.package_id
WHERE ($1 = '' OR q.tenant_id = $1)
  AND ($2 = '' OR q.status = $2)
ORDER BY q.created_at DESC
LIMIT $3`, dedicatedQuoteSelect, schema, schema, schema)
	rows, err := s.pg.Query(ctx, query, strings.TrimSpace(tenantID), strings.TrimSpace(status), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DedicatedQuoteRequest, 0)
	for rows.Next() {
		item, scanErr := scanDedicatedQuote(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func validQuoteTransition(from, to string) bool {
	allowed := map[string]map[string]bool{
		QuoteSubmitted:         {QuoteUnderReview: true, QuoteWithdrawn: true},
		QuoteUnderReview:       {QuoteCapacityConfirmed: true, QuoteRejected: true, QuoteWithdrawn: true},
		QuoteCapacityConfirmed: {QuoteQuoted: true, QuoteRejected: true},
		QuoteQuoted:            {QuoteAccepted: true, QuoteExpired: true, QuoteRejected: true},
		QuoteAccepted:          {QuoteProvisioning: true},
		QuoteProvisioning:      {QuoteActive: true, QuoteRejected: true},
	}
	return allowed[from][to]
}

func validCommercialCurrency(value string) bool {
	if len(value) < 3 || len(value) > 8 {
		return false
	}
	for _, char := range value {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

func (s *Store) TransitionDedicatedQuote(ctx context.Context, id string, input QuoteTransitionInput) (*DedicatedQuoteRequest, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	target := strings.ToLower(strings.TrimSpace(input.Status))
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	current, err := scanDedicatedQuote(tx.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.dedicated_quote_requests q
JOIN %s.package_versions pv ON pv.id = q.package_version_id
JOIN %s.packages p ON p.id = pv.package_id
WHERE q.id = $1
FOR UPDATE OF q`, dedicatedQuoteSelect, schema, schema, schema), strings.TrimSpace(id)))
	if err != nil {
		return nil, err
	}
	if target == current.Status {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return current, nil
	}
	if !validQuoteTransition(current.Status, target) {
		return nil, ErrQuoteStateConflict
	}
	if target != QuoteCapacityConfirmed && input.CapacitySnapshot != nil {
		return nil, fmt.Errorf("%w: capacity_snapshot can only be set during capacity confirmation", ErrInvalidCommercialRequest)
	}
	if target != QuoteQuoted &&
		(input.QuotedAmountCents != nil || strings.TrimSpace(input.Currency) != "" || input.QuoteExpiresAt != nil) {
		return nil, fmt.Errorf("%w: quote terms can only be set when issuing the quotation", ErrInvalidCommercialRequest)
	}

	amount := current.QuotedAmountCents
	currency := current.Currency
	expires := current.QuoteExpiresAt
	capacity := current.CapacitySnapshot
	if input.QuotedAmountCents != nil {
		amount = input.QuotedAmountCents
	}
	if strings.TrimSpace(input.Currency) != "" {
		currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	}
	if input.QuoteExpiresAt != nil {
		value := input.QuoteExpiresAt.UTC()
		expires = &value
	}
	if input.CapacitySnapshot != nil {
		capacity = input.CapacitySnapshot
	}
	if target == QuoteCapacityConfirmed && len(capacity) == 0 {
		return nil, fmt.Errorf("%w: capacity_snapshot is required", ErrInvalidCommercialRequest)
	}
	if target == QuoteQuoted {
		if amount == nil || *amount <= 0 || *amount > 2_000_000_000 ||
			!validCommercialCurrency(currency) || expires == nil || !expires.After(time.Now().UTC()) {
			return nil, fmt.Errorf("%w: positive quoted_amount_cents and future quote_expires_at are required", ErrInvalidCommercialRequest)
		}
	}
	if target == QuoteAccepted && (current.QuoteExpiresAt == nil || !current.QuoteExpiresAt.After(time.Now().UTC())) {
		return nil, ErrQuoteStateConflict
	}
	capacityJSON, err := json.Marshal(capacity)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid capacity_snapshot", ErrInvalidCommercialRequest)
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.dedicated_quote_requests
SET status = $2, quoted_amount_cents = $3, currency = $4, quote_expires_at = $5,
    capacity_snapshot = $6::jsonb, updated_by = $7
WHERE id = $1`, schema), current.ID, target, amount, currency, expires, string(capacityJSON), actor)
	if err != nil {
		return nil, err
	}
	if target == QuoteActive {
		if err := s.activateDedicatedQuoteTx(ctx, tx, current, amount, currency, actor); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetDedicatedQuote(ctx, current.ID)
}

func (s *Store) activateDedicatedQuoteTx(
	ctx context.Context,
	tx pgx.Tx,
	quote *DedicatedQuoteRequest,
	amount *int,
	currency string,
	actor string,
) error {
	if amount == nil || *amount <= 0 {
		return fmt.Errorf("%w: active quote requires a quoted amount", ErrInvalidCommercialRequest)
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var rulesSchemaID string
	var rulesJSON []byte
	if err := tx.QueryRow(ctx, fmt.Sprintf(`
SELECT pl.rules_schema_id, pv.rules_snapshot
FROM %s.package_versions pv
JOIN %s.package_limits pl ON pl.package_id = pv.package_id
WHERE pv.id = $1`, schema, schema), quote.PackageVersionID).Scan(&rulesSchemaID, &rulesJSON); err != nil {
		return err
	}
	entitlementID := "ent_" + quote.TenantID + "_" + quote.PackageID + "_" + newStoreID()
	_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_entitlements SET status = 'revoked', updated_by = $2
WHERE tenant_id = $1 AND status = 'active'`, schema), quote.TenantID, actor)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_entitlements (
  id, tenant_id, package_id, rules_schema_id, rules_snapshot, status, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5::jsonb,'active',$6,$6)`, schema),
		entitlementID, quote.TenantID, quote.PackageID, rulesSchemaID, string(rulesJSON), actor)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	end := addBillingInterval(now, BillingIntervalMonthly)
	calculation := PriceCalculation{
		PackageID: quote.PackageID, PackageVersionID: quote.PackageVersionID, PackageName: quote.PackageName,
		DeploymentMode: DeploymentDedicatedVM, PurchaseMode: PurchaseModeQuote,
		BillingInterval: BillingIntervalMonthly, BasePriceCents: *amount, SubtotalCents: *amount,
		TaxableAmountCents: *amount, AmountDueCents: *amount, Currency: currency,
		CheckoutEligible: false, QuoteRequired: true, CalculatedAt: now,
	}
	snapshot, err := json.Marshal(calculation)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_subscriptions
SET status = 'cancelled', next_bill_at = NULL, updated_by = $2
WHERE tenant_id = $1 AND status IN ('active','past_due','grace','suspended')`, schema), quote.TenantID, actor)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_subscriptions (
  id, tenant_id, package_version_id, entitlement_id, quote_request_id,
  billing_interval, status, billing_anchor, current_period_start, current_period_end,
  next_bill_at, calculation_snapshot, provider, payment_method, created_by, updated_by
) VALUES (
  $1,$2,$3,$4,$5,'monthly','active',$6,$6,$7,$7,$8::jsonb,'finance','invoice',$9,$9
)`, schema),
		"sub_"+newStoreID(), quote.TenantID, quote.PackageVersionID, entitlementID, quote.ID,
		now, end, string(snapshot), actor)
	return err
}

func addBillingInterval(start time.Time, interval string) time.Time {
	start = start.UTC()
	months := 1
	if interval == BillingIntervalAnnual {
		months = 12
	}
	targetMonthStart := time.Date(start.Year(), start.Month()+time.Month(months), 1,
		start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
	lastTargetDay := targetMonthStart.AddDate(0, 1, -1).Day()
	day := start.Day()
	if day > lastTargetDay {
		day = lastTargetDay
	}
	return time.Date(targetMonthStart.Year(), targetMonthStart.Month(), day,
		start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
}

func (s *Store) CreateCommercialPaymentOrder(ctx context.Context, input CreateCommercialOrderInput) (*PaymentOrder, *TenantSubscription, error) {
	if s == nil || s.pg == nil {
		return nil, nil, errors.New("postgres unavailable")
	}
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.PackageID = strings.TrimSpace(input.PackageID)
	if input.TenantID == "" || input.PackageID == "" {
		return nil, nil, ErrInvalidCommercialRequest
	}
	if input.At.IsZero() {
		input.At = time.Now().UTC()
	}
	calculation, err := s.CalculatePackagePrice(ctx, input.PackageID, input.BillingInterval, input.At)
	if err != nil {
		return nil, nil, err
	}
	if !calculation.CheckoutEligible || calculation.QuoteRequired {
		return nil, nil, ErrPackageRequiresQuote
	}

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	orderID := newPaymentOrderID()
	orderNo := newPaymentOrderNo(input.TenantID)
	subscriptionID := "sub_" + newStoreID()
	start := input.At.UTC()
	end := addBillingInterval(start, calculation.BillingInterval)
	snapshot, err := json.Marshal(calculation)
	if err != nil {
		return nil, nil, err
	}
	provider := strings.TrimSpace(input.Provider)
	method := strings.TrimSpace(input.PaymentMethod)
	if method == "" {
		method = "credit_card"
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_orders (
  id, tenant_id, package_id, order_no, amount_cents, currency, status,
  provider, payment_method, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,'pending',$7,$8,$9,$9)`, schema),
		orderID, input.TenantID, input.PackageID, orderNo, calculation.AmountDueCents,
		calculation.Currency, provider, method, actor)
	if err != nil {
		return nil, nil, err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_subscriptions (
  id, tenant_id, package_version_id, initial_order_id, billing_interval, status,
  billing_anchor, current_period_start, current_period_end, calculation_snapshot,
  provider, payment_method, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,'pending_payment',$6,$6,$7,$8::jsonb,$9,$10,$11,$11)`, schema),
		subscriptionID, input.TenantID, calculation.PackageVersionID, orderID,
		calculation.BillingInterval, start, end, string(snapshot), provider, method, actor)
	if err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	order, err := s.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		return nil, nil, err
	}
	subscription, err := s.GetTenantSubscription(ctx, subscriptionID)
	return order, subscription, err
}

func scanTenantSubscription(row pgx.Row) (*TenantSubscription, error) {
	var out TenantSubscription
	var snapshot []byte
	var quoteID, orderID *string
	err := row.Scan(
		&out.ID, &out.TenantID, &out.PackageVersionID, &out.PackageID,
		&out.EntitlementID, &quoteID, &orderID, &out.BillingInterval, &out.Status,
		&out.BillingAnchor, &out.CurrentPeriodStart, &out.CurrentPeriodEnd,
		&out.NextBillAt, &out.GraceUntil, &snapshot, &out.Provider, &out.PaymentMethod,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}
	if quoteID != nil {
		out.QuoteRequestID = *quoteID
	}
	if orderID != nil {
		out.InitialOrderID = *orderID
	}
	_ = json.Unmarshal(snapshot, &out.CalculationSnapshot)
	return &out, nil
}

const tenantSubscriptionSelect = `s.id, s.tenant_id, s.package_version_id, pv.package_id,
  s.entitlement_id, s.quote_request_id, s.initial_order_id, s.billing_interval, s.status,
  s.billing_anchor, s.current_period_start, s.current_period_end,
  s.next_bill_at, s.grace_until, s.calculation_snapshot, s.provider, s.payment_method,
  s.created_at, s.updated_at`

func (s *Store) GetTenantSubscription(ctx context.Context, id string) (*TenantSubscription, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanTenantSubscription(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.tenant_subscriptions s
JOIN %s.package_versions pv ON pv.id = s.package_version_id
WHERE s.id = $1`, tenantSubscriptionSelect, schema, schema), strings.TrimSpace(id)))
}

func (s *Store) GetLiveTenantSubscription(ctx context.Context, tenantID string) (*TenantSubscription, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanTenantSubscription(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.tenant_subscriptions s
JOIN %s.package_versions pv ON pv.id = s.package_version_id
WHERE s.tenant_id = $1 AND s.status IN ('active','past_due','grace','suspended')
ORDER BY s.created_at DESC LIMIT 1`, tenantSubscriptionSelect, schema, schema), strings.TrimSpace(tenantID)))
}

// GetOrderCommercialCalculation returns the immutable package/price snapshot
// attached to an initial subscription order or a renewal billing cycle.
func (s *Store) GetOrderCommercialCalculation(ctx context.Context, orderID string) (*PriceCalculation, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var raw []byte
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT calculation_snapshot
FROM (
  SELECT calculation_snapshot, created_at
  FROM %s.tenant_subscriptions
  WHERE initial_order_id = $1
  UNION ALL
  SELECT calculation_snapshot, created_at
  FROM %s.billing_cycles
  WHERE order_id = $1
) snapshots
ORDER BY created_at DESC
LIMIT 1`, schema, schema), strings.TrimSpace(orderID)).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}
	var calculation PriceCalculation
	if err := json.Unmarshal(raw, &calculation); err != nil {
		return nil, err
	}
	return &calculation, nil
}

func (s *Store) activateSubscriptionForPaidOrderTx(ctx context.Context, tx pgx.Tx, order *PaymentOrder, entitlementID, actor string) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_subscriptions
SET status = 'cancelled', next_bill_at = NULL, updated_by = $3
WHERE tenant_id = $1
  AND status IN ('active','past_due','grace','suspended')
  AND COALESCE(initial_order_id, '') <> $2`, schema), order.TenantID, order.ID, actor)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_subscriptions
SET entitlement_id = $2, status = 'active', next_bill_at = current_period_end,
    grace_until = NULL, updated_by = $3
WHERE initial_order_id = $1 AND status = 'pending_payment'`, schema), order.ID, entitlementID, actor)
	return err
}
