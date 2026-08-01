package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/libra/monti-jarvis/internal/auditctx"
)

const PaymentGatewayConfigID = "default"

var (
	ErrPaymentGatewayNotConfigured = errors.New("payment gateway not configured")
	ErrPaymentOrderNotFound        = errors.New("payment order not found")
)

const (
	PaymentOrderStatusPending   = "pending"
	PaymentOrderStatusPaid      = "paid"
	PaymentOrderStatusFailed    = "failed"
	PaymentOrderStatusCancelled = "cancelled"
)

type PaymentOrder struct {
	ID                 string
	TenantID           string
	PackageID          string
	OrderNo            string
	AmountCents        int
	Currency           string
	Status             string
	Provider           string
	PaymentMethod      string
	TransactionID      string
	PaymentURL         string
	ProviderSessionID  string
	ProviderPaymentID  string
	ProviderStatus     string
	CheckoutExpiresAt  *time.Time
	LastProviderSyncAt *time.Time
	PaidAt             *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreatePaymentOrderInput struct {
	TenantID      string
	PackageID     string
	AmountCents   int
	Currency      string
	Provider      string
	PaymentMethod string
}

// PaymentDocument is a receipt or tax invoice issued for a paid order.
type PaymentDocument struct {
	ID             string
	OrderID        string
	TenantID       string
	DocType        string // receipt | tax_invoice
	DocNumber      string
	Status         string // issued | voided
	BuyerName      string
	BuyerAddress   string
	BuyerTaxID     string
	SellerName     string
	SellerAddress  string
	SellerTaxID    string
	PackageName    string
	AmountCents    int
	Currency       string
	VATRateBps     int // basis points, e.g. 700 = 7%
	NetCents       int
	VATCents       int
	PaymentMethod  string
	ReissuedFromID string
	VoidReason     string
	VoidedAt       *time.Time
	IssuedAt       time.Time
	CreatedAt      time.Time
}

const (
	PaymentDocTypeReceipt    = "receipt"
	PaymentDocTypeTaxInvoice = "tax_invoice"
	PaymentDocStatusIssued   = "issued"
	PaymentDocStatusVoided   = "voided"
)

// PlatformSellerBranding is the platform-wide invoice/receipt seller block (singleton).
type PlatformSellerBranding struct {
	ID        string
	Name      string
	Address   string
	TaxID     string
	Branch    string
	UpdatedAt time.Time
}

const PlatformSellerBrandingID = "default"

// TenantTaxProfile holds buyer tax invoice fields (Sprint 12).
type TenantTaxProfile struct {
	TenantID    string
	CompanyName string
	TaxID       string
	Branch      string
	Address     string
	UpdatedAt   time.Time
}

// PaymentOrderListFilter for platform billing ledger (Sprint 10).
type PaymentOrderListFilter struct {
	TenantID string
	Status   string
	Limit    int
	Offset   int
}

// PaymentOrderListItem joins package name for ledger display.
type PaymentOrderListItem struct {
	PaymentOrder
	PackageName string
	TenantName  string
}

// PaymentDocumentListFilter for platform receipt console (Sprint 11).
type PaymentDocumentListFilter struct {
	TenantID string
	DocType  string
	Status   string
	Limit    int
	Offset   int
}

type PaymentFulfillResult struct {
	Order              PaymentOrder
	EntitlementChanged bool
}

type PaymentGatewayConfig struct {
	ID                   string
	Provider             string
	Mode                 string
	Status               string
	MerchantCode         string
	APIKey               string
	MD5Key               string
	BaseURL              string
	RouteNo              int
	Currency             string
	CallbackURL          string
	ReturnURL            string
	StripePublishableKey string
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripeAPIBaseURL     string
	StripeSuccessURL     string
	StripeCancelURL      string
	LastTestStatus       string
	LastTestedAt         *time.Time
	LastTestError        string
	LastWebhookStatus    string
	LastWebhookAt        *time.Time
	UpdatedAt            time.Time
}

type PaymentGatewayUpsert struct {
	Provider               string
	Mode                   string
	Status                 string
	MerchantCode           string
	APIKey                 string
	MD5Key                 string
	BaseURL                string
	RouteNo                int
	Currency               string
	CallbackURL            string
	ReturnURL              string
	SetAPIKey              bool
	SetMD5Key              bool
	StripePublishableKey   string
	StripeSecretKey        string
	StripeWebhookSecret    string
	StripeAPIBaseURL       string
	StripeSuccessURL       string
	StripeCancelURL        string
	SetStripeSecret        bool
	SetStripeWebhookSecret bool
}

type PaymentCallbackEvent struct {
	ID                string
	Provider          string
	TransactionID     string
	OrderNo           string
	ProviderEventID   string
	EventType         string
	SignatureVerified bool
	ProcessingStatus  string
	ProcessedAt       *time.Time
	ErrorCode         string
	PaymentStatus     string
	Amount            string
	CustomerID        string
	PayloadHash       string
	ReceivedAt        time.Time
}

func (s *Store) ensurePaymentSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.payment_gateway_configs (
  id text PRIMARY KEY,
  provider text NOT NULL DEFAULT '',
  mode text NOT NULL DEFAULT 'test' CHECK (mode IN ('test', 'live')),
  status text NOT NULL DEFAULT 'inactive' CHECK (status IN ('inactive', 'active')),
  merchant_code text NOT NULL DEFAULT '',
  api_key text NOT NULL DEFAULT '',
  md5_key text NOT NULL DEFAULT '',
  base_url text NOT NULL DEFAULT '',
  route_no integer NOT NULL DEFAULT 1,
  currency text NOT NULL DEFAULT '764',
  callback_url text NOT NULL DEFAULT '',
  return_url text NOT NULL DEFAULT '',
  stripe_publishable_key text NOT NULL DEFAULT '',
  stripe_secret_key text NOT NULL DEFAULT '',
  stripe_webhook_secret text NOT NULL DEFAULT '',
  stripe_api_base_url text NOT NULL DEFAULT '',
  stripe_success_url text NOT NULL DEFAULT '',
  stripe_cancel_url text NOT NULL DEFAULT '',
  last_test_status text NOT NULL DEFAULT 'unknown',
  last_tested_at timestamptz,
  last_test_error text NOT NULL DEFAULT '',
  last_webhook_status text NOT NULL DEFAULT 'unknown',
  last_webhook_at timestamptz,%s
);
CREATE TABLE IF NOT EXISTS %s.payment_callback_events (
  id text PRIMARY KEY,
  provider text NOT NULL,
  transaction_id text NOT NULL,
  order_no text NOT NULL DEFAULT '',
  provider_event_id text NOT NULL DEFAULT '',
  event_type text NOT NULL DEFAULT '',
  signature_verified boolean NOT NULL DEFAULT false,
  processing_status text NOT NULL DEFAULT 'received',
  processed_at timestamptz,
  error_code text NOT NULL DEFAULT '',
  payment_status text NOT NULL DEFAULT '',
  amount text NOT NULL DEFAULT '',
  customer_id text NOT NULL DEFAULT '',
  payload_hash text NOT NULL DEFAULT '',
  received_at timestamptz NOT NULL DEFAULT now(),%s,
  UNIQUE (provider, transaction_id)
);
CREATE INDEX IF NOT EXISTS payment_callback_events_received_idx
  ON %s.payment_callback_events (received_at DESC);
CREATE TABLE IF NOT EXISTS %s.payment_orders (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  package_id text NOT NULL REFERENCES %s.packages(id),
  order_no text NOT NULL UNIQUE,
  amount_cents int NOT NULL,
  currency text NOT NULL DEFAULT '764',
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
  provider text NOT NULL DEFAULT '',
  payment_method text NOT NULL DEFAULT 'credit_card',
  transaction_id text NOT NULL DEFAULT '',
  payment_url text NOT NULL DEFAULT '',
  provider_session_id text NOT NULL DEFAULT '',
  provider_payment_id text NOT NULL DEFAULT '',
  provider_status text NOT NULL DEFAULT '',
  checkout_expires_at timestamptz,
  last_provider_sync_at timestamptz,
  paid_at timestamptz,%s
);
CREATE INDEX IF NOT EXISTS payment_orders_tenant_status_idx
  ON %s.payment_orders (tenant_id, status);
CREATE INDEX IF NOT EXISTS payment_orders_tenant_created_idx
  ON %s.payment_orders (tenant_id, created_at DESC);
CREATE TABLE IF NOT EXISTS %s.payment_documents (
  id text PRIMARY KEY,
  order_id text NOT NULL REFERENCES %s.payment_orders(id) ON DELETE CASCADE,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  doc_type text NOT NULL CHECK (doc_type IN ('receipt', 'tax_invoice')),
  doc_number text NOT NULL UNIQUE,
  buyer_name text NOT NULL DEFAULT '',
  buyer_address text NOT NULL DEFAULT '',
  buyer_tax_id text NOT NULL DEFAULT '',
  seller_name text NOT NULL DEFAULT '',
  seller_address text NOT NULL DEFAULT '',
  seller_tax_id text NOT NULL DEFAULT '',
  package_name text NOT NULL DEFAULT '',
  amount_cents int NOT NULL DEFAULT 0,
  currency text NOT NULL DEFAULT '764',
  vat_rate_bps int NOT NULL DEFAULT 700,
  net_cents int NOT NULL DEFAULT 0,
  vat_cents int NOT NULL DEFAULT 0,
  payment_method text NOT NULL DEFAULT '',
  issued_at timestamptz NOT NULL DEFAULT now(),%s,
  UNIQUE (order_id, doc_type)
);
CREATE INDEX IF NOT EXISTS payment_documents_tenant_idx
  ON %s.payment_documents (tenant_id, issued_at DESC);`,
		schema, auditColumnsDDL, schema, auditColumnsDDL, schema, schema, schema, schema, auditColumnsDDL, schema, schema,
		schema, schema, schema, auditColumnsDDL, schema))
	if err != nil {
		return err
	}
	// Migrations for existing payment_orders / payment_documents rows.
	migrations := []string{
		fmt.Sprintf(`ALTER TABLE %s.payment_orders ADD COLUMN IF NOT EXISTS payment_method text NOT NULL DEFAULT 'credit_card'`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_orders ADD COLUMN IF NOT EXISTS provider_session_id text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_orders ADD COLUMN IF NOT EXISTS provider_payment_id text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_orders ADD COLUMN IF NOT EXISTS provider_status text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_orders ADD COLUMN IF NOT EXISTS checkout_expires_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_orders ADD COLUMN IF NOT EXISTS last_provider_sync_at timestamptz`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS payment_orders_provider_session_idx ON %s.payment_orders (provider, provider_session_id)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS payment_orders_provider_payment_idx ON %s.payment_orders (provider, provider_payment_id)`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS stripe_publishable_key text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS stripe_secret_key text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS stripe_webhook_secret text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS stripe_api_base_url text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS stripe_success_url text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS stripe_cancel_url text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS last_test_status text NOT NULL DEFAULT 'unknown'`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS last_tested_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS last_test_error text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS last_webhook_status text NOT NULL DEFAULT 'unknown'`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_gateway_configs ADD COLUMN IF NOT EXISTS last_webhook_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_callback_events ADD COLUMN IF NOT EXISTS provider_event_id text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_callback_events ADD COLUMN IF NOT EXISTS event_type text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_callback_events ADD COLUMN IF NOT EXISTS signature_verified boolean NOT NULL DEFAULT false`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_callback_events ADD COLUMN IF NOT EXISTS processing_status text NOT NULL DEFAULT 'received'`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_callback_events ADD COLUMN IF NOT EXISTS processed_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_callback_events ADD COLUMN IF NOT EXISTS error_code text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS payment_callback_events_provider_event_uidx
  ON %s.payment_callback_events (provider, provider_event_id) WHERE provider_event_id <> ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_documents ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'issued'`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_documents ADD COLUMN IF NOT EXISTS void_reason text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_documents ADD COLUMN IF NOT EXISTS voided_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_documents ADD COLUMN IF NOT EXISTS reissued_from_id text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.payment_documents DROP CONSTRAINT IF EXISTS payment_documents_order_id_doc_type_key`, schema),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS payment_documents_active_order_type_uidx
  ON %s.payment_documents (order_id, doc_type) WHERE status = 'issued'`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.platform_seller_branding (
  id text PRIMARY KEY,
  name text NOT NULL DEFAULT 'Monti Jarvis Platform',
  address text NOT NULL DEFAULT 'Bangkok, Thailand',
  tax_id text NOT NULL DEFAULT '0-0000-00000-00-0',
  branch text NOT NULL DEFAULT '00000',%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`INSERT INTO %s.platform_seller_branding (id, name, address, tax_id, branch)
VALUES ('default', 'Monti Jarvis Platform', 'Bangkok, Thailand', '0-0000-00000-00-0', '00000')
ON CONFLICT (id) DO NOTHING`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_tax_profiles (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  company_name text NOT NULL DEFAULT '',
  tax_id text NOT NULL DEFAULT '',
  branch text NOT NULL DEFAULT '00000',
  address text NOT NULL DEFAULT '',%s
)`, schema, schema, auditColumnsDDL),
	}
	for _, m := range migrations {
		if _, err := s.pg.Exec(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetPaymentGatewayConfig(ctx context.Context) (PaymentGatewayConfig, error) {
	if s.pg == nil {
		return PaymentGatewayConfig{}, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var row PaymentGatewayConfig
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, provider, mode, status, merchant_code, api_key, md5_key, base_url, route_no, currency,
       callback_url, return_url,
       COALESCE(stripe_publishable_key, ''), COALESCE(stripe_secret_key, ''), COALESCE(stripe_webhook_secret, ''),
       COALESCE(stripe_api_base_url, ''), COALESCE(stripe_success_url, ''), COALESCE(stripe_cancel_url, ''),
       COALESCE(last_test_status, 'unknown'), last_tested_at, COALESCE(last_test_error, ''),
       COALESCE(last_webhook_status, 'unknown'), last_webhook_at, updated_at
FROM %s.payment_gateway_configs
WHERE id = $1`, schema), PaymentGatewayConfigID).Scan(
		&row.ID, &row.Provider, &row.Mode, &row.Status, &row.MerchantCode, &row.APIKey, &row.MD5Key,
		&row.BaseURL, &row.RouteNo, &row.Currency, &row.CallbackURL, &row.ReturnURL,
		&row.StripePublishableKey, &row.StripeSecretKey, &row.StripeWebhookSecret,
		&row.StripeAPIBaseURL, &row.StripeSuccessURL, &row.StripeCancelURL,
		&row.LastTestStatus, &row.LastTestedAt, &row.LastTestError,
		&row.LastWebhookStatus, &row.LastWebhookAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PaymentGatewayConfig{
			ID: PaymentGatewayConfigID, Mode: "test", Status: "inactive", RouteNo: 1, Currency: "764",
			LastTestStatus: "unknown", LastWebhookStatus: "unknown",
		}, nil
	}
	if err != nil {
		return PaymentGatewayConfig{}, err
	}
	return row, nil
}

