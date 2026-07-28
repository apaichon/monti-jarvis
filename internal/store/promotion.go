package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrPromotionGrantNotFound  = errors.New("promotion grant not found")
	ErrPackageNotActive        = errors.New("package is not active")
	ErrPromotionReasonRequired = errors.New("reason is required")
	ErrIdempotencyConflict     = errors.New("idempotency key conflict")
)

const (
	PromotionProvider       = "promotion"
	PromotionPaymentMethod  = "promotion"
	PromotionGrantStatusOK  = "issued"
	PromotionGrantStatusFail = "failed"
)

// PromotionGrant is an admin promotional package grant audit row.
type PromotionGrant struct {
	ID            string
	TenantID      string
	PackageID     string
	OrderID       string
	EntitlementID string
	TaxInvoiceID  string
	Reason        string
	IdempotencyKey string
	ValidUntil    *time.Time
	AmountCents   int
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
}

// PromotionGrantInput creates a promotional grant.
type PromotionGrantInput struct {
	TenantID       string
	PackageID      string
	Reason         string
	ValidUntil     *time.Time
	AmountCents    *int // nil → 0
	IdempotencyKey string
}

// PromotionGrantResult is the atomic grant outcome.
type PromotionGrantResult struct {
	Grant       PromotionGrant
	Entitlement *TenantEntitlement
	TaxInvoice  *PaymentDocument
	Replayed    bool
}

