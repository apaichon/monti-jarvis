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

var ErrPaymentGatewayNotConfigured = errors.New("payment gateway not configured")

type PaymentGatewayConfig struct {
	ID           string
	Provider     string
	Mode         string
	Status       string
	MerchantCode string
	APIKey       string
	MD5Key       string
	BaseURL      string
	RouteNo      int
	Currency     string
	CallbackURL  string
	ReturnURL    string
	UpdatedAt    time.Time
}

type PaymentGatewayUpsert struct {
	Provider     string
	Mode         string
	Status       string
	MerchantCode string
	APIKey       string
	MD5Key       string
	BaseURL      string
	RouteNo      int
	Currency     string
	CallbackURL  string
	ReturnURL    string
	SetAPIKey    bool
	SetMD5Key    bool
}

type PaymentCallbackEvent struct {
	ID             string
	Provider       string
	TransactionID  string
	OrderNo        string
	PaymentStatus  string
	Amount         string
	CustomerID     string
	PayloadHash    string
	ReceivedAt     time.Time
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
  return_url text NOT NULL DEFAULT '',%s
);
CREATE TABLE IF NOT EXISTS %s.payment_callback_events (
  id text PRIMARY KEY,
  provider text NOT NULL,
  transaction_id text NOT NULL,
  order_no text NOT NULL DEFAULT '',
  payment_status text NOT NULL DEFAULT '',
  amount text NOT NULL DEFAULT '',
  customer_id text NOT NULL DEFAULT '',
  payload_hash text NOT NULL DEFAULT '',
  received_at timestamptz NOT NULL DEFAULT now(),%s
  UNIQUE (provider, transaction_id)
);
CREATE INDEX IF NOT EXISTS payment_callback_events_received_idx
  ON %s.payment_callback_events (received_at DESC);`,
		schema, auditColumnsDDL, schema, auditColumnsDDL, schema))
	return err
}

func (s *Store) GetPaymentGatewayConfig(ctx context.Context) (PaymentGatewayConfig, error) {
	if s.pg == nil {
		return PaymentGatewayConfig{}, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var row PaymentGatewayConfig
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, provider, mode, status, merchant_code, api_key, md5_key, base_url, route_no, currency,
       callback_url, return_url, updated_at
FROM %s.payment_gateway_configs
WHERE id = $1`, schema), PaymentGatewayConfigID).Scan(
		&row.ID, &row.Provider, &row.Mode, &row.Status, &row.MerchantCode, &row.APIKey, &row.MD5Key,
		&row.BaseURL, &row.RouteNo, &row.Currency, &row.CallbackURL, &row.ReturnURL, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PaymentGatewayConfig{ID: PaymentGatewayConfigID, Mode: "test", Status: "inactive", RouteNo: 1, Currency: "764"}, nil
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

	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_gateway_configs (
  id, provider, mode, status, merchant_code, api_key, md5_key, base_url, route_no, currency,
  callback_url, return_url, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$13)
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
  updated_by = EXCLUDED.updated_by`,
		schema),
		PaymentGatewayConfigID, provider, mode, status, strings.TrimSpace(in.MerchantCode),
		apiKey, md5Key, strings.TrimSpace(in.BaseURL), in.RouteNo, strings.TrimSpace(in.Currency),
		strings.TrimSpace(in.CallbackURL), strings.TrimSpace(in.ReturnURL), actor,
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
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_callback_events (
  id, provider, transaction_id, order_no, payment_status, amount, customer_id, payload_hash,
  received_at, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE($9, now()),$10,$10)
ON CONFLICT (provider, transaction_id) DO NOTHING`, schema),
		ev.ID, ev.Provider, ev.TransactionID, ev.OrderNo, ev.PaymentStatus, ev.Amount, ev.CustomerID,
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

// SeedPaymentGatewayFromEnv upserts active ChillPay config when env credentials are set.
func (s *Store) SeedPaymentGatewayFromEnv(ctx context.Context) error {
	if strings.TrimSpace(s.cfg.ChillPayMerchantCode) == "" {
		return nil
	}
	row, err := s.GetPaymentGatewayConfig(ctx)
	if err != nil {
		return err
	}
	if row.Status == "active" && strings.TrimSpace(row.Provider) != "" {
		return nil
	}
	callbackURL := strings.TrimSpace(s.cfg.ChillPayCallbackURL)
	if callbackURL == "" {
		callbackURL = strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/") + "/api/callbacks/chillpay"
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
		ReturnURL:    s.cfg.ChillPayReturnURL,
		SetAPIKey:    true,
		SetMD5Key:    true,
	})
	return err
}