func (s *Store) UpsertPaymentGatewayConfig(ctx context.Context, in PaymentGatewayUpsert) (PaymentGatewayConfig, error) {
	if s.pg == nil {
		return PaymentGatewayConfig{}, errors.New("postgres unavailable")
	}
	current, err := s.GetPaymentGatewayConfig(ctx)
	if err != nil {
		return PaymentGatewayConfig{}, err
	}

	apiKey := current.APIKey
	if in.SetAPIKey {
		apiKey = in.APIKey
	}
	md5Key := current.MD5Key
	if in.SetMD5Key {
		md5Key = in.MD5Key
	}
	stripeSecret := current.StripeSecretKey
	if in.SetStripeSecret {
		stripeSecret = in.StripeSecretKey
	}
	stripeWebhookSecret := current.StripeWebhookSecret
	if in.SetStripeWebhookSecret {
		stripeWebhookSecret = in.StripeWebhookSecret
	}

	provider := strings.TrimSpace(in.Provider)
	if provider == "" {
		provider = current.Provider
	}
	mode := strings.TrimSpace(in.Mode)
	if mode == "" {
		mode = current.Mode
	}
	if mode == "" {
		mode = "test"
	}
	status := strings.TrimSpace(in.Status)
	if status == "" {
		if provider != "" {
			status = "active"
		} else {
			status = "inactive"
		}
	}
	routeNo := in.RouteNo
	if routeNo <= 0 {
		routeNo = current.RouteNo
	}
	if routeNo <= 0 {
		routeNo = 1
	}
	currency := strings.TrimSpace(in.Currency)
	if currency == "" {
		currency = current.Currency
	}
	if currency == "" {
		currency = "764"
	}
	callbackURL := strings.TrimSpace(in.CallbackURL)
	if callbackURL == "" {
		callbackURL = current.CallbackURL
	}
	returnURL := strings.TrimSpace(in.ReturnURL)
	if returnURL == "" {
		returnURL = current.ReturnURL
	}
	stripeAPIBaseURL := strings.TrimSpace(in.StripeAPIBaseURL)
	if stripeAPIBaseURL == "" {
		stripeAPIBaseURL = current.StripeAPIBaseURL
	}
	stripeSuccessURL := strings.TrimSpace(in.StripeSuccessURL)
	if stripeSuccessURL == "" {
		stripeSuccessURL = current.StripeSuccessURL
	}
	stripeCancelURL := strings.TrimSpace(in.StripeCancelURL)
	if stripeCancelURL == "" {
		stripeCancelURL = current.StripeCancelURL
	}

	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_gateway_configs (
  id, provider, mode, status, merchant_code, api_key, md5_key, base_url, route_no, currency,
  callback_url, return_url,
  stripe_publishable_key, stripe_secret_key, stripe_webhook_secret, stripe_api_base_url, stripe_success_url, stripe_cancel_url,
  created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$19)
ON CONFLICT (id) DO UPDATE SET
  provider = EXCLUDED.provider,
  mode = EXCLUDED.mode,
  status = EXCLUDED.status,
  merchant_code = EXCLUDED.merchant_code,
  api_key = EXCLUDED.api_key,
  md5_key = EXCLUDED.md5_key,
  base_url = EXCLUDED.base_url,
  route_no = EXCLUDED.route_no,
  currency = EXCLUDED.currency,
  callback_url = EXCLUDED.callback_url,
  return_url = EXCLUDED.return_url,
  stripe_publishable_key = EXCLUDED.stripe_publishable_key,
  stripe_secret_key = EXCLUDED.stripe_secret_key,
  stripe_webhook_secret = EXCLUDED.stripe_webhook_secret,
  stripe_api_base_url = EXCLUDED.stripe_api_base_url,
  stripe_success_url = EXCLUDED.stripe_success_url,
  stripe_cancel_url = EXCLUDED.stripe_cancel_url,
  updated_by = EXCLUDED.updated_by,
  updated_at = now()`,
		schema),
		PaymentGatewayConfigID, provider, mode, status, strings.TrimSpace(in.MerchantCode),
		apiKey, md5Key, strings.TrimSpace(in.BaseURL), routeNo, currency,
		callbackURL, returnURL,
		strings.TrimSpace(in.StripePublishableKey), stripeSecret, stripeWebhookSecret,
		stripeAPIBaseURL, stripeSuccessURL, stripeCancelURL, actor,
	)
	if err != nil {
		return PaymentGatewayConfig{}, err
	}
	return s.GetPaymentGatewayConfig(ctx)
}

func (s *Store) InsertPaymentCallbackEvent(ctx context.Context, ev PaymentCallbackEvent) (inserted bool, err error) {
	if s.pg == nil {
		return false, errors.New("postgres unavailable")
	}
	if strings.TrimSpace(ev.TransactionID) == "" {
		return false, errors.New("transaction_id is required")
	}
	if ev.ID == "" {
		ev.ID = "pce_" + newStoreID()
	}
	if ev.Provider == "" {
		ev.Provider = "chillpay"
	}
	if ev.ProviderEventID == "" {
		ev.ProviderEventID = ev.TransactionID
	}
	if ev.ProcessingStatus == "" {
		ev.ProcessingStatus = "received"
	}
	if ev.ReceivedAt.IsZero() {
		ev.ReceivedAt = time.Now().UTC()
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_callback_events (
  id, provider, transaction_id, order_no, provider_event_id, event_type, signature_verified,
  processing_status, processed_at, error_code, payment_status, amount, customer_id, payload_hash,
  received_at, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,COALESCE($15, now()),$16,$16)
ON CONFLICT (provider, transaction_id) DO NOTHING`, schema),
		ev.ID, ev.Provider, ev.TransactionID, ev.OrderNo, ev.ProviderEventID, ev.EventType, ev.SignatureVerified,
		ev.ProcessingStatus, ev.ProcessedAt, ev.ErrorCode, ev.PaymentStatus, ev.Amount, ev.CustomerID,
		ev.PayloadHash, ev.ReceivedAt, actor,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) LastPaymentCallbackAt(ctx context.Context) (*time.Time, error) {
	if s.pg == nil {
		return nil, nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var ts time.Time
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT received_at FROM %s.payment_callback_events ORDER BY received_at DESC LIMIT 1`, schema)).Scan(&ts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ts, nil
}

func (s *Store) UpdatePaymentGatewayTestStatus(ctx context.Context, status, message string) error {
	if s.pg == nil {
		return errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_gateway_configs
SET last_test_status = $1, last_tested_at = now(), last_test_error = $2, updated_by = $3, updated_at = now()
WHERE id = $4`, schema), strings.TrimSpace(status), strings.TrimSpace(message), actor, PaymentGatewayConfigID)
	return err
}