func (s *Store) ensurePromotionSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.promotion_grants (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  package_id text NOT NULL REFERENCES %s.packages(id),
  order_id text NOT NULL UNIQUE REFERENCES %s.payment_orders(id) ON DELETE CASCADE,
  entitlement_id text NOT NULL DEFAULT '',
  tax_invoice_id text NOT NULL DEFAULT '',
  reason text NOT NULL,
  idempotency_key text NOT NULL DEFAULT '',
  valid_until timestamptz,
  amount_cents int NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'issued' CHECK (status IN ('issued', 'failed')),%s
);
CREATE INDEX IF NOT EXISTS promotion_grants_tenant_created_idx
  ON %s.promotion_grants (tenant_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS promotion_grants_tenant_idempotency_uidx
  ON %s.promotion_grants (tenant_id, idempotency_key)
  WHERE idempotency_key <> '';`,
		schema, schema, schema, schema, auditColumnsDDL, schema, schema))
	return err
}

// newPromotionOrderNo builds an alphanumeric order number (≤20) for promo orders.
// Format: PR + 2-char tenant fingerprint + 16 hex = 20 chars.
func newPromotionOrderNo(tenantID string) string {
	id := newStoreID()
	fp := "00"
	if t := strings.TrimSpace(tenantID); t != "" {
		var sum uint32
		for i := 0; i < len(t); i++ {
			sum = sum*33 + uint32(t[i])
		}
		fp = fmt.Sprintf("%02x", sum&0xff)
	}
	return "PR" + fp + id
}

// GrantPromotion sets the tenant active plan and issues a tax invoice atomically.
func (s *Store) GrantPromotion(ctx context.Context, in PromotionGrantInput) (*PromotionGrantResult, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	tenantID := strings.TrimSpace(in.TenantID)
	packageID := strings.TrimSpace(in.PackageID)
	reason := strings.TrimSpace(in.Reason)
	idemKey := strings.TrimSpace(in.IdempotencyKey)
	if tenantID == "" {
		return nil, ErrTenantNotFound
	}
	if packageID == "" {
		return nil, ErrPackageNotFound
	}
	if reason == "" {
		return nil, ErrPromotionReasonRequired
	}
	amount := 0
	if in.AmountCents != nil {
		if *in.AmountCents < 0 {
			return nil, fmt.Errorf("amount_cents must be >= 0")
		}
		amount = *in.AmountCents
	}

	exists, err := s.TenantExists(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTenantNotFound
	}

	pkg, err := s.GetPackage(ctx, packageID)
	if err != nil {
		return nil, err
	}
	if pkg.Status != "active" {
		return nil, ErrPackageNotActive
	}

	if idemKey != "" {
		if prior, err := s.GetPromotionGrantByIdempotency(ctx, tenantID, idemKey); err == nil {
			if prior.PackageID != packageID || prior.Reason != reason || prior.AmountCents != amount {
				return nil, ErrIdempotencyConflict
			}
			return s.loadPromotionGrantResult(ctx, prior, true)
		} else if !errors.Is(err, ErrPromotionGrantNotFound) {
			return nil, err
		}
	}

	snapJSON, err := json.Marshal(pkg.Rules)
	if err != nil {
		return nil, err
	}

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	orderID := newPaymentOrderID()
	orderNo := newPromotionOrderNo(tenantID)
	entID := "ent_" + tenantID + "_" + packageID + "_" + newStoreID()
	grantID := "pgr_" + newStoreID()
	taxDocID := "pdoc_" + newStoreID()
	docNo := fmt.Sprintf("TAX-%s", orderNo)
	currency := strings.TrimSpace(pkg.Currency)
	if currency == "" {
		currency = "764"
	}
	txnID := "promo_" + grantID

	buyerName, buyerAddr, buyerTaxID := s.resolveBuyerFields(ctx, tenantID)
	seller, _ := s.GetSellerBranding(ctx)
	const vatRateBps = 700
	net, vat := splitVATInclusive(amount, vatRateBps)
	now := time.Now().UTC()

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Re-check idempotency inside TX to avoid races.
	if idemKey != "" {
		var existingID string
		err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id FROM %s.promotion_grants
WHERE tenant_id = $1 AND idempotency_key = $2`, schema), tenantID, idemKey).Scan(&existingID)
		if err == nil {
			if err := tx.Rollback(ctx); err != nil {
				return nil, err
			}
			prior, gerr := s.GetPromotionGrant(ctx, existingID)
			if gerr != nil {
				return nil, gerr
			}
			return s.loadPromotionGrantResult(ctx, prior, true)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_orders (
  id, tenant_id, package_id, order_no, amount_cents, currency, status, provider, payment_method,
  transaction_id, payment_url, paid_at, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,'paid',$7,$8,$9,'',$10,$11,$11)`, schema),
		orderID, tenantID, packageID, orderNo, amount, currency,
		PromotionProvider, PromotionPaymentMethod, txnID, now, actor)
	if err != nil {
		return nil, fmt.Errorf("insert promotion order: %w", err)
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_entitlements SET status = 'revoked', updated_by = $2
WHERE tenant_id = $1 AND status = 'active'`, schema), tenantID, actor)
	if err != nil {
		return nil, fmt.Errorf("revoke prior entitlement: %w", err)
	}

	if in.ValidUntil != nil {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_entitlements (
  id, tenant_id, package_id, rules_schema_id, rules_snapshot, status, valid_from, valid_until, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5::jsonb,'active',now(),$6,$7,$7)`, schema),
			entID, tenantID, packageID, pkg.RulesSchemaID, string(snapJSON), in.ValidUntil.UTC(), actor)
	} else {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_entitlements (
  id, tenant_id, package_id, rules_schema_id, rules_snapshot, status, valid_from, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5::jsonb,'active',now(),$6,$6)`, schema),
			entID, tenantID, packageID, pkg.RulesSchemaID, string(snapJSON), actor)
	}
	if err != nil {
		return nil, fmt.Errorf("insert entitlement: %w", err)
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
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
		taxDocID, orderID, tenantID, PaymentDocTypeTaxInvoice, docNo,
		buyerName, buyerAddr, buyerTaxID,
		seller.Name, seller.Address, seller.TaxID,
		pkg.Name, amount, currency, vatRateBps, net, vat,
		PromotionPaymentMethod, now, actor)
	if err != nil {
		return nil, fmt.Errorf("insert tax invoice: %w", err)
	}

	// Optional receipt when amount > 0 (parity with paid path).
	if amount > 0 {
		rcpID := "pdoc_" + newStoreID()
		rcpNo := fmt.Sprintf("RCP-%s", orderNo)
		_, err = tx.Exec(ctx, fmt.Sprintf(`
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
			rcpID, orderID, tenantID, PaymentDocTypeReceipt, rcpNo,
			buyerName, buyerAddr, buyerTaxID,
			seller.Name, seller.Address, seller.TaxID,
			pkg.Name, amount, currency, vatRateBps, net, vat,
			PromotionPaymentMethod, now, actor)
		if err != nil {
			return nil, fmt.Errorf("insert receipt: %w", err)
		}
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.promotion_grants (
  id, tenant_id, package_id, order_id, entitlement_id, tax_invoice_id,
  reason, idempotency_key, valid_until, amount_cents, status, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'issued',$11,$11)`, schema),
		grantID, tenantID, packageID, orderID, entID, taxDocID,
		reason, idemKey, in.ValidUntil, amount, actor)
	if err != nil {
		return nil, fmt.Errorf("insert promotion grant: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	grant, err := s.GetPromotionGrant(ctx, grantID)
	if err != nil {
		return nil, err
	}
	return s.loadPromotionGrantResult(ctx, grant, false)
}

func (s *Store) loadPromotionGrantResult(ctx context.Context, grant *PromotionGrant, replayed bool) (*PromotionGrantResult, error) {
	ent, err := s.GetActiveEntitlement(ctx, grant.TenantID)
	if err != nil {
		return nil, err
	}
	tax, err := s.GetPaymentDocumentByID(ctx, grant.TaxInvoiceID)
	if err != nil {
		return nil, err
	}
	return &PromotionGrantResult{
		Grant:       *grant,
		Entitlement: ent,
		TaxInvoice:  tax,
		Replayed:    replayed,
	}, nil
}

func scanPromotionGrant(row pgx.Row) (*PromotionGrant, error) {
	var g PromotionGrant
	var validUntil *time.Time
	err := row.Scan(
		&g.ID, &g.TenantID, &g.PackageID, &g.OrderID, &g.EntitlementID, &g.TaxInvoiceID,
		&g.Reason, &g.IdempotencyKey, &validUntil, &g.AmountCents, &g.Status,
		&g.CreatedAt, &g.UpdatedAt, &g.CreatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPromotionGrantNotFound
	}
	if err != nil {
		return nil, err
	}
	g.ValidUntil = validUntil
	return &g, nil
}

const promotionGrantSelectCols = `id, tenant_id, package_id, order_id, entitlement_id, tax_invoice_id,
  reason, COALESCE(idempotency_key,''), valid_until, amount_cents, status,
  created_at, updated_at, COALESCE(created_by,'')`

func (s *Store) GetPromotionGrant(ctx context.Context, grantID string) (*PromotionGrant, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanPromotionGrant(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s FROM %s.promotion_grants WHERE id = $1`, promotionGrantSelectCols, schema), strings.TrimSpace(grantID)))
}

func (s *Store) GetPromotionGrantByIdempotency(ctx context.Context, tenantID, key string) (*PromotionGrant, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanPromotionGrant(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s FROM %s.promotion_grants
WHERE tenant_id = $1 AND idempotency_key = $2`, promotionGrantSelectCols, schema),
		strings.TrimSpace(tenantID), strings.TrimSpace(key)))
}

func (s *Store) ListPromotionGrants(ctx context.Context, tenantID string, limit int) ([]PromotionGrant, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT %s FROM %s.promotion_grants
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2`, promotionGrantSelectCols, schema), strings.TrimSpace(tenantID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PromotionGrant
	for rows.Next() {
		g, err := scanPromotionGrant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, rows.Err()
}