func (s *Store) UpdatePaymentGatewayWebhookStatus(ctx context.Context, status, message string) error {
	if s.pg == nil {
		return errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_gateway_configs
SET last_webhook_status = $1, last_webhook_at = now(), updated_by = $2, updated_at = now()
WHERE id = $3`, schema), strings.TrimSpace(status), actor, PaymentGatewayConfigID)
	_ = message
	return err
}

// SeedPaymentGatewayFromEnv upserts active payment gateway config when env credentials are set.
func (s *Store) SeedPaymentGatewayFromEnv(ctx context.Context) error {
	if strings.TrimSpace(s.cfg.ChillPayMerchantCode) == "" && strings.TrimSpace(s.cfg.StripeSecretKey) == "" {
		return nil
	}
	row, err := s.GetPaymentGatewayConfig(ctx)
	if err != nil {
		return err
	}
	if row.Status == "active" && strings.TrimSpace(row.Provider) != "" {
		return nil
	}
	if strings.TrimSpace(s.cfg.StripeSecretKey) != "" && strings.TrimSpace(s.cfg.ChillPayMerchantCode) == "" {
		callbackURL := strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + "/api/callbacks/stripe"
		returnURL := strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + "/tenant/billing/return"
		cancelURL := strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + "/tenant/billing"
		_, err = s.UpsertPaymentGatewayConfig(ctx, PaymentGatewayUpsert{
			Provider:               "stripe",
			Mode:                   "test",
			Status:                 "active",
			Currency:               "THB",
			CallbackURL:            callbackURL,
			ReturnURL:              returnURL,
			StripePublishableKey:   s.cfg.StripePublishableKey,
			StripeSecretKey:        s.cfg.StripeSecretKey,
			StripeWebhookSecret:    s.cfg.StripeWebhookSecret,
			StripeAPIBaseURL:       s.cfg.StripeAPIBaseURL,
			StripeSuccessURL:       firstStoreNonEmpty(s.cfg.StripeSuccessURL, returnURL),
			StripeCancelURL:        firstStoreNonEmpty(s.cfg.StripeCancelURL, cancelURL),
			SetStripeSecret:        true,
			SetStripeWebhookSecret: true,
		})
		return err
	}
	callbackURL := strings.TrimSpace(s.cfg.ChillPayCallbackURL)
	if callbackURL == "" {
		callbackURL = strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + "/api/callbacks/chillpay"
	}
	returnURL := strings.TrimSpace(s.cfg.ChillPayReturnURL)
	if returnURL == "" {
		returnURL = strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + "/tenant/billing/return"
	}
	_, err = s.UpsertPaymentGatewayConfig(ctx, PaymentGatewayUpsert{
		Provider:     "chillpay",
		Mode:         "test",
		Status:       "active",
		MerchantCode: s.cfg.ChillPayMerchantCode,
		APIKey:       s.cfg.ChillPayAPIKey,
		MD5Key:       s.cfg.ChillPayMD5Key,
		BaseURL:      s.cfg.ChillPayBaseURL,
		RouteNo:      s.cfg.ChillPayRouteNo,
		Currency:     s.cfg.ChillPayCurrency,
		CallbackURL:  callbackURL,
		ReturnURL:    returnURL,
		SetAPIKey:    true,
		SetMD5Key:    true,
	})
	return err
}

func newPaymentOrderID() string {
	return "ord_" + newStoreID()
}

// newPaymentOrderNo builds a ChillPay-safe OrderNo.
// ChillPay Merchant Integration Manual (Table 2.2): OrderNo is max 20 chars,
// alphanumeric only (A–Z a–z 0–9) — no underscores, hyphens, or other symbols.
// Format: MJ + 2-char tenant fingerprint + 16-char id = 20 chars.
func newPaymentOrderNo(tenantID string) string {
	id := newStoreID() // 16 hex digits
	fp := "00"
	if t := strings.TrimSpace(tenantID); t != "" {
		var sum uint32
		for i := 0; i < len(t); i++ {
			sum = sum*33 + uint32(t[i])
		}
		fp = fmt.Sprintf("%02x", sum&0xff)
	}
	return "MJ" + fp + id
}

func (s *Store) CreatePaymentOrder(ctx context.Context, in CreatePaymentOrderInput) (*PaymentOrder, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	tenantID := strings.TrimSpace(in.TenantID)
	packageID := strings.TrimSpace(in.PackageID)
	if tenantID == "" || packageID == "" {
		return nil, errors.New("tenant_id and package_id are required")
	}
	id := newPaymentOrderID()
	orderNo := newPaymentOrderNo(tenantID)
	currency := strings.TrimSpace(in.Currency)
	if currency == "" {
		currency = "764"
	}
	provider := strings.TrimSpace(in.Provider)
	method := strings.TrimSpace(in.PaymentMethod)
	if method == "" {
		method = "credit_card"
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_orders (
  id, tenant_id, package_id, order_no, amount_cents, currency, status, provider, payment_method,
  created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,'pending',$7,$8,$9,$9)`, schema),
		id, tenantID, packageID, orderNo, in.AmountCents, currency, provider, method, actor,
	)
	if err != nil {
		return nil, err
	}
	return s.GetPaymentOrderByID(ctx, id)
}

func (s *Store) UpdatePaymentOrderInit(ctx context.Context, orderID, transactionID, paymentURL string) error {
	if s.pg == nil {
		return errors.New("postgres unavailable")
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_orders
SET transaction_id = $2, payment_url = $3, updated_by = $4
WHERE id = $1`, schema), orderID, transactionID, paymentURL, actor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentOrderNotFound
	}
	return nil
}

func (s *Store) UpdatePaymentOrderProviderRefs(ctx context.Context, orderID, transactionID, paymentURL, sessionID, paymentID, providerStatus string, expiresAt *time.Time) error {
	if s.pg == nil {
		return errors.New("postgres unavailable")
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_orders
SET transaction_id = CASE WHEN $2 <> '' THEN $2 ELSE transaction_id END,
    payment_url = CASE WHEN $3 <> '' THEN $3 ELSE payment_url END,
    provider_session_id = CASE WHEN $4 <> '' THEN $4 ELSE provider_session_id END,
    provider_payment_id = CASE WHEN $5 <> '' THEN $5 ELSE provider_payment_id END,
    provider_status = CASE WHEN $6 <> '' THEN $6 ELSE provider_status END,
    checkout_expires_at = COALESCE($7, checkout_expires_at),
    last_provider_sync_at = now(),
    updated_by = $8,
    updated_at = now()
WHERE id = $1`, schema), orderID, strings.TrimSpace(transactionID), strings.TrimSpace(paymentURL),
		strings.TrimSpace(sessionID), strings.TrimSpace(paymentID), strings.TrimSpace(providerStatus), expiresAt, actor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentOrderNotFound
	}
	return nil
}

func scanPaymentOrder(row pgx.Row) (PaymentOrder, error) {
	var o PaymentOrder
	err := row.Scan(
		&o.ID, &o.TenantID, &o.PackageID, &o.OrderNo, &o.AmountCents, &o.Currency, &o.Status,
		&o.Provider, &o.PaymentMethod, &o.TransactionID, &o.PaymentURL,
		&o.ProviderSessionID, &o.ProviderPaymentID, &o.ProviderStatus, &o.CheckoutExpiresAt, &o.LastProviderSyncAt,
		&o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if o.PaymentMethod == "" {
		o.PaymentMethod = "credit_card"
	}
	return o, err
}

const paymentOrderSelectCols = `id, tenant_id, package_id, order_no, amount_cents, currency, status, provider,
       COALESCE(payment_method, 'credit_card'), transaction_id, payment_url,
       COALESCE(provider_session_id, ''), COALESCE(provider_payment_id, ''), COALESCE(provider_status, ''),
       checkout_expires_at, last_provider_sync_at, paid_at, created_at, updated_at`

func (s *Store) GetPaymentOrderByID(ctx context.Context, id string) (*PaymentOrder, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	o, err := scanPaymentOrder(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.payment_orders WHERE id = $1`, paymentOrderSelectCols, schema), id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) GetPaymentOrderByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	o, err := scanPaymentOrder(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.payment_orders WHERE order_no = $1`, paymentOrderSelectCols, schema), orderNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) GetPaymentOrderByProviderSession(ctx context.Context, provider, sessionID string) (*PaymentOrder, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	provider = strings.TrimSpace(provider)
	sessionID = strings.TrimSpace(sessionID)
	if provider == "" || sessionID == "" {
		return nil, ErrPaymentOrderNotFound
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	o, err := scanPaymentOrder(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.payment_orders WHERE provider = $1 AND provider_session_id = $2`, paymentOrderSelectCols, schema), provider, sessionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) ListPaymentOrdersForProviderReconcile(ctx context.Context, provider string, since *time.Time, limit int) ([]PaymentOrder, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.payment_orders
WHERE provider = $1
  AND ($2::timestamptz IS NULL OR created_at >= $2)
  AND provider_session_id <> ''
ORDER BY created_at DESC
LIMIT $3`, paymentOrderSelectCols, schema), provider, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PaymentOrder
	for rows.Next() {
		o, err := scanPaymentOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) FulfillPaymentOrder(ctx context.Context, orderNo, transactionID, paymentStatus string) (PaymentFulfillResult, error) {
	if s.pg == nil {
		return PaymentFulfillResult{}, errors.New("postgres unavailable")
	}
	order, err := s.GetPaymentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return PaymentFulfillResult{}, err
	}

	status := strings.TrimSpace(paymentStatus)
	switch status {
	case "0":
		if order.Status == PaymentOrderStatusPaid {
			return PaymentFulfillResult{Order: *order, EntitlementChanged: false}, nil
		}
		if order.Status != PaymentOrderStatusPending {
			return PaymentFulfillResult{Order: *order, EntitlementChanged: false}, nil
		}
		changed, err := s.markOrderPaidAndAssignEntitlement(ctx, order, transactionID)
		if err != nil {
			return PaymentFulfillResult{}, err
		}
		updated, err := s.GetPaymentOrderByID(ctx, order.ID)
		if err != nil {
			return PaymentFulfillResult{}, err
		}
		return PaymentFulfillResult{Order: *updated, EntitlementChanged: changed}, nil
	case "2":
		if order.Status == PaymentOrderStatusPaid {
			return PaymentFulfillResult{Order: *order, EntitlementChanged: false}, nil
		}
		if order.Status != PaymentOrderStatusPending {
			return PaymentFulfillResult{Order: *order, EntitlementChanged: false}, nil
		}
		if err := s.updatePaymentOrderFailed(ctx, order.ID, transactionID); err != nil {
			return PaymentFulfillResult{}, err
		}
		updated, err := s.GetPaymentOrderByID(ctx, order.ID)
		if err != nil {
			return PaymentFulfillResult{}, err
		}
		return PaymentFulfillResult{Order: *updated, EntitlementChanged: false}, nil
	default:
		return PaymentFulfillResult{Order: *order, EntitlementChanged: false}, nil
	}
}

func (s *Store) markOrderPaidAndAssignEntitlement(ctx context.Context, order *PaymentOrder, transactionID string) (bool, error) {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	txnID := strings.TrimSpace(transactionID)

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_orders
SET status = 'paid', paid_at = now(),
    transaction_id = CASE WHEN $3 <> '' THEN $3 ELSE transaction_id END,
    provider_payment_id = CASE WHEN $3 <> '' THEN $3 ELSE provider_payment_id END,
    provider_status = 'paid',
    last_provider_sync_at = now(),
    updated_by = $2,
    updated_at = now()
WHERE id = $1 AND status = 'pending'`, schema), order.ID, actor, txnID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return false, nil
	}

	// Renewal orders already have an immutable subscription and entitlement.
	// Do not run the initial-purchase replacement path: doing so would apply the
	// latest package rules to a historical subscription and churn entitlements.
	var renewalSubscriptionID string
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT subscription_id
FROM %s.billing_cycles
WHERE order_id = $1
FOR UPDATE`, schema), order.ID).Scan(&renewalSubscriptionID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		if err := s.issueDocumentsAndSettlePaidOrder(ctx, order.ID); err != nil {
			// Payment is durable. The scheduler replay path recovers document or
			// settlement failures without charging or granting again.
			return false, nil
		}
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	// Initial commercial checkout reads the package-version rules captured by
	// the pending subscription. Legacy orders without a subscription fall back
	// to the current package row for backward compatibility.
	var packageID, rulesSchemaID string
	var rulesJSON []byte
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT pv.package_id, pl.rules_schema_id, pv.rules_snapshot
FROM %s.tenant_subscriptions s
JOIN %s.package_versions pv ON pv.id = s.package_version_id
JOIN %s.package_limits pl ON pl.package_id = pv.package_id
WHERE s.initial_order_id = $1 AND s.status = 'pending_payment'
FOR UPDATE OF s`, schema, schema, schema), order.ID).Scan(&packageID, &rulesSchemaID, &rulesJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT p.id, pl.rules_schema_id, pl.rules
FROM %s.packages p
JOIN %s.package_limits pl ON pl.package_id = p.id
WHERE p.id = $1`, schema, schema), order.PackageID).Scan(&packageID, &rulesSchemaID, &rulesJSON)
	}
	if err != nil {
		return false, err
	}

	// New id each fulfillment so re-buy after revoke does not collide on PK.
	entID := "ent_" + order.TenantID + "_" + packageID + "_" + newStoreID()

	// Revoke current active plan, then assign purchased package rules (quota).
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_entitlements SET status = 'revoked', updated_by = $2
WHERE tenant_id = $1 AND status = 'active'`, schema), order.TenantID, actor)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_entitlements (id, tenant_id, package_id, rules_schema_id, rules_snapshot, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5::jsonb, 'active', $6, $6)`, schema),
		entID, order.TenantID, packageID, rulesSchemaID, string(rulesJSON), actor)
	if err != nil {
		return false, err
	}
	if err := s.activateSubscriptionForPaidOrderTx(ctx, tx, order, entID, actor); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	// Failures after entitlement do not roll back payment. Document GET and the
	// scheduler recovery path can replay both operations idempotently.
	_ = s.issueDocumentsAndSettlePaidOrder(ctx, order.ID)
	return true, nil
}

func (s *Store) issueDocumentsAndSettlePaidOrder(ctx context.Context, orderID string) error {
	if err := s.IssuePaymentDocuments(ctx, orderID); err != nil {
		return err
	}
	// Initial checkout orders have no cycle and are a harmless no-op.
	return s.SettleBillingCycleForPaidOrder(ctx, orderID)
}

func (s *Store) updatePaymentOrderFailed(ctx context.Context, orderID, transactionID string) error {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	txnID := strings.TrimSpace(transactionID)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_orders
SET status = 'failed',
    transaction_id = CASE WHEN $3 <> '' THEN $3 ELSE transaction_id END,
    provider_payment_id = CASE WHEN $3 <> '' THEN $3 ELSE provider_payment_id END,
    provider_status = 'failed',
    last_provider_sync_at = now(),
    updated_by = $2,
    updated_at = now()
WHERE id = $1 AND status = 'pending'`, schema), orderID, actor, txnID)
	return err
}

func firstStoreNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// splitVATInclusive treats amount as VAT-inclusive and returns net + vat at rateBps (700 = 7%).
func splitVATInclusive(amountCents, rateBps int) (netCents, vatCents int) {
	if amountCents <= 0 || rateBps <= 0 {
		return amountCents, 0
	}
	// net = amount * 10000 / (10000 + rateBps)
	netCents = (amountCents * 10000) / (10000 + rateBps)
	vatCents = amountCents - netCents
	return netCents, vatCents
}

// IssuePaymentDocuments creates receipt + tax invoice for a paid order (idempotent for active docs).
func (s *Store) IssuePaymentDocuments(ctx context.Context, orderID string) error {
	if s.pg == nil {
		return errors.New("postgres unavailable")
	}
	order, err := s.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != PaymentOrderStatusPaid {
		return fmt.Errorf("order is not paid")
	}

	existing, err := s.ListPaymentDocumentsByOrder(ctx, orderID)
	if err != nil {
		return err
	}

	pkg, err := s.GetPackage(ctx, order.PackageID)
	if err != nil {
		return err
	}
	packageName := pkg.Name
	vatRateBps := 700 // legacy order fallback
	net, vat := splitVATInclusive(order.AmountCents, vatRateBps)
	if calculation, calculationErr := s.GetOrderCommercialCalculation(ctx, order.ID); calculationErr == nil {
		if strings.TrimSpace(calculation.PackageName) != "" {
			packageName = calculation.PackageName
		}
		vatRateBps = calculation.TaxRateBps
		net, vat = splitVATInclusive(order.AmountCents, vatRateBps)
		if calculation.TaxableAmountCents+calculation.TaxCents == order.AmountCents {
			net = calculation.TaxableAmountCents
			vat = calculation.TaxCents
		}
	}
	buyerName, buyerAddr, buyerTaxID := s.resolveBuyerFields(ctx, order.TenantID)
	seller, _ := s.GetSellerBranding(ctx)

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	issuedAt := time.Now().UTC()
	if order.PaidAt != nil {
		issuedAt = order.PaidAt.UTC()
	}

	docs := []struct {
		docType string
		prefix  string
	}{
		{PaymentDocTypeReceipt, "RCP"},
		{PaymentDocTypeTaxInvoice, "TAX"},
	}
	for _, d := range docs {
		if hasActiveDocType(existing, d.docType) {
			continue
		}
		id := "pdoc_" + newStoreID()
		docNo := fmt.Sprintf("%s-%s", d.prefix, order.OrderNo)
		// If re-issuing after void, suffix unique doc number.
		for _, e := range existing {
			if e.DocType == d.docType {
				docNo = fmt.Sprintf("%s-%s-%s", d.prefix, order.OrderNo, newStoreID()[:8])
				break
			}
		}
		_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_documents (
  id, order_id, tenant_id, doc_type, doc_number, status,
  buyer_name, buyer_address, buyer_tax_id,
  seller_name, seller_address, seller_tax_id,
  package_name, amount_cents, currency, vat_rate_bps, net_cents, vat_cents,
  payment_method, issued_at, created_by, updated_by
) VALUES (
  $1,$2,$3,$4,$5,'issued',
  $6,$7,$8,
  $9,$10,$11,
  $12,$13,$14,$15,$16,$17,
  $18,$19,$20,$20
)`, schema),
			id, order.ID, order.TenantID, d.docType, docNo,
			buyerName, buyerAddr, buyerTaxID,
			seller.Name, seller.Address, seller.TaxID,
			packageName, order.AmountCents, order.Currency, vatRateBps, net, vat,
			order.PaymentMethod, issuedAt, actor,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) resolveBuyerFields(ctx context.Context, tenantID string) (name, addr, taxID string) {
	name = tenantID
	if tax, err := s.GetTenantTaxProfile(ctx, tenantID); err == nil {
		if strings.TrimSpace(tax.CompanyName) != "" {
			name = tax.CompanyName
		}
		if strings.TrimSpace(tax.Address) != "" {
			addr = tax.Address
		}
		taxID = strings.TrimSpace(tax.TaxID)
	}
	if reg, regErr := s.GetTenantRegistration(ctx, tenantID); regErr == nil && strings.TrimSpace(reg.CompanyName) != "" {
		if name == tenantID {
			name = reg.CompanyName
		}
	}
	if kyc, kycErr := s.GetTenantKYCProfile(ctx, tenantID); kycErr == nil {
		if strings.TrimSpace(kyc.ContactAddress) != "" && addr == "" {
			addr = kyc.ContactAddress
		}
		if strings.TrimSpace(kyc.ContactName) != "" && name == tenantID {
			name = kyc.ContactName
		}
	}
	return name, addr, taxID
}

func hasActiveDocType(docs []PaymentDocument, docType string) bool {
	for _, d := range docs {
		if d.DocType == docType && (d.Status == "" || d.Status == PaymentDocStatusIssued) {
			return true
		}
	}
	return false
}

func scanPaymentDocument(row pgx.Row) (PaymentDocument, error) {
	var d PaymentDocument
	var voidedAt *time.Time
	err := row.Scan(
		&d.ID, &d.OrderID, &d.TenantID, &d.DocType, &d.DocNumber, &d.Status,
		&d.BuyerName, &d.BuyerAddress, &d.BuyerTaxID,
		&d.SellerName, &d.SellerAddress, &d.SellerTaxID,
		&d.PackageName, &d.AmountCents, &d.Currency, &d.VATRateBps, &d.NetCents, &d.VATCents,
		&d.PaymentMethod, &d.ReissuedFromID, &d.VoidReason, &voidedAt, &d.IssuedAt, &d.CreatedAt,
	)
	if d.Status == "" {
		d.Status = PaymentDocStatusIssued
	}
	d.VoidedAt = voidedAt
	return d, err
}

const paymentDocumentSelectCols = `id, order_id, tenant_id, doc_type, doc_number, COALESCE(status, 'issued'),
  buyer_name, buyer_address, buyer_tax_id,
  seller_name, seller_address, seller_tax_id,
  package_name, amount_cents, currency, vat_rate_bps, net_cents, vat_cents,
  payment_method, COALESCE(reissued_from_id, ''), COALESCE(void_reason, ''), voided_at, issued_at, created_at`

func (s *Store) ListPaymentDocumentsByOrder(ctx context.Context, orderID string) ([]PaymentDocument, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT %s FROM %s.payment_documents WHERE order_id = $1 ORDER BY issued_at DESC, doc_type`, paymentDocumentSelectCols, schema), orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PaymentDocument
	for rows.Next() {
		d, err := scanPaymentDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) GetPaymentDocument(ctx context.Context, orderID, docType string) (*PaymentDocument, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	d, err := scanPaymentDocument(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s FROM %s.payment_documents
WHERE order_id = $1 AND doc_type = $2 AND COALESCE(status, 'issued') = 'issued'
ORDER BY issued_at DESC LIMIT 1`, paymentDocumentSelectCols, schema),
		orderID, docType))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) GetPaymentDocumentByID(ctx context.Context, docID string) (*PaymentDocument, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	d, err := scanPaymentDocument(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s FROM %s.payment_documents WHERE id = $1`, paymentDocumentSelectCols, schema), docID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}